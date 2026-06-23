// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !goexperiment.greenteagc

package runtime

import (
	"internal/goarch"
	"internal/runtime/atomic"
	"internal/runtime/gc"
	"internal/runtime/sys"
	"unsafe"
)

func (s *mspan) markBitsForIndex(objIndex uintptr) markBits {
	bytep, mask := s.gcmarkBits.bitp(objIndex)
	return markBits{bytep, mask, objIndex}
}

func (s *mspan) markBitsForBase() markBits {
	return markBits{&s.gcmarkBits.x, uint8(1), 0}
}

func (s *mspan) initInlineMarkBits() {
}

func (s *mspan) moveInlineMarks(to *gcBits) {
	throw("unimplemented")
}

func gcUsesSpanInlineMarkBits(_ uintptr) bool {
	return false
}

func (s *mspan) inlineMarkBits() *spanInlineMarkBits {
	return nil
}

func (s *mspan) scannedBitsForIndex(objIndex uintptr) markBits {
	throw("unimplemented")
	return markBits{}
}

type spanInlineMarkBits struct{}

func (q *spanInlineMarkBits) tryAcquire() bool {
	return false
}

type spanQueue struct{}

func (q *spanQueue) flush() {
}

func (q *spanQueue) empty() bool {
	return true
}

func (q *spanQueue) destroy() {
}

type spanSPMC struct {
	_       sys.NotInHeap
	allnode listNodeManual
}

func freeDeadSpanSPMCs() {
	return
}

type objptr uintptr

func (w *gcWork) tryGetSpanFast() objptr {
	return 0
}

func (w *gcWork) tryGetSpan() objptr {
	return 0
}

func (w *gcWork) tryStealSpan() objptr {
	return 0
}

func scanSpan(p objptr, gcw *gcWork) {
	throw("unimplemented")
}

type sizeClassScanStats struct {
	sparseObjsScanned uint64
}

func dumpScanStats() {
	var sparseObjsScanned uint64
	for _, stats := range memstats.lastScanStats {
		sparseObjsScanned += stats.sparseObjsScanned
	}
	print("scan: total ", sparseObjsScanned, " objs\n")
	for i, stats := range memstats.lastScanStats {
		if stats == (sizeClassScanStats{}) {
			continue
		}
		if i == 0 {
			print("scan: class L ")
		} else {
			print("scan: class ", gc.SizeClassToSize[i], "B ")
		}
		print(stats.sparseObjsScanned, " objs\n")
	}
}

func (w *gcWork) flushScanStats(dst *[gc.NumSizeClasses]sizeClassScanStats) {
	for i := range w.stats {
		dst[i].sparseObjsScanned += w.stats[i].sparseObjsScanned
	}
	clear(w.stats[:])
}

// gcMarkWorkAvailable reports whether there's any non-local work available to do.
func gcMarkWorkAvailable() bool {
	if !work.full.empty() {
		return true // global work available
	}
	if work.markrootNext.Load() < work.markrootJobs.Load() {
		return true // root scan work available
	}
	return false
}

// scanObject scans the object starting at b, adding pointers to gcw.
// b must point to the beginning of a heap object or an oblet.
// scanObject consults the GC bitmap for the pointer mask and the
// spans for the size of the object.
//
//go:nowritebarrier
func scanObject(b uintptr, gcw *gcWork) {
	// Prefetch object before we scan it.
	//
	// This will overlap fetching the beginning of the object with initial
	// setup before we start scanning the object.
	sys.Prefetch(b)

	// Find the bits for b and the size of the object at b.
	//
	// b is either the beginning of an object, in which case this
	// is the size of the object to scan, or it points to an
	// oblet, in which case we compute the size to scan below.
	s := spanOfUnchecked(b)
	n := s.elemsize
	if n == 0 {
		throw("scanObject n == 0")
	}
	if s.spanclass.noscan() {
		// Correctness-wise this is ok, but it's inefficient
		// if noscan objects reach here.
		throw("scanObject of a noscan object")
	}

	var tp typePointers
	if n > maxObletBytes {
		// Large object. Break into oblets for better
		// parallelism and lower latency.
		if b == s.base() {
			// Enqueue the other oblets to scan later.
			// Some oblets may be in b's scalar tail, but
			// these will be marked as "no more pointers",
			// so we'll drop out immediately when we go to
			// scan those.
			for oblet := b + maxObletBytes; oblet < s.base()+s.elemsize; oblet += maxObletBytes {
				if !gcw.putObjFast(oblet) {
					gcw.putObj(oblet)
				}
			}
		}

		// Compute the size of the oblet. Since this object
		// must be a large object, s.base() is the beginning
		// of the object.
		n = s.base() + s.elemsize - b
		n = min(n, maxObletBytes)
		tp = s.typePointersOfUnchecked(s.base())
		tp = tp.fastForward(b-tp.addr, b+n)
	} else {
		tp = s.typePointersOfUnchecked(b)
	}

	var scanSize uintptr
	for {
		var addr uintptr
		if tp, addr = tp.nextFast(); addr == 0 {
			if tp, addr = tp.next(b + n); addr == 0 {
				break
			}
		}

		// Keep track of farthest pointer we found, so we can
		// update heapScanWork. TODO: is there a better metric,
		// now that we can skip scalar portions pretty efficiently?
		scanSize = addr - b + goarch.PtrSize

		// Work here is duplicated in scanblock and above.
		// If you make changes here, make changes there too.
		obj := *(*uintptr)(unsafe.Pointer(addr))

		// At this point we have extracted the next potential pointer.
		// Quickly filter out nil and pointers back to the current object.
		if obj != 0 && obj-b >= n {
			// Test if obj points into the Go heap and, if so,
			// mark the object.
			//
			// Note that it's possible for findObject to
			// fail if obj points to a just-allocated heap
			// object because of a race with growing the
			// heap. In this case, we know the object was
			// just allocated and hence will be marked by
			// allocation itself.
			gcEnqueue(obj, b, addr, gcw)
		}
	}
	gcw.bytesMarked += uint64(n)
	gcw.heapScanWork += int64(scanSize)
	if debug.gctrace > 1 {
		gcw.stats[s.spanclass.sizeclass()].sparseObjsScanned++
	}
}

//go:nowritebarrier
func gcEnqueue(p uintptr, base, addr uintptr, gcw *gcWork) {
	if obj, span, objIndex := findObject(p, base, addr-base); obj != 0 {
		greyobject(obj, base, addr-base, span, gcw, objIndex)
	}
}

//go:nowritebarrier
func gcEnqueueBatch(objs []uintptr, base uintptr, gcw *gcWork) {
	pos := 0
	for _, p := range objs {
		if p < minLegalPointer {
			continue
		}
		// Quickly to see if this is a span that has inline mark bits.
		ha := heapArenaOf(p)
		if ha == nil {
			continue
		}
		pageIdx := ((p / pageSize) / 8) % uintptr(len(ha.pageInUse))
		pageMask := byte(1 << ((p / pageSize) % 8))

		span := ha.spans[(p/pageSize)%pagesPerArena]
		// If span is nil, the virtual address has never been part of the heap.
		// This pointer may be to some mmap'd region, so we allow it.
		if span == nil {
			if (GOARCH == "amd64" || GOARCH == "arm64") && p == clobberdeadPtr && debug.invalidptr != 0 {
				// Crash if clobberdeadPtr is seen. Only on AMD64 and ARM64 for now,
				// as they are the only platform where compiler's clobberdead mode is
				// implemented. On these platforms clobberdeadPtr cannot be a valid address.
				badPointer(span, p, base, 0)
			}
			continue
		}
		// If p is a bad pointer, it may not be in s's bounds.
		//
		// Check s.state to synchronize with span initialization
		// before checking other fields. See also spanOfHeap.
		if state := span.state.get(); state != mSpanInUse || p < span.base() || p >= span.limit {
			// Pointers into stacks are also ok, the runtime manages these explicitly.
			if state == mSpanManual {
				continue
			}
			// The following ensures that we are rigorous about what data
			// structures hold valid pointers.
			if debug.invalidptr != 0 {
				badPointer(span, p, base, 0)
			}
			continue
		}

		objIndex := span.objIndex(p)
		objBase := span.base() + objIndex*span.elemsize
		mbits := span.markBitsForIndex(objIndex)
		if useCheckmark {
			if setCheckmark(objBase, base, 0, mbits) {
				continue
			}
			if debug.checkfinalizers > 1 {
				print("  mark ", hex(objBase), " found at *(", hex(base), "+ ???)\n")
			}
		} else {
			if debug.gccheckmark > 0 && span.isFree(objIndex) {
				print("runtime: marking free object ", hex(objBase), " found at *(", hex(base), "+ ???\n")
				gcDumpObject("base", base, 0)
				gcDumpObject("obj", objBase, ^uintptr(0))
				getg().m.traceback = 2
				throw("marking free object")
			}
		}
		if mbits.isMarked() {
			continue
		}
		mbits.setMarked()

		if ha.pageMarks[pageIdx]&pageMask == 0 {
			atomic.Or8(&ha.pageMarks[pageIdx], pageMask)
		}

		if span.spanclass.noscan() {
			gcw.bytesMarked += uint64(span.elemsize)
			continue
		}
		objs[pos] = objBase
		pos++
	}
	if pos > 0 {
		// Enqueue the greyed objects.
		gcw.putObjBatch(objs[:pos])
	}
}

//go:nowritebarrier
func gcEnqueueBatchIndirect(slots []uintptr, base uintptr, gcw *gcWork) {
	for _, slot := range slots {
		obj := *(*uintptr)(unsafe.Pointer(slot))
		if obj < minLegalPointer {
			continue
		}
		gcEnqueue(obj, base, slot, gcw)
	}
}
