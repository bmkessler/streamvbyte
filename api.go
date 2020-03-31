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

/*
Package streamvbyte implements integer compression using the Stream VByte algorithm.

"Stream VByte: Faster Byte-Oriented Integer Compression"
Daniel Lemire, Nathan Kurz, Christoph Rupp
Information Processing Letters
Volume 130, February 2018, Pages 1-6
https://arxiv.org/abs/1709.08990

The focus of this method is fast decoding speed on processors with
SIMD instructions.  As such, this package contains fast assembly implementations
for decoding on amd64 processors with SSE3 instructions.  Other processors
will use the slower pure go implementation.

Assembly implementations were generated using the excellent avo package https://github.com/mmcloughlin/avo
*/
package streamvbyte

// MaxSize32 returns the maximum possible size of an encoded
// slice of 32-bit integers. Usage:
//
//   encoded := make([]byte, MaxSize32(len(data)))
//
// This will ensure that the slice is large enough to hold the
// encoded data in the worst case of no compression.
func MaxSize32(length int) int {
	numControlBytes := (length + 3) / 4
	maxNumDataBytes := 4 * length
	return numControlBytes + maxNumDataBytes
}

// EncodeUint32 encodes data using the Stream VByte
// algorithm into encoded and returns the encoded size.
// This function assumes that the size of encoded is sufficient to hold the
// compressed data.  Use
//   encoded := make([]byte, MaxSize32(len(data)))
// to obtain a worst case size.
func EncodeUint32(encoded []byte, data []uint32) int {
	return encodeUint32scalar(encoded, data)
}

// DecodeUint32 decodes len(data) uint32 from encoded using the Stream
// Vbyte algorithm.  encoded must contain exactly len(data) encoded uint32.
func DecodeUint32(data []uint32, encoded []byte) {
	decodeUint32(data, encoded)
}

// EncodeDeltaUint32 encodes data using the Stream VByte
// algorithm and delta encoding with a step size of 1, i.e. it encodes
//   delta[n] = data[n] - data[n-1],
// where the initial value
//   data[-1] := previous
// The return value is the encoded size.  This function assumes
// that the size of encoded is sufficient to hold the
// compressed data.  Use
//   encoded := make([]byte, MaxSize32(len(data)))
// to obtain a worst case size.
func EncodeDeltaUint32(encoded []byte, data []uint32, previous uint32) int {
	return encodeDeltaUint32scalar(encoded, data, previous)
}

// DecodeDeltaUint32 decodes len(data) uint32 from encoded using the Stream
// Vbyte algorithm with delta encoding using the initial value previous.
// encoded must contain exactly len(data) encoded uint32.
func DecodeDeltaUint32(data []uint32, encoded []byte, previous uint32) {
	decodeDeltaUint32(data, encoded, previous)
}

// EncodeInt32 encodes data using the Stream VByte
// algorithm with zigzag encoding into encoded and returns the encoded size.
// This function assumes that the size of encoded is sufficient to hold the
// compressed data.  Use
//   encoded := make([]byte, MaxSize32(len(data)))
// to obtain a worst case size.
func EncodeInt32(encoded []byte, data []int32) int {
	return encodeInt32scalar(encoded, data)
}

// DecodeInt32 decodes len(data) int32 from encoded using the Stream
// Vbyte algorithm.  encoded must contain exactly len(data) encoded int32.
func DecodeInt32(data []int32, encoded []byte) {
	decodeInt32(data, encoded)
}

// EncodeDeltaInt32 encodes data using the Stream VByte
// algorithm and delta encoding with a step size of 1, i.e. it encodes
//   delta[n] = data[n] - data[n-1]
// where the initial value
//   data[-1] := previous
// followed by zigzag encoding the deltas.
// The return value is the encoded size.  This function assumes
// that the size of encoded is sufficient to hold the
// compressed data.  Use
//   encoded := make([]byte, MaxSize32(len(data)))
// to obtain a worst case size.
func EncodeDeltaInt32(encoded []byte, data []int32, previous int32) int {
	return encodeDeltaInt32scalar(encoded, data, previous)
}

// DecodeDeltaInt32 decodes len(data) int32 from encoded using the Stream
// Vbyte algorithm with delta and zigzag encoding using the initial value previous.
// encoded must contain exactly len(data) encoded uint32.
func DecodeDeltaInt32(data []int32, encoded []byte, previous int32) {
	decodeDeltaInt32(data, encoded, previous)
}
