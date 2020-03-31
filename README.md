# streamvbyte

This package implements integer compression using the Stream VByte algorithm.

"Stream VByte: Faster Byte-Oriented Integer Compression"
Daniel Lemire, Nathan Kurz, Christoph Rupp
Information Processing Letters
Volume 130, February 2018, Pages 1-6
https://arxiv.org/abs/1709.08990

The focus of this method is fast decoding speed on processors with
SIMD instructions.  As such, this package contains fast assembly implementations
for decoding on amd64 processors with SSE3 instructions.  Other processors
will use the slower pure go implementation.

Assembly implementations were generated using the excellent [avo](https://github.com/mmcloughlin/avo)

Reference benchmarks on 1,000,000 Zipfian-distributed 32-bit integers.

Intel(R) Core(TM) i3-4010U CPU @ 1.70GHz
go version go1.14.2 linux/amd64

| Type           | Decode (SSE3) | Decode (pure go) | Encode (pure go) | compression ratio |
| -------------- | ------------- | ---------------- | ---------------- | ----------------- |
| uint32         | 4.35GB/s ± 1% |   291MB/s ± 1%   |   272MB/s ± 0%   |       0.51        |
| uint32 (delta) | 3.90GB/s ± 4% |   890MB/s ± 5%   |   948MB/s ± 1%   |       0.34        |
| int32          | 3.99GB/s ± 4% |   276MB/s ± 0%   |   257MB/s ± 2%   |       0.51        |
| int32  (delta) | 3.08GB/s ± 3% |   887MB/s ± 1%   |   866MB/s ± 2%   |       0.35        |

For reference, a pure memory copy of the uncompressed data runs at 6.58GB/s ± 2%.
