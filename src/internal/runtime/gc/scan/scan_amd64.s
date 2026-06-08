// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

GLOBL expandAVX512_sparse_inShuf0<>(SB), RODATA, $0x40
DATA  expandAVX512_sparse_inShuf0<>+0x00(SB)/8, $0x0706050403020100
DATA  expandAVX512_sparse_inShuf0<>+0x08(SB)/8, $0x0706050403020100
DATA  expandAVX512_sparse_inShuf0<>+0x10(SB)/8, $0x0f0e0d0c0b0a0908
DATA  expandAVX512_sparse_inShuf0<>+0x18(SB)/8, $0x0f0e0d0c0b0a0908
DATA  expandAVX512_sparse_inShuf0<>+0x20(SB)/8, $0x1716151413121110
DATA  expandAVX512_sparse_inShuf0<>+0x28(SB)/8, $0x1716151413121110
DATA  expandAVX512_sparse_inShuf0<>+0x30(SB)/8, $0x1f1e1d1c1b1a1918
DATA  expandAVX512_sparse_inShuf0<>+0x38(SB)/8, $0x1f1e1d1c1b1a1918

GLOBL expandAVX512_sparse_mat0<>(SB), RODATA, $0x40
DATA  expandAVX512_sparse_mat0<>+0x00(SB)/8, $0x0101020204040808
DATA  expandAVX512_sparse_mat0<>+0x08(SB)/8, $0x1010202040408080
DATA  expandAVX512_sparse_mat0<>+0x10(SB)/8, $0x0101020204040808
DATA  expandAVX512_sparse_mat0<>+0x18(SB)/8, $0x1010202040408080
DATA  expandAVX512_sparse_mat0<>+0x20(SB)/8, $0x0101020204040808
DATA  expandAVX512_sparse_mat0<>+0x28(SB)/8, $0x1010202040408080
DATA  expandAVX512_sparse_mat0<>+0x30(SB)/8, $0x0101020204040808
DATA  expandAVX512_sparse_mat0<>+0x38(SB)/8, $0x1010202040408080

GLOBL expandAVX512_sparse_inShuf1<>(SB), RODATA, $0x40
DATA  expandAVX512_sparse_inShuf1<>+0x00(SB)/8, $0x2726252423222120
DATA  expandAVX512_sparse_inShuf1<>+0x08(SB)/8, $0x2726252423222120
DATA  expandAVX512_sparse_inShuf1<>+0x10(SB)/8, $0x2f2e2d2c2b2a2928
DATA  expandAVX512_sparse_inShuf1<>+0x18(SB)/8, $0x2f2e2d2c2b2a2928
DATA  expandAVX512_sparse_inShuf1<>+0x20(SB)/8, $0x3736353433323130
DATA  expandAVX512_sparse_inShuf1<>+0x28(SB)/8, $0x3736353433323130
DATA  expandAVX512_sparse_inShuf1<>+0x30(SB)/8, $0x3f3e3d3c3b3a3938
DATA  expandAVX512_sparse_inShuf1<>+0x38(SB)/8, $0x3f3e3d3c3b3a3938

GLOBL expandAVX512_sparse_outShufLo(SB), RODATA, $0x40
DATA  expandAVX512_sparse_outShufLo+0x00(SB)/8, $0x0b030a0209010800
DATA  expandAVX512_sparse_outShufLo+0x08(SB)/8, $0x0f070e060d050c04
DATA  expandAVX512_sparse_outShufLo+0x10(SB)/8, $0x1b131a1219111810
DATA  expandAVX512_sparse_outShufLo+0x18(SB)/8, $0x1f171e161d151c14
DATA  expandAVX512_sparse_outShufLo+0x20(SB)/8, $0x2b232a2229212820
DATA  expandAVX512_sparse_outShufLo+0x28(SB)/8, $0x2f272e262d252c24
DATA  expandAVX512_sparse_outShufLo+0x30(SB)/8, $0x3b333a3239313830
DATA  expandAVX512_sparse_outShufLo+0x38(SB)/8, $0x3f373e363d353c34

TEXT expandSparseAVX512(SB), NOSPLIT, $0-0
  VPXORQ Z7, Z7, Z7
  VPUNPCKLDQ Z7, Z0, Z1
  VPUNPCKHDQ Z7, Z0, Z2

  VPSLLQ X6, Z1, Z3
  VPSLLQ X6, Z2, Z4
  VPSUBQ Z1, Z3, Z1
  VPSUBQ Z2, Z4, Z2
  
  VPUNPCKLQDQ Z2, Z1, Z3
  VPUNPCKHQDQ Z2, Z1, Z4

  MOVW $0xFFFE, CX
	KMOVW CX, K1
  VPEXPANDD Z4, K1, Z7

  VPORQ Z3, Z7, Z4

	VMOVDQU64 expandAVX512_sparse_inShuf0<>(SB), Z0
	VMOVDQU64 expandAVX512_sparse_mat0<>(SB), Z1
	VMOVDQU64 expandAVX512_sparse_inShuf1<>(SB), Z2
	VMOVDQU64 expandAVX512_sparse_outShufLo(SB), Z3
	VPERMB Z4, Z0, Z0
	VGF2P8AFFINEQB $0, Z1, Z0, Z0
	VPERMB Z4, Z2, Z2
	VGF2P8AFFINEQB $0, Z1, Z2, Z2
	VPERMB Z0, Z3, Z1
	VPERMB Z2, Z3, Z2
	RET

// Test-only.
TEXT ·ExpandSparseAVX512(SB), NOSPLIT, $0-24
  VMOVQ elemsize+0(FP), X6
  MOVQ packed+8(FP), AX
	VMOVDQU64 (AX), Z0

  CALL expandSparseAVX512(SB)

	MOVQ unpacked+16(FP), DI // Expanded output bitmap pointer
	VMOVDQU64 Z1, 0(DI)
	VMOVDQU64 Z2, 64(DI)
	VZEROUPPER
	RET

// Test-only.
TEXT ·ExpandAVX512(SB), NOSPLIT, $0-24
	MOVQ sizeClass+0(FP), CX
	MOVQ packed+8(FP), AX

	// Call the expander for this size class
	LEAQ ·gcExpandersAVX512(SB), BX
	CALL (BX)(CX*8)

	MOVQ unpacked+16(FP), DI // Expanded output bitmap pointer
	VMOVDQU64 Z1, 0(DI)
	VMOVDQU64 Z2, 64(DI)
	VZEROUPPER
	RET

TEXT ·scanSpanPackedAVX512(SB), NOSPLIT, $256-44
	// Z1+Z2 = Expand the grey object mask into a grey word mask
	MOVQ objMarks+16(FP), AX
	MOVQ sizeClass+24(FP), CX
	LEAQ ·gcExpandersAVX512(SB), BX
	CALL (BX)(CX*8)

	// Z3+Z4 = Load the pointer mask
	MOVQ ptrMask+32(FP), AX
	VMOVDQU64 0(AX), Z3
	VMOVDQU64 64(AX), Z4

	// Z1+Z2 = Combine the grey word mask with the pointer mask to get the scan mask
	VPANDQ Z1, Z3, Z1
	VPANDQ Z2, Z4, Z2

	// Now each bit of Z1+Z2 represents one word of the span.
	// Thus, each byte covers 64 bytes of memory, which is also how
	// much we can fix in a Z register.
	//
	// We do a load/compress for each 64 byte frame.
	//
	// Z3+Z4 [128]uint8 = Number of memory words to scan in each 64 byte frame
	VPOPCNTB Z1, Z3 // Requires BITALG
	VPOPCNTB Z2, Z4

	// Store the scan mask and word counts at 0(SP) and 128(SP).
	//
	// TODO: Is it better to read directly from the registers?
	VMOVDQU64 Z1, 0(SP)
	VMOVDQU64 Z2, 64(SP)
	VMOVDQU64 Z3, 128(SP)
	VMOVDQU64 Z4, 192(SP)

	// SI = Current address in span
	MOVQ mem+0(FP), SI
	// DI = Scan buffer base
	MOVQ bufp+8(FP), DI
	// DX = Index in scan buffer, (DI)(DX*8) = Current position in scan buffer
	MOVQ $0, DX

	// AX = address in scan mask, 128(AX) = address in popcount
	LEAQ 0(SP), AX

	// Loop over the 64 byte frames in this span.
	// BX = 1 past the end of the scan mask
	LEAQ 128(SP), BX

	// Align loop to a cache line so that performance is less sensitive
	// to how this function ends up laid out in memory. This is a hot
	// function in the GC, and this is a tight loop. We don't want
	// performance to waver wildly due to unrelated changes.
	PCALIGN $64
loop:
	// CX = Fetch the mask of words to load from this frame.
	MOVBQZX 0(AX), CX
	// Skip empty frames.
	TESTQ CX, CX
	JZ skip

	// Load the 64 byte frame.
	KMOVB CX, K1
	VMOVDQA64 0(SI), Z1

	// Collect just the pointers from the greyed objects into the scan buffer,
	// i.e., copy the word indices in the mask from Z1 into contiguous memory.
	//
	// N.B. VPCOMPRESSQ supports a memory destination. Unfortunately, on
	// AMD Genoa / Zen 4, using VPCOMPRESSQ with a memory destination
	// imposes a severe performance penalty of around an order of magnitude
	// compared to a register destination.
	//
	// This workaround is unfortunate on other microarchitectures, where a
	// memory destination is slightly faster than adding an additional move
	// instruction, but no where near an order of magnitude. It would be
	// nice to have a Genoa-only variant here.
	//
	// AMD Turin / Zen 5 fixes this issue.
	//
	// See
	// https://lemire.me/blog/2025/02/14/avx-512-gotcha-avoid-compressing-words-to-memory-with-amd-zen-4-processors/.
	VPCOMPRESSQ Z1, K1, Z2
	VMOVDQU64 Z2, (DI)(DX*8)

	// Advance the scan buffer position by the number of pointers.
	MOVBQZX 128(AX), CX
	ADDQ CX, DX

skip:
	ADDQ $64, SI
	ADDQ $1, AX
	CMPQ AX, BX
	JB loop

end:
	MOVL DX, count+40(FP)
	VZEROUPPER
	RET

TEXT ·scanSpanPackedSparseAVX512(SB), NOSPLIT, $256-44
	// Z1+Z2 = Expand the grey object mask into a grey word mask
  VMOVQ elemsize+24(FP), X6
	MOVQ objMarks+16(FP), AX
	VMOVDQU64 (AX), Z0
  CALL expandSparseAVX512(SB)

	// Z3+Z4 = Load the pointer mask
	MOVQ ptrMask+32(FP), AX
	VMOVDQU64 0(AX), Z3
	VMOVDQU64 64(AX), Z4

	// Z1+Z2 = Combine the grey word mask with the pointer mask to get the scan mask
	VPANDQ Z1, Z3, Z1
	VPANDQ Z2, Z4, Z2

	// Now each bit of Z1+Z2 represents one word of the span.
	// Thus, each byte covers 64 bytes of memory, which is also how
	// much we can fix in a Z register.
	//
	// We do a load/compress for each 64 byte frame.
	//
	// Z3+Z4 [128]uint8 = Number of memory words to scan in each 64 byte frame
	VPOPCNTB Z1, Z3 // Requires BITALG
	VPOPCNTB Z2, Z4

	// Store the scan mask and word counts at 0(SP) and 128(SP).
	//
	// TODO: Is it better to read directly from the registers?
	VMOVDQU64 Z1, 0(SP)
	VMOVDQU64 Z2, 64(SP)
	VMOVDQU64 Z3, 128(SP)
	VMOVDQU64 Z4, 192(SP)

	// SI = Current address in span
	MOVQ mem+0(FP), SI
	// DI = Scan buffer base
	MOVQ bufp+8(FP), DI
	// DX = Index in scan buffer, (DI)(DX*8) = Current position in scan buffer
	MOVQ $0, DX

	// AX = address in scan mask, 128(AX) = address in popcount
	LEAQ 0(SP), AX

	// Loop over the 64 byte frames in this span.
	// BX = 1 past the end of the scan mask
	LEAQ 128(SP), BX

	// Align loop to a cache line so that performance is less sensitive
	// to how this function ends up laid out in memory. This is a hot
	// function in the GC, and this is a tight loop. We don't want
	// performance to waver wildly due to unrelated changes.
	PCALIGN $64
loop:
	// CX = Fetch the mask of words to load from this frame.
	MOVBQZX 0(AX), CX
	// Skip empty frames.
	TESTQ CX, CX
	JZ skip

	// Load the 64 byte frame.
	KMOVB CX, K1
	VMOVDQA64 0(SI), Z1

	// Collect just the pointers from the greyed objects into the scan buffer,
	// i.e., copy the word indices in the mask from Z1 into contiguous memory.
	//
	// N.B. VPCOMPRESSQ supports a memory destination. Unfortunately, on
	// AMD Genoa / Zen 4, using VPCOMPRESSQ with a memory destination
	// imposes a severe performance penalty of around an order of magnitude
	// compared to a register destination.
	//
	// This workaround is unfortunate on other microarchitectures, where a
	// memory destination is slightly faster than adding an additional move
	// instruction, but no where near an order of magnitude. It would be
	// nice to have a Genoa-only variant here.
	//
	// AMD Turin / Zen 5 fixes this issue.
	//
	// See
	// https://lemire.me/blog/2025/02/14/avx-512-gotcha-avoid-compressing-words-to-memory-with-amd-zen-4-processors/.
	VPCOMPRESSQ Z1, K1, Z2
	VMOVDQU64 Z2, (DI)(DX*8)

	// Advance the scan buffer position by the number of pointers.
	MOVBQZX 128(AX), CX
	ADDQ CX, DX

skip:
	ADDQ $64, SI
	ADDQ $1, AX
	CMPQ AX, BX
	JB loop

end:
	MOVL DX, count+40(FP)
	VZEROUPPER
	RET
