// This code is a go port of LZSS encoder-decoder (Haruhiko Okumura; public domain)

package lzss

import (
	"bytes"

	"github.com/icza/bitio"
)

func Decompress(data []byte) []byte {
	input := bitio.NewReader(bytes.NewBuffer(data))
	output := make([]byte, 0)

	buffer := make([]byte, bufsz*2)
	for i := 0; i < bufsz-looksz; i++ {
		buffer[i] = ' '
	}

	r := bufsz - looksz
	var char byte
	var isChar bool
	var err error
	for {
		isChar, err = input.ReadBool()
		if err != nil {
			break
		}

		if isChar {
			char, err = input.ReadByte()
			if err != nil {
				break
			}
			output = append(output, char)
			buffer[r] = char
			r++
			r &= bufsz - 1
		} else {
			var i, j uint64
			i, err = input.ReadBits(idxsz)
			if err != nil {
				break
			}
			j, err = input.ReadBits(lensz)
			if err != nil {
				break
			}

			for k := 0; k <= int(j)+1; k++ {
				char = buffer[(int(i)+k)&(bufsz-1)]
				output = append(output, char)
				buffer[r] = char
				r++
				r &= bufsz - 1
			}
		}
	}

	return output
}
