//go:generate go run gen_decode_sse3.go -out decode_sse3_amd64.s

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
	"golang.org/x/sys/cpu"
)

// uint32

func decodeUint32(data []uint32, encoded []byte) {
	if cpu.X86.HasSSE3 {
		decodeUint32SSE3(data, encoded)
		return
	}
	decodeUint32scalar(data, encoded)
	return
}

func decodeUint32SSE3(data []uint32, encoded []byte)

func decodeDeltaUint32(data []uint32, encoded []byte, previous uint32) {
	if cpu.X86.HasSSE3 {
		decodeDeltaUint32scalar(data, encoded, previous)
		return
	}
	decodeDeltaUint32scalar(data, encoded, previous)
	return
}

func decodeDeltaUint32SSE3(data []uint32, encoded []byte, previous uint32)

// int32

func decodeInt32(data []int32, encoded []byte) {
	if cpu.X86.HasSSE3 {
		decodeInt32SSE3(data, encoded)
		return
	}
	decodeInt32scalar(data, encoded)
	return
}

func decodeInt32SSE3(data []int32, encoded []byte)

func decodeDeltaInt32(data []int32, encoded []byte, previous int32) {
	if cpu.X86.HasSSE3 {
		decodeDeltaInt32scalar(data, encoded, previous)
		return
	}
	decodeDeltaInt32scalar(data, encoded, previous)
	return
}

func decodeDeltaInt32SSE3(data []int32, encoded []byte, previous int32)
