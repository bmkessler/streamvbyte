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
	"testing"

	"golang.org/x/sys/cpu"
)

// uint32
func TestRoundTripUint32SSE3(t *testing.T) {
	if !cpu.X86.HasSSE3 {
		t.Skip("CPU does not support SSE3 instructions")
	}
	testUniformAndRandomUint32(t, encodeUint32scalar, decodeUint32SSE3)
}

func decodeDeltaUint32SSE3Test(data []uint32, encoded []byte) {
	decodeDeltaUint32SSE3(data, encoded, 0)
}

func TestRoundTripDeltaUint32SSE3(t *testing.T) {
	testUniformDeltaAndRandomUint32(t, encodeDeltaUint32scalarTest, decodeDeltaUint32SSE3Test)
}

func BenchmarkDecodeUint32SSE3(b *testing.B) {
	if !cpu.X86.HasSSE3 {
		b.Skip("CPU does not support SSE3 instructions")
	}
	b.SetBytes(int64(4 * benchSize))
	benchEncodedSize = encodeUint32scalar(benchEncoded, benchUint32Data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeUint32SSE3(benchUint32Data, benchEncoded)
	}
}

func BenchmarkDecodeDeltaUint32SSE3(b *testing.B) {
	if !cpu.X86.HasSSE3 {
		b.Skip("CPU does not support SSE3 instructions")
	}
	b.SetBytes(int64(4 * benchSize))
	benchEncodedSize = encodeDeltaUint32scalar(benchEncoded, benchUint32DataSorted, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeDeltaUint32SSE3(benchUint32DataSorted, benchEncoded, 0)
	}
}

// int32

func TestRoundTripInt32SSE3(t *testing.T) {
	if !cpu.X86.HasSSE3 {
		t.Skip("CPU does not support SSE3 instructions")
	}
	testUniformAndRandomInt32(t, encodeInt32scalar, decodeInt32SSE3)
}

func decodeDeltaInt32SSE3Test(data []int32, encoded []byte) {
	decodeDeltaInt32SSE3(data, encoded, 0)
}

func TestRoundTripDeltaInt32SSE3(t *testing.T) {
	testUniformDeltaAndRandomInt32(t, encodeDeltaInt32scalarTest, decodeDeltaInt32SSE3Test)
}

func BenchmarkDecodeInt32SSE3(b *testing.B) {
	if !cpu.X86.HasSSE3 {
		b.Skip("CPU does not support SSE3 instructions")
	}
	b.SetBytes(int64(4 * benchSize))
	benchEncodedSize = encodeInt32scalar(benchEncoded, benchInt32Data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeInt32SSE3(benchInt32Data, benchEncoded)
	}
}

func BenchmarkDecodeDeltaInt32SSE3(b *testing.B) {
	if !cpu.X86.HasSSE3 {
		b.Skip("CPU does not support SSE3 instructions")
	}
	b.SetBytes(int64(4 * benchSize))
	benchEncodedSize = encodeDeltaInt32scalar(benchEncoded, benchInt32DataSorted, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeDeltaInt32SSE3(benchInt32DataSorted, benchEncoded, 0)
	}
}
