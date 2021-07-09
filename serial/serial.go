package serial

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/howeyc/crc16"
	"go.bug.st/serial"
)

type Serial struct {
	port serial.Port
}

func NewSerial() *Serial {
	s := &Serial{}
	return s
}

func (s *Serial) Connect(address string) error {
	mode := &serial.Mode{
		BaudRate: 57600,
	}
	port, err := serial.Open(address, mode)
	if err != nil {
		err = fmt.Errorf("%s: %w", "connecting to serial port", err)
		return err
	}
	s.port = port

	s.port.SetReadTimeout(time.Millisecond * 2000)
	return nil
}

func (s *Serial) Send(cmd Command, payload []byte) error {
	payload = append([]byte{byte(cmd)}, payload...)
	msg := encode(Cmd, payload)

	_, err := s.port.Write(msg)
	if err != nil {
		err = fmt.Errorf("%s: %w", "sending message through serial", err)
		return err
	}

	return nil
}

func (s *Serial) SendReceive(cmd Command, payload []byte) ([]byte, error) {
	err := s.Send(cmd, payload)
	if err != nil {
		return nil, err
	}
	return s.Receive()
}

func (s *Serial) Receive() ([]byte, error) {
	buff := make([]byte, 1000)
	var resp []byte

	received := 0
	packetLen := 0
	for received < packetLen+5+2 {
		n, err := s.port.Read(buff)
		if err != nil {
			err = fmt.Errorf("%s: %w", "receiving from serial", err)
			return nil, err
		}
		if n == 0 {
			break
		}
		received += n
		resp = append(resp, buff[:n]...)

		if packetLen == 0 && received >= 5 {
			packetLen = int(binary.BigEndian.Uint16(resp[3:5]))
		}
	}

	if received == 0 {
		err := errors.New("receiving from serial: timeout, nothing received")
		return nil, err
	}

	if !bytes.Equal(resp[received-2:], msgEnd[:]) {
		err := errors.New("receiving from serial: end of message (0xAA, 0x55) not found")
		return nil, err
	}

	payload := resp[5 : packetLen+5-2]
	ch := crc16.Checksum(payload, crc16.CCITTTable)
	cp := binary.BigEndian.Uint16(resp[packetLen+5-2 : packetLen+5])
	if ch != cp {
		err := errors.New("receiving from serial: signature of received message is not valid")
		return nil, err
	}

	return payload, nil
}

func (s *Serial) Close() error {
	return s.port.Close()
}

func encode(mType MsgType, msg []byte) []byte {
	packet := append(msgStart[:], byte(mType))

	bLen := make([]byte, 2)
	binary.BigEndian.PutUint16(bLen, (uint16(len(msg) + 2)))
	packet = append(packet, bLen...)

	ch := crc16.Checksum(msg, crc16.CCITTTable)
	checksum := make([]byte, 2)
	binary.BigEndian.PutUint16(checksum, ch)
	packet = append(packet, msg...)
	packet = append(packet, checksum...)

	packet = append(packet, msgEnd[:]...)
	return packet
}
