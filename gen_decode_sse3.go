// +build ignore

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

package main

import (
	"encoding/binary"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

// preamble loads the input data and returns variables referencing those values
func preamble(dataByteCount, dataByteMask Mem) (encoded Mem, encodedCap Register, data Mem, dataLen Register, dataTail GPVirtual, ci GPVirtual, di GPVirtual, n GPVirtual, byteCountPtr Mem, byteMaskptr Mem) {
	encoded = Mem{Base: Load(Param("encoded").Base(), GP64())}
	encodedCap = Load(Param("encoded").Cap(), GP64())
	Comment("Revert to scalar processing if we are within 16 bytes of the end.")
	SUBQ(Imm(16), encodedCap)

	data = Mem{Base: Load(Param("data").Base(), GP64())}
	dataLen = Load(Param("data").Len(), GP64())
	dataTail = GP64()
	Comment("Revert to scalar processing if we have less than 4 values to process.")
	MOVQ(dataLen, dataTail)
	SUBQ(Imm(4), dataTail)

	Comment("Initialize the control index.")
	ci = GP64()
	XORQ(ci, ci)

	Comment("Initialize the data index. (len(data) + 3) >> 2")
	di = GP64()
	MOVQ(dataLen, di)
	ADDQ(Imm(3), di)
	SHRQ(Imm(2), di)

	Comment("Initialize the output index.")
	n = GP64()
	XORQ(n, n)

	Comment("The byte count lookup table.")
	byteCountPtr = Mem{Base: GP64()}
	LEAQ(dataByteCount, byteCountPtr.Base)

	Comment("The byte mask lookup table.")
	byteMaskptr = Mem{Base: GP64()}
	LEAQ(dataByteMask, byteMaskptr.Base)
	return encoded, encodedCap, data, dataLen, dataTail, ci, di, n, byteCountPtr, byteMaskptr
}

// decodeSIMDUint32 reads control byte and 4 uint32 from data bytes and returns the count of bytes read aong with the dataBytes
func decodeSIMDUint32(encoded Mem, ci, di GPVirtual, byteCountPtr, byteMaskptr Mem) (VecVirtual, GPVirtual) {
	Comment("Load control byte.")
	cb := GP64()
	MOVBQZX(encoded.Idx(ci, 1), cb)
	INCQ(ci)

	Comment("Load 16 data bytes into XMM.")
	dataBytes := XMM()
	MOVOU(encoded.Idx(di, 1), dataBytes)

	Comment("Lookup count to increment data index.")
	byteCount := GP64()
	MOVBQZX(byteCountPtr.Idx(cb, 1), byteCount)

	Comment("Lookup the PSHUFB mask.")
	SHLQ(Imm(4), cb)

	Comment("Use mask to shuffle the relevant bytes into place.")
	PSHUFB(byteMaskptr.Idx(cb, 1), dataBytes)

	return dataBytes, byteCount
}

// decodeScalarUint32 reads control byte and returns the decoded uint32 value
func decodeScalarUint32(n, ci, di GPVirtual, encoded, data Mem) (val GPVirtual) {
	Comment("Determine if we need to load a new control byte.")
	TESTQ(U32(3), n)
	JNE(LabelRef("loadBytes"))

	Comment("Load control byte.")
	cb := GP64()
	MOVBQZX(encoded.Idx(ci, 1), cb)
	INCQ(ci)

	Label("loadBytes")
	Comment("Switch on the low two bits of the control byte.")
	switchVal := GP64()
	MOVQ(cb, switchVal)
	ANDQ(Imm(3), switchVal)

	JE(LabelRef("oneByte"))
	CMPQ(switchVal, Imm(1))
	JE(LabelRef("twoByte"))
	CMPQ(switchVal, Imm(2))
	JE(LabelRef("threeByte"))

	val = GP32()

	Label("fourByte")
	MOVL(encoded.Idx(di, 1), val) // val = binary.LittleEndian.Uint32(encoded[di:])
	ADDQ(Imm(4), di)              // di += 4
	JMP(LabelRef("shiftControl"))

	Label("threeByte")
	hi := GP32()
	MOVWLZX(encoded.Idx(di, 1), val)          // val = uint32(binary.LittleEndian.Uint16(encoded[di:]))
	MOVBLZX(encoded.Idx(di, 1).Offset(2), hi) // hi = uint32(encoded[di+2])
	SHLL(Imm(16), hi)                         // hi <<= 16
	ORL(hi, val)                              // val = (hi | val)
	ADDQ(Imm(3), di)                          // di +=3
	JMP(LabelRef("shiftControl"))

	Label("twoByte")
	MOVWLZX(encoded.Idx(di, 1), val) // val = uint32(binary.LittleEndian.Uint16(encoded[di:]))
	ADDQ(Imm(2), di)                 // di += 2
	JMP(LabelRef("shiftControl"))

	Label("oneByte")
	MOVBLZX(encoded.Idx(di, 1), val) // val = uint32(encoded[di])
	INCQ(di)                         // di++

	Label("shiftControl")
	Comment("Shift control byte to get next value.")
	SHRQ(Imm(2), cb)

	return val
}

func prefixSumSIMD(dataBytes, previousX VecVirtual) {
	shifted := XMM()
	Comment("Calculate prefix sum.")
	//copy dataBytes to shifted
	MOVOU(dataBytes, shifted)
	Comment("(0, 0, delta_0, delta_1)")
	PSLLDQ(Imm(8), shifted)
	Comment("(delta_0, delta_1, delta_2 + delta_0, delta_3 + delta_1)")
	PADDD(shifted, dataBytes)
	// copy dataBytes to shifted
	MOVOU(dataBytes, shifted)
	Comment("(0, delta_0, delta_1, delta_2 + delta_0)")
	PSLLDQ(Imm(4), shifted)
	Comment("(delta_0, delta_0 + delta_1, delta_0 + delta_1 + delta_2, delta_0 + delta_1 + delta_2 + delta_delta_3)")
	PADDD(shifted, dataBytes)
	Comment("Add the previous last decoded value to all lanes.")
	PADDD(previousX, dataBytes)
	Comment("Propagate last decoded value to all lanes of previous.")
	PSHUFD(Imm(0b_11_11_11_11), dataBytes, previousX)
}

func zigzagDecodeScalar(val GPVirtual) {
	Comment("Zigzag decode.")
	tmp := GP32()
	MOVL(val, tmp)
	SHRL(Imm(1), tmp)
	ANDL(Imm(1), val)
	NEGL(val)
	XORL(tmp, val)
}

func zigzagDecodeSIMD(dataBytes VecVirtual) {
	Comment("Zigzag decode.")
	tmpX := XMM()
	MOVOU(dataBytes, tmpX)
	Comment("(x >> 1)")
	PSRLL(Imm(1), tmpX)
	oneX := XMM()
	Comment("Set to all ones.")
	PCMPEQL(oneX, oneX)
	Comment("Shift to one in each lane.")
	PSRLL(Imm(31), oneX)
	Comment("(x & 1)")
	PAND(dataBytes, oneX)
	Comment("Set to all zeroes.")
	PXOR(dataBytes, dataBytes)
	Comment("-(x & 1)")
	PSUBL(oneX, dataBytes)
	Comment("(x >> 1) ^ - (x & 1)")
	PXOR(tmpX, dataBytes)
}

func main() {

	// Lookup table of the count of data bytes (4 to 16) referenced by a control byte.
	dataByteCount := GLOBL("dataByteCount", RODATA|NOPTR)
	for i := 0; i < 256; i++ {
		count := byte(i&3) + byte((i>>2)&3) + byte((i>>4)&3) + byte((i>>6)&3) + 4
		DATA(i, U8(count))
	}

	// Lookup table of the PSUFB mask referenced by a control byte to move data bytes
	// into the correct location.
	dataByteMask := GLOBL("dataByteMask", RODATA|NOPTR)
	for i := 0; i < 256; i++ {
		curIndex, controlByte := byte(0), byte(i)
		mask := [16]byte{}
		for j := 0; j < 4; j++ {
			byteCount := controlByte & 3
			for k := 0; k < 4; k++ {
				if k <= int(byteCount) {
					mask[4*j+k] = curIndex
					curIndex++
				} else {
					mask[4*j+k] = 0xFF
				}
			}
			controlByte >>= 2
		}
		lowerHalf := binary.LittleEndian.Uint64(mask[0:8])
		upperHalf := binary.LittleEndian.Uint64(mask[8:16])
		DATA(16*i, U64(lowerHalf))
		DATA(16*i+8, U64(upperHalf))
	}

	TEXT("decodeUint32SSE3", NOSPLIT, "func (data []uint32, encoded []byte)")
	Doc("decodeUint32SSE3 decodes 4 uint32 at a time using SSE3 instructions (PSHUFB)")
	{
		encoded, encodedCap, data, dataLen, dataTail, ci, di, n, byteCountPtr, byteMaskptr := preamble(dataByteCount, dataByteMask)

		Label("simd")
		Comment("Check if less than 16 encoded bytes remain and jump to scalar.")
		CMPQ(di, encodedCap)
		JGT(LabelRef("scalar"))
		Comment("Check if less than 4 values remain and jump to scalar.")
		CMPQ(n, dataTail)
		JGT(LabelRef("scalar"))

		dataBytes, bytecount := decodeSIMDUint32(encoded, ci, di, byteCountPtr, byteMaskptr)

		Comment("Store 4 uint32.")
		MOVOU(dataBytes, data.Idx(n, 4))

		Comment("Increment the indices.")
		ADDQ(Imm(4), n)
		ADDQ(bytecount, di)

		JMP(LabelRef("simd"))

		Label("scalar")
		Comment("Process a single value at a time.")

		CMPQ(n, dataLen)
		JE(LabelRef("done"))

		val := decodeScalarUint32(n, ci, di, encoded, data)

		MOVL(val, data.Idx(n, 4)) // data[i] = val
		INCQ(n)
		JMP(LabelRef("scalar"))

		Label("done")
		RET()
	}

	TEXT("decodeDeltaUint32SSE3", NOSPLIT, "func (data []uint32, encoded []byte, previous uint32)")
	Doc("decodeDeltaUint32SSE3 decodes 4 uint32 at a time using SSE3 instructions (PSHUFB)")
	{
		encoded, encodedCap, data, dataLen, dataTail, ci, di, n, byteCountPtr, byteMaskptr := preamble(dataByteCount, dataByteMask)
		previous := Load(Param("previous"), GP32())

		previousX := XMM()
		MOVD(previous, previousX)
		PSHUFD(Imm(0b_00_00_00_00), previousX, previousX)

		Label("simd")
		Comment("Check if less than 16 encoded bytes remain and jump to scalar.")
		CMPQ(di, encodedCap)
		JGT(LabelRef("scalar"))
		Comment("Check if less than 4 values remain and jump to scalar.")
		CMPQ(n, dataTail)
		JGT(LabelRef("scalar"))

		dataBytes, bytecount := decodeSIMDUint32(encoded, ci, di, byteCountPtr, byteMaskptr)

		prefixSumSIMD(dataBytes, previousX)

		MOVD(previousX, previous)

		Comment("Store 4 uint32.")
		MOVOU(dataBytes, data.Idx(n, 4))

		Comment("Increment the indices.")
		ADDQ(Imm(4), n)
		ADDQ(bytecount, di)

		JMP(LabelRef("simd"))

		Label("scalar")
		Comment("Process a single value at a time.")

		CMPQ(n, dataLen)
		JE(LabelRef("done"))

		val := decodeScalarUint32(n, ci, di, encoded, data)

		Comment("Add the previous decoded value to the delta.")
		ADDL(val, previous)            // previous += val
		MOVL(previous, data.Idx(n, 4)) // data[i] = val
		INCQ(n)
		JMP(LabelRef("scalar"))

		Label("done")

		RET()
	}

	TEXT("decodeInt32SSE3", NOSPLIT, "func (data []int32, encoded []byte)")
	Doc("decodeInt32SSE3 decodes 4 int32 at a time using SSE3 instructions (PSHUFB)")
	{
		encoded, encodedCap, data, dataLen, dataTail, ci, di, n, byteCountPtr, byteMaskptr := preamble(dataByteCount, dataByteMask)

		Label("simd")
		Comment("Check if less than 16 encoded bytes remain and jump to scalar.")
		CMPQ(di, encodedCap)
		JGT(LabelRef("scalar"))
		Comment("Check if less than 4 values remain and jump to scalar.")
		CMPQ(n, dataTail)
		JGT(LabelRef("scalar"))

		dataBytes, bytecount := decodeSIMDUint32(encoded, ci, di, byteCountPtr, byteMaskptr)

		zigzagDecodeSIMD(dataBytes)

		Comment("Store 4 uint32.")
		MOVOU(dataBytes, data.Idx(n, 4))

		Comment("Increment the indices.")
		ADDQ(Imm(4), n)
		ADDQ(bytecount, di)

		JMP(LabelRef("simd"))

		Label("scalar")
		Comment("Process a single value at a time.")

		CMPQ(n, dataLen)
		JE(LabelRef("done"))

		val := decodeScalarUint32(n, ci, di, encoded, data)
		zigzagDecodeScalar(val)

		MOVL(val, data.Idx(n, 4)) // data[i] = val
		INCQ(n)
		JMP(LabelRef("scalar"))

		Label("done")
		RET()
	}

	TEXT("decodeDeltaInt32SSE3", NOSPLIT, "func (data []int32, encoded []byte, previous int32)")
	Doc("decodeDeltaInt32SSE3 decodes 4 int32 at a time using SSE3 instructions (PSHUFB)")
	{
		encoded, encodedCap, data, dataLen, dataTail, ci, di, n, byteCountPtr, byteMaskptr := preamble(dataByteCount, dataByteMask)
		previous := Load(Param("previous"), GP32())

		previousX := XMM()
		MOVD(previous, previousX)
		PSHUFD(Imm(0b_00_00_00_00), previousX, previousX)

		Label("simd")
		Comment("Check if less than 16 encoded bytes remain and jump to scalar.")
		CMPQ(di, encodedCap)
		JGT(LabelRef("scalar"))
		Comment("Check if less than 4 values remain and jump to scalar.")
		CMPQ(n, dataTail)
		JGT(LabelRef("scalar"))

		dataBytes, bytecount := decodeSIMDUint32(encoded, ci, di, byteCountPtr, byteMaskptr)

		zigzagDecodeSIMD(dataBytes)
		prefixSumSIMD(dataBytes, previousX)

		MOVD(previousX, previous)

		Comment("Store 4 uint32.")
		MOVOU(dataBytes, data.Idx(n, 4))

		Comment("Increment the indices.")
		ADDQ(Imm(4), n)
		ADDQ(bytecount, di)

		JMP(LabelRef("simd"))

		Label("scalar")
		Comment("Process a single value at a time.")

		CMPQ(n, dataLen)
		JE(LabelRef("done"))

		val := decodeScalarUint32(n, ci, di, encoded, data)
		zigzagDecodeScalar(val)

		Comment("Add the previous decoded value to the delta.")
		ADDL(val, previous)            // previous += val
		MOVL(previous, data.Idx(n, 4)) // data[i] = previous
		INCQ(n)
		JMP(LabelRef("scalar"))

		Label("done")

		RET()
	}

	Generate()
}
