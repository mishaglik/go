// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan

import (
	"internal/goarch"
	"internal/runtime/gc"
	"unsafe"
)

// ScanSpanPackedReference is the reference implementation of ScanScanPacked. It prioritizes clarity over performance.
//
// Concretely, ScanScanPacked functions read pointers from mem, assumed to be gc.PageSize-aligned and gc.PageSize in size,
// and writes them to bufp, which is large enough to guarantee that even if pointer-word of mem is a pointer, it will fit.
// Therefore bufp, is always at least gc.PageSize in size.
//
// ScanSpanPacked is supposed to identify pointers by first filtering words by objMarks, where each bit of the mask
// represents gc.SizeClassToSize[sizeClass] bytes of memory, and then filtering again by the bits in ptrMask.
func ScanSpanPackedReference(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32) {
	buf := unsafe.Slice(bufp, gc.PageWords)
	expandBy := uintptr(gc.SizeClassToSize[sizeClass]) / goarch.PtrSize
	for word := range gc.PageWords {
		objI := uintptr(word) / expandBy
		if objMarks[objI/goarch.PtrBits]&(1<<(objI%goarch.PtrBits)) == 0 {
			continue
		}
		if ptrMask[word/goarch.PtrBits]&(1<<(word%goarch.PtrBits)) == 0 {
			continue
		}
		ptr := *(*uintptr)(unsafe.Add(mem, word*goarch.PtrSize))
		if ptr == 0 {
			continue
		}
		buf[count] = ptr
		count++
	}
	return count
}

// ScanObjectLargeReference is the reference implementation of ScanObjectLarge. It prioritizes clarity over performance.
func ScanObjectLargeReference(mem unsafe.Pointer, bufp *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) (count uintptr) {
	buf := unsafe.Slice(bufp, gc.PageWords)
	ptrMask := unsafe.Slice(ptrmap, ptrsize/goarch.PtrBits)
	// Iterate over array of elements
	for elem := 0; uintptr(unsafe.Add(mem, elem)) < limit; elem += int(ptrsize) + int(elemdiff) {
		// Iterate over words in single element
		for elemWord := 0; elemWord*goarch.PtrSize < int(ptrsize); elemWord++ {
			if ptrMask[elemWord/goarch.PtrBits]&(1<<(elemWord%goarch.PtrBits)) == 0 {
				continue
			}
			ptr := *(*uintptr)(unsafe.Add(mem, elem+elemWord*goarch.PtrSize))
			if ptr == 0 {
				continue
			}
			buf[count] = ptr
			count++
		}
	}
	return count
}
