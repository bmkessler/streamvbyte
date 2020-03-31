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
)

func encodeDeltaInt32scalarTest(encoded []byte, data []int32) int {
	return encodeDeltaInt32scalar(encoded, data, 0)
}

func decodeDeltaInt32scalarTest(data []int32, encoded []byte) {
	decodeDeltaInt32scalar(data, encoded, 0)
}
func TestRoundTripDeltaInt32Scalar(t *testing.T) {
	testUniformDeltaAndRandomInt32(t, encodeDeltaInt32scalarTest, decodeDeltaInt32scalarTest)
}

func BenchmarkEncodeDeltaInt32Scalar(b *testing.B) {
	b.SetBytes(int64(4 * benchSize))
	for i := 0; i < b.N; i++ {
		benchEncodedSize = encodeDeltaInt32scalar(benchEncoded, benchInt32DataSorted, 0)
	}
}

func BenchmarkDecodeDeltaInt32Scalar(b *testing.B) {
	b.SetBytes(int64(4 * benchSize))
	benchEncodedSize = encodeDeltaInt32scalar(benchEncoded, benchInt32DataSorted, 0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decodeDeltaInt32scalar(benchInt32DataSorted, benchEncoded, 0)
	}
}
