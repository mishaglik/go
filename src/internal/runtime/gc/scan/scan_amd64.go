// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build amd64

package scan

import (
	"internal/cpu"
	"internal/runtime/gc"
	"unsafe"
)

func ScanSpanPacked(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32) {
	if CanAVX512() {
		return ScanSpanPackedAVX512(mem, bufp, objMarks, sizeClass, ptrMask)
	}
	panic("not implemented")
}

func HasFastScanSpanPacked() bool {
	return avx512ScanPackedReqsMet
}

const ScanLargeGranularity uintptr = 64

func HasFastScanObjectLarge() bool {
	return avx512ScanPackedReqsMet
}

func ScanObjectLarge(b unsafe.Pointer, dst *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) uintptr {
	if CanAVX512() {
		return scanObjectLargeAVX512(b, dst, ptrsize, ptrmap, elemdiff, limit)
	}
	panic("not implemented")
}

// -- AVX512 --

func CanAVX512() bool {
	return avx512ScanPackedReqsMet
}

func ScanSpanPackedAVX512(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32) {
	return FilterNilAVX512(bufp, scanSpanPackedAVX512(mem, bufp, objMarks, sizeClass, ptrMask))
}

//go:noescape
func scanObjectLargeAVX512(mem unsafe.Pointer, bufp *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) (count uintptr)

//go:noescape
func scanSpanPackedAVX512(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32)

var avx512ScanPackedReqsMet = cpu.X86.HasAVX512VL &&
	cpu.X86.HasAVX512BW &&
	cpu.X86.HasGFNI &&
	cpu.X86.HasAVX512BITALG &&
	cpu.X86.HasAVX512DQ && // for kmovb, see #79871
	cpu.X86.HasAVX512VBMI &&
	cpu.X86.HasPOPCNT
