/*
Copyright (c) 2020 Brian M. Kessler

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package streamvbyte

import (
	"encoding/binary"
)

func encodeUint32scalar(encoded []byte, data []uint32) int {
	// index of the control bytes
	ci := 0
	// index of the data bytes
	di := (len(data) + 3) >> 2

	controlByte := byte(0)
	for i, v := range data {
		controlByte >>= 2
		switch {
		case v < 1<<8:
			encoded[di] = byte(v)
			di++
		case v < 1<<16:
			binary.LittleEndian.PutUint16(encoded[di:], uint16(v))
			di += 2
			controlByte ^= 0b_01_00_00_00
		case v < 1<<24:
			encoded[di] = byte(v)
			encoded[di+1] = byte(v >> 8)
			encoded[di+2] = byte(v >> 16)
			di += 3
			controlByte ^= 0b_10_00_00_00
		default:
			binary.LittleEndian.PutUint32(encoded[di:], uint32(v))
			di += 4
			controlByte ^= 0b_11_00_00_00
		}
		if (i+1)&3 == 0 {
			encoded[ci] = controlByte
			controlByte = 0
			ci++
		}
	}
	// Check if the last block was complete or the control byte
	// needs to be shifted and written.
	if rem := len(data) & 3; rem != 0 {
		shift := uint(4-rem) * 2
		encoded[ci] = controlByte >> shift
	}
	return di
}

func decodeUint32scalar(data []uint32, encoded []byte) {
	// index of the control bytes
	ci := 0
	// index of the data bytes
	di := (len(data) + 3) >> 2

	var controlByte byte
	for i := range data {
		if i&3 == 0 {
			controlByte = encoded[ci]
			ci++
		}
		switch controlByte & 3 {
		case 0:
			data[i] = uint32(encoded[di])
			di++
		case 1:
			v := binary.LittleEndian.Uint16(encoded[di:])
			data[i] = uint32(v)
			di += 2
		case 2:
			data[i] = uint32(encoded[di+2])<<16 | uint32(encoded[di+1])<<8 | uint32(encoded[di])
			di += 3
		default:
			data[i] = binary.LittleEndian.Uint32(encoded[di:])
			di += 4
		}
		controlByte >>= 2
	}
}
