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
	"math"
	"math/rand"
	"sort"
	"testing"
)

const (
	benchSize = 1000000
	zipfV     = 1.0
	zipfS     = 1.1
)

var (
	testSizes = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16,
		17, 18, 19, 1022, 1023, 1024, 1025, 1026, 1027, 1028}
	// uint32
	fourByteUint32Data       = make([]uint32, benchSize)
	threeByteUint32Data      = make([]uint32, benchSize)
	twoByteUint32Data        = make([]uint32, benchSize)
	oneByteUint32Data        = make([]uint32, benchSize)
	fourByteDeltaUint32Data  = makeUniformDeltaUint32(benchSize, 0, 0xDEADBEEF)
	threeByteDeltaUint32Data = makeUniformDeltaUint32(benchSize, 0, 0xADBEEF)
	twoByteDeltaUint32Data   = makeUniformDeltaUint32(benchSize, 0, 0xBEEF)
	oneByteDeltaUint32Data   = makeUniformDeltaUint32(benchSize, 0, 0xEF)
	benchUint32Data          = make([]uint32, benchSize)
	benchUint32DataSorted    = make([]uint32, benchSize)
	// int32
	fourByteInt32Data       = make([]int32, benchSize)
	threeByteInt32Data      = make([]int32, benchSize)
	twoByteInt32Data        = make([]int32, benchSize)
	oneByteInt32Data        = make([]int32, benchSize)
	fourByteDeltaInt32Data  = makeUniformDeltaInt32(benchSize, 0, 0x3EADBEEF)
	threeByteDeltaInt32Data = makeUniformDeltaInt32(benchSize, 0, 0x3DBEEF)
	twoByteDeltaInt32Data   = makeUniformDeltaInt32(benchSize, 0, 0x3EEF)
	oneByteDeltaInt32Data   = makeUniformDeltaInt32(benchSize, 0, 0x3F)
	benchInt32Data          = make([]int32, benchSize)
	benchInt32DataSorted    = make([]int32, benchSize)
	// byte
	benchEncoded     = make([]byte, MaxSize32(len(benchUint32Data)))
	benchEncodedSize int
)

func init() {
	// benchmark data
	seed := int64(42)
	zipf := rand.NewZipf(rand.New(rand.NewSource(seed)), zipfS, zipfV, math.MaxUint32)
	for i := range benchUint32Data {
		randUint32 := uint32(zipf.Uint64())
		benchUint32Data[i] = randUint32
		benchInt32Data[i] = int32((randUint32 >> 1) ^ -(randUint32 & 1))
	}
	copy(benchUint32DataSorted, benchUint32Data)
	sort.Slice(benchUint32DataSorted, func(i, j int) bool { return benchUint32DataSorted[i] < benchUint32DataSorted[j] })

	copy(benchInt32DataSorted, benchInt32Data)
	sort.Slice(benchInt32DataSorted, func(i, j int) bool { return benchInt32DataSorted[i] < benchInt32DataSorted[j] })

	// test data
	for i := range fourByteUint32Data {
		// uint32
		fourByteUint32Data[i] = uint32(0xDEADBEEF)
		threeByteUint32Data[i] = uint32(0xADBEEF)
		twoByteUint32Data[i] = uint32(0xBEEF)
		oneByteUint32Data[i] = uint32(0xEF)
		// int32
		fourByteInt32Data[i] = int32(0x3EADBEEF) * int32(1-(i&1)<<1)
		threeByteInt32Data[i] = int32(0x3DBEEF) * int32(1-(i&1)<<1)
		twoByteInt32Data[i] = int32(0x3EEF) * int32(1-(i&1)<<1)
		oneByteInt32Data[i] = int32(0x3F) * int32(1-(i&1)<<1)
	}
}

// testRoundTripUint32 tests that encoder and decdoder correctly round trip a slice
// of inputData of length size.  If width is in (1, 2, 3, 4) the input values
// will be truncated to that width bytes and the encoded size will be verified.
func testRoundTripUint32(t *testing.T, encoder func([]byte, []uint32) int, decoder func([]uint32, []byte), data []uint32, expectedSize int) {
	encodedRaw := make([]byte, MaxSize32(len(data)))
	encodedSize := encoder(encodedRaw, data)
	if expectedSize >= 0 && encodedSize != expectedSize {
		t.Errorf("got encodedSize: %d, expected: %d", encodedSize, expectedSize)
	}
	encoded := make([]byte, encodedSize, encodedSize) // ensure the encoded size is precise
	copy(encoded, encodedRaw)
	decodedData := make([]uint32, len(data), len(data))
	decoder(decodedData, encoded)
	for i := range data {
		if decodedData[i] != data[i] {
			t.Errorf("got decodedData[%d]: %d, expected: %d", i, decodedData[i], data[i])
		}
	}
}

// testRoundTripInt32 tests that encoder and decdoder correctly round trip a slice
// of inputData of length size.  If width is in (1, 2, 3, 4) the input values
// will be truncated to that width bytes and the encoded size will be verified.
func testRoundTripInt32(t *testing.T, encoder func([]byte, []int32) int, decoder func([]int32, []byte), data []int32, expectedSize int) {
	encodedRaw := make([]byte, MaxSize32(len(data)))
	encodedSize := encoder(encodedRaw, data)
	if expectedSize >= 0 && encodedSize != expectedSize {
		t.Errorf("got encodedSize: %d, expected: %d", encodedSize, expectedSize)
	}
	encoded := make([]byte, encodedSize, encodedSize) // ensure the encoded size is precise
	copy(encoded, encodedRaw)
	decodedData := make([]int32, len(data), len(data))
	decoder(decodedData, encoded)
	for i := range data {
		if decodedData[i] != data[i] {
			t.Errorf("got decodedData[%d]: %d, expected: %d", i, decodedData[i], data[i])
		}
	}
}

func makeUniformDeltaUint32(size int, previous, delta uint32) []uint32 {
	uniformDelta := make([]uint32, size, size)
	for i := range uniformDelta {
		previous += delta
		uniformDelta[i] = previous
	}
	return uniformDelta
}

func makeUniformDeltaInt32(size int, previous, delta int32) []int32 {
	uniformDelta := make([]int32, size, size)
	for i := range uniformDelta {
		previous += delta
		uniformDelta[i] = previous
	}
	return uniformDelta
}

// uint32

func testUniformAndRandomUint32(t *testing.T, encoder func([]byte, []uint32) int, decoder func([]uint32, []byte)) {
	for _, size := range testSizes {
		expectedSize := (size+3)/4 + size
		testRoundTripUint32(t, encoder, decoder, oneByteUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*2
		testRoundTripUint32(t, encoder, decoder, twoByteUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*3
		testRoundTripUint32(t, encoder, decoder, threeByteUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*4
		testRoundTripUint32(t, encoder, decoder, fourByteUint32Data[0:size:size], expectedSize)
		testRoundTripUint32(t, encoder, decoder, benchUint32Data[0:size:size], -1)
	}
}

func testUniformDeltaAndRandomUint32(t *testing.T, encoder func([]byte, []uint32) int, decoder func([]uint32, []byte)) {
	for _, size := range testSizes {
		expectedSize := (size+3)/4 + size
		testRoundTripUint32(t, encoder, decoder, oneByteDeltaUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*2
		testRoundTripUint32(t, encoder, decoder, twoByteDeltaUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*3
		testRoundTripUint32(t, encoder, decoder, threeByteDeltaUint32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*4
		testRoundTripUint32(t, encoder, decoder, fourByteDeltaUint32Data[0:size:size], expectedSize)
		testRoundTripUint32(t, encoder, decoder, benchUint32Data[0:size:size], -1)
	}
}

func TestRoundTripUint32(t *testing.T) {
	testUniformAndRandomUint32(t, EncodeUint32, DecodeUint32)
}

func EncodeDeltaUint32Test(encoded []byte, data []uint32) int {
	return EncodeDeltaUint32(encoded, data, 0)
}
func DecodeDeltaUint32Test(data []uint32, encoded []byte) {
	DecodeDeltaUint32(data, encoded, 0)
}

func TestRoundTripDeltaUint32(t *testing.T) {
	testUniformDeltaAndRandomUint32(t, EncodeDeltaUint32Test, DecodeDeltaUint32Test)
}

// int32

func testUniformAndRandomInt32(t *testing.T, encoder func([]byte, []int32) int, decoder func([]int32, []byte)) {
	for _, size := range testSizes {
		expectedSize := (size+3)/4 + size
		testRoundTripInt32(t, encoder, decoder, oneByteInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*2
		testRoundTripInt32(t, encoder, decoder, twoByteInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*3
		testRoundTripInt32(t, encoder, decoder, threeByteInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*4
		testRoundTripInt32(t, encoder, decoder, fourByteInt32Data[0:size:size], expectedSize)
		testRoundTripInt32(t, encoder, decoder, benchInt32Data[0:size:size], -1)
	}
}

func testUniformDeltaAndRandomInt32(t *testing.T, encoder func([]byte, []int32) int, decoder func([]int32, []byte)) {
	for _, size := range testSizes {
		expectedSize := (size+3)/4 + size
		testRoundTripInt32(t, encoder, decoder, oneByteDeltaInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*2
		testRoundTripInt32(t, encoder, decoder, twoByteDeltaInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*3
		testRoundTripInt32(t, encoder, decoder, threeByteDeltaInt32Data[0:size:size], expectedSize)
		expectedSize = (size+3)/4 + size*4
		testRoundTripInt32(t, encoder, decoder, fourByteDeltaInt32Data[0:size:size], expectedSize)
		testRoundTripInt32(t, encoder, decoder, benchInt32Data[0:size:size], -1)
	}
}

func TestRoundTripInt32(t *testing.T) {
	testUniformAndRandomInt32(t, EncodeInt32, DecodeInt32)
}

func EncodeDeltaInt32Test(encoded []byte, data []int32) int {
	return EncodeDeltaInt32(encoded, data, 0)
}
func DecodeDeltaInt32Test(data []int32, encoded []byte) {
	DecodeDeltaInt32(data, encoded, 0)
}

func TestRoundTripDeltaInt32(t *testing.T) {
	testUniformDeltaAndRandomInt32(t, EncodeDeltaInt32Test, DecodeDeltaInt32Test)
}

func BenchmarkCopy32(b *testing.B) {
	b.SetBytes(int64(4 * benchSize))
	dummySink := make([]uint32, benchSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy(dummySink, benchUint32Data)
	}
}
