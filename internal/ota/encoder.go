package ota

import (
	"bufio"
	"encoding/binary"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/arduino/arduino-cloud-cli/internal/lzss"
	"github.com/juju/errors"
)

// A writer is a buffered, flushable writer.
type writer interface {
	io.Writer
	Flush() error
}

// encoder encodes a binary into an .ota file.
type encoder struct {
	// w is the writer that compressed bytes are written to.
	w writer

	// vendorID is the ID of the board vendor
	vendorID string

	// is the ID of the board vendor is the ID of the board model
	productID string
}

// NewWriter creates a new `WriteCloser` for the the given VID/PID.
func NewWriter(w io.Writer, vendorID, productID string) io.WriteCloser {
	bw, ok := w.(writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &encoder{
		w:         bw,
		vendorID:  vendorID,
		productID: productID,
	}
}

// Write writes a compressed representation of p to e's underlying writer.
func (e *encoder) Write(binaryData []byte) (int, error) {
	//log.Println("original binaryData is", len(binaryData), "bytes length")

	// Magic number (VID/PID)
	magicNumber := make([]byte, 4)
	vid, err := strconv.ParseUint(e.vendorID, 16, 16)
	if err != nil {
		return 0, errors.Annotate(err, "OTA encoder: failed to parse vendorID")
	}
	pid, err := strconv.ParseUint(e.productID, 16, 16)
	if err != nil {
		return 0, errors.Annotate(err, "OTA encoder: failed to parse productID")
	}

	binary.LittleEndian.PutUint16(magicNumber[0:2], uint16(pid))
	binary.LittleEndian.PutUint16(magicNumber[2:4], uint16(vid))

	// Version field (byte array of size 8)
	version := Version{
		Compression: true,
	}

	// Compress the compiled binary
	compressed, err := e.compress(&binaryData)
	if err != nil {
		return 0, err
	}

	// Prepend magic number and version field to payload
	var binDataComplete []byte
	binDataComplete = append(binDataComplete, magicNumber...)
	binDataComplete = append(binDataComplete, version.AsBytes()...)
	binDataComplete = append(binDataComplete, compressed...)
	//log.Println("binDataComplete is", len(binDataComplete), "bytes length")

	headerSize, err := e.writeHeader(&binDataComplete)
	if err != nil {
		return headerSize, err
	}

	payloadSize, err := e.writePayload(&binDataComplete)
	if err != nil {
		return payloadSize, err
	}

	return headerSize + payloadSize, nil
}

// Close closes the encoder, flushing any pending output. It does not close or
// flush e's underlying writer.
func (e *encoder) Close() error {
	return e.w.Flush()
}

func (e *encoder) writeHeader(binDataComplete *[]byte) (int, error) {

	//
	// Write the length of the content
	//
	lengthAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthAsBytes, uint32(len(*binDataComplete)))

	n, err := e.w.Write(lengthAsBytes)
	if err != nil {
		return n, err
	}

	//
	// Calculate the checksum for binDataComplete
	//
	crc := crc32.ChecksumIEEE(*binDataComplete)

	// encode the checksum uint32 value as 4 bytes
	crcAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcAsBytes, crc)

	n, err = e.w.Write(crcAsBytes)
	if err != nil {
		return n, err
	}

	return len(lengthAsBytes) + len(crcAsBytes), nil
}

func (e *encoder) writePayload(data *[]byte) (int, error) {

	// write the payload
	payloadSize, err := e.w.Write(*data)
	if err != nil {
		return payloadSize, err
	}

	return payloadSize, nil
}

func (e *encoder) compress(data *[]byte) ([]byte, error) {

	// create a tmp file for input
	inputFile, err := ioutil.TempFile("", "ota-lzss-input")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer os.Remove(inputFile.Name())

	// create a tmp file for output
	outputFile, err := ioutil.TempFile("", "ota-lzss-output")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer os.Remove(outputFile.Name())

	// write data in the input file
	ioutil.WriteFile(inputFile.Name(), *data, 644)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Compress the binary data using LZSS
	lzss.Encode(inputFile.Name(), outputFile.Name())

	// reads compressed data from output file and write it into
	// the writer
	compressed, err := ioutil.ReadFile(outputFile.Name())
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return compressed, nil
}
