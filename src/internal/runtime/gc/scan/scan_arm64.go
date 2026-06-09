// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64

package scan

import (
	"internal/cpu"
	"internal/runtime/gc"
	"unsafe"
)

func ScanSpanPacked(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32) {
	if CanSVE() {
		return ScanSpanPackedSVE(mem, bufp, objMarks, sizeClass, ptrMask)
	}
	panic("not implemented")
}

func HasFastScanSpanPacked() bool {
	return sveScanPackedReqsMet
}

// -- SVE --

func CanSVE() bool {
	return sveScanPackedReqsMet
}

func ScanSpanPackedSVE(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32) {
		return scanSpanPackedSparseSVE(mem, bufp, objMarks, uintptr(gc.SizeClassToSize[sizeClass]) / gc.MarkBitsSparseDistance, ptrMask)
}

//go:noescape
func scanSpanPackedSparseSVE(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, elemsize uintptr, ptrMask *gc.PtrMask) (count int32)

var sveScanPackedReqsMet = gc.MarkBitsAreSparse &&
	cpu.ARM64.HasSVE
