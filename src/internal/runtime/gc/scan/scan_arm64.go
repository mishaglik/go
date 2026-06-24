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
	panic("not implemented")
}

func HasFastScanSpanPacked() bool {
	return false
}

const ScanLargeGranularity uintptr = 256

func HasFastScanObjectLarge() bool {
	return cpu.ARM64.HasSVE
}

func ScanObjectLarge(b unsafe.Pointer, dst *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) uintptr {
	if CanSVE() {
		return scanObjectLargeSVE(b, dst, ptrsize, ptrmap, elemdiff, limit)
	}
	panic("not implemented")
}

// -- SVE --

func CanSVE() bool {
	return cpu.ARM64.HasSVE
}

//go:noescape
func scanObjectLargeSVE(mem unsafe.Pointer, bufp *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) (count uintptr)
