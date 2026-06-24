// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scan_test

import (
	"fmt"
	"internal/cpu"
	"internal/goarch"
	"internal/runtime/gc"
	"internal/runtime/gc/scan"
	"internal/runtime/sys"
	"math/bits"
	"math/rand/v2"
	"slices"
	"sync"
	"testing"
	"unsafe"
)

type scanSpanFunc func(mem unsafe.Pointer, bufp *uintptr, objMarks *gc.ObjMask, sizeClass uintptr, ptrMask *gc.PtrMask) (count int32)
type scanObjLFunc func(mem unsafe.Pointer, bufp *uintptr, ptrsize uintptr, ptrmap *uintptr, elemdiff uintptr, limit uintptr) (count uintptr)

func dumpSliceCmp(t *testing.T, name string, buf1, buf2 []uintptr) {
	t.Logf("%s:", name)
	for i := range max(len(buf1), len(buf2)) {
		if i < len(buf1) && i < len(buf2) {
			t.Logf("  [%03d]: %08x %08x", i, buf1[i], buf2[i])
		} else {
			if i >= len(buf1) {
				t.Logf("  [%03d]: -------- %08x", i, buf2[i])
			} else {
				t.Logf("  [%03d]: %08x ---------", i, buf1[i])
			}
		}
	}
}

func testScanSpanPacked(t *testing.T, scanF scanSpanFunc) {
	scanR := scan.ScanSpanPackedReference

	// Construct a fake memory
	mem, free := makeMem(t, 1)
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = uintptr(int(gc.PageSize) + i + 1)
	}

	// Construct a random pointer mask
	rnd := rand.New(rand.NewPCG(42, 42))
	var ptrs gc.PtrMask
	for i := range ptrs {
		ptrs[i] = uintptr(rnd.Uint64())
	}

	bufF := make([]uintptr, gc.PageWords)
	bufR := make([]uintptr, gc.PageWords)
	testSmallObjs(t, func(t *testing.T, sizeClass int, objs *gc.ObjMask) {
		nF := scanF(unsafe.Pointer(&mem[0]), &bufF[0], objs, uintptr(sizeClass), &ptrs)
		nR := scanR(unsafe.Pointer(&mem[0]), &bufR[0], objs, uintptr(sizeClass), &ptrs)

		if nR != nF {
			t.Errorf("want %d count, got %d", nR, nF)
		} else if !slices.Equal(bufF[:nF], bufR[:nR]) {
			t.Errorf("want scanned pointers %d, got %d", bufR[:nR], bufF[:nF])
		}
	})
}

func testScanObjectLarge(t *testing.T, scanF scanObjLFunc) {
	scanR := scan.ScanObjectLargeReference

	// Construct a fake memory
	mem, free := makeMem(t, gc.MaxSmallSize/gc.PageSize)
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = uintptr(int(gc.PageSize) + i + 1)
	}

	bufF := make([]uintptr, gc.PageWords)
	bufR := make([]uintptr, gc.PageWords)
	testLargeObjs(t, func(t *testing.T, sizeClass int, elemsize uintptr) {
		// Construct a random pointer mask
		rnd := rand.New(rand.NewPCG(42, 42))
		var ptrs gc.PtrMask
		ptrsize := uintptr(0)
		ptrcnt := 0
		nelems := uintptr(gc.SizeClassToSize[sizeClass]) / elemsize
		for i := range min((elemsize+goarch.PtrBits-1)/goarch.PtrBits, uintptr(len(ptrs))) {
			ptrs[i] = uintptr(rnd.Uint64())
			ptrcnt += sys.OnesCount64(uint64(ptrs[i]))
			if nelems*uintptr(ptrcnt) > gc.PageWords {
				for nelems*uintptr(ptrcnt) > gc.PageWords {
					ptrs[i] &= ptrs[i] - 1
					ptrcnt--
				}
				ptrsize = ((64 - uintptr(sys.LeadingZeros64(uint64(ptrs[i])))) + uintptr(i*64)) * goarch.PtrSize
				break
			}
			ptrsize = ((64 - uintptr(sys.LeadingZeros64(uint64(ptrs[i])))) + uintptr(i*64)) * goarch.PtrSize
		}
		if elemsize%goarch.PtrBits != 0 && elemsize/goarch.PtrBits < uintptr(len(ptrs)) {
			ptrs[elemsize/goarch.PtrBits] &= (uintptr(1) << (elemsize % goarch.PtrBits)) - 1
			ptrsize = min(ptrsize, elemsize)
		}

		ptrsizeGranular := (ptrsize + scan.ScanLargeGranularity - 1) &^ (scan.ScanLargeGranularity - 1)
		delta := elemsize - ptrsizeGranular
		limit := uintptr(unsafe.Pointer(&mem[0])) + nelems*elemsize

		nF := scanF(unsafe.Pointer(&mem[0]), &bufF[0], ptrsizeGranular, &ptrs[0], delta, limit)
		nR := scanR(unsafe.Pointer(&mem[0]), &bufR[0], ptrsizeGranular, &ptrs[0], delta, limit)

		if nR != nF {
			dumpSliceCmp(t, "scanBuf want got", bufR[:nR], bufF[:nF])
			t.Errorf("want %d count, got %d", nR, nF)
		} else if !slices.Equal(bufF[:nF], bufR[:nR]) {
			dumpSliceCmp(t, "scanBuf want got", bufR[:nR], bufF[:nF])
			t.Errorf("want scanned pointers %d, got %d", bufR[:nR], bufF[:nF])
		}
	})
}

func testLargeObjs(t *testing.T, f func(t *testing.T, sizeClass int, elemsize uintptr)) {
	for sizeClass := range gc.NumSizeClasses {
		if sizeClass == 0 {
			continue
		}
		size := uintptr(gc.SizeClassToSize[sizeClass])
		if size < 1024 {
			continue
		}
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			for elemsize := uintptr(goarch.PtrBits*goarch.PtrSize) / 2; elemsize <= size; elemsize += 8 {
				t.Run(fmt.Sprintf("elemsize=%d", elemsize), func(t *testing.T) {
					f(t, sizeClass, elemsize)
				})
			}
		})
	}
}

func testSmallObjs(t *testing.T, f func(t *testing.T, sizeClass int, objMask *gc.ObjMask)) {
	for sizeClass := range gc.NumSizeClasses {
		if sizeClass == 0 {
			continue
		}
		size := uintptr(gc.SizeClassToSize[sizeClass])
		if size > gc.MinSizeForMallocHeader {
			break // Pointer/scalar metadata is not packed for larger sizes.
		}
		t.Run(fmt.Sprintf("size=%d", size), func(t *testing.T) {
			// Scan a few objects near i to test boundary conditions.
			const objMask = 0x101
			nObj := uintptr(gc.SizeClassToNPages[sizeClass]) * gc.PageSize / size
			for i := range nObj - uintptr(bits.Len(objMask)-1) {
				t.Run(fmt.Sprintf("objs=0x%x<<%d", objMask, i), func(t *testing.T) {
					var objs gc.ObjMask
					objs[i/goarch.PtrBits] = objMask << (i % goarch.PtrBits)
					f(t, sizeClass, &objs)
				})
			}
		})
	}
}

var dataCacheSizes = sync.OnceValue(func() []uintptr {
	cs := cpu.DataCacheSizes()
	for i, c := range cs {
		fmt.Printf("# L%d cache: %d (%d Go pages)\n", i+1, c, c/gc.PageSize)
	}
	return cs
})

func BenchmarkScanObjectLarge(b *testing.B) {
	benchmarkScanObjectLargeAllSizeClasses(b)
}

func BenchmarkScanSpanPacked(b *testing.B) {
	benchmarkCacheSizes(b, benchmarkScanSpanPackedAllSizeClasses)
}

func benchmarkCacheSizes(b *testing.B, fn func(b *testing.B, heapPages int)) {
	cacheSizes := dataCacheSizes()
	b.Run("cache=tiny/pages=1", func(b *testing.B) {
		fn(b, 1)
	})
	for i, cacheBytes := range cacheSizes {
		pages := int(cacheBytes*3/4) / gc.PageSize
		b.Run(fmt.Sprintf("cache=L%d/pages=%d", i+1, pages), func(b *testing.B) {
			fn(b, pages)
		})
	}
	if len(cacheSizes) == 0 {
		return
	}
	ramPages := int(cacheSizes[len(cacheSizes)-1]*3/2) / gc.PageSize
	b.Run(fmt.Sprintf("cache=ram/pages=%d", ramPages), func(b *testing.B) {
		fn(b, ramPages)
	})
}

func benchmarkScanObjectLargeAllSizeClasses(b *testing.B) {
	for sc := range gc.NumSizeClasses {
		if sc == 0 {
			continue
		}
		size := gc.SizeClassToSize[sc]
		if size <= gc.MinSizeForMallocHeader {
			continue
		}
		b.Run(fmt.Sprintf("sizeclass=%d", sc), func(b *testing.B) {
			benchmarkScanLargeObject(b, sc)
		})
	}
}

func benchmarkScanSpanPackedAllSizeClasses(b *testing.B, nPages int) {
	for sc := range gc.NumSizeClasses {
		if sc == 0 {
			continue
		}
		size := gc.SizeClassToSize[sc]
		if size >= gc.MinSizeForMallocHeader {
			break
		}
		b.Run(fmt.Sprintf("sizeclass=%d", sc), func(b *testing.B) {
			benchmarkScanSpanPacked(b, nPages, sc)
		})
	}
}

type typInfo struct {
	size     uintptr
	ptrbytes uintptr
	gcdata   [gc.MaxSmallSize / goarch.PtrSize / goarch.PtrBits]uintptr
}

func makeObjectInfo(sizeClass int, n int) (infos []typInfo) {
	rnd := rand.New(rand.NewPCG(42, 42))
	infos = make([]typInfo, n)

	// Threre is gcdata optimization that expands gcdata if it is too small.
	// So limit element size from below.
	const minElemWords = gc.PageSize / goarch.PtrSize / 2

	// Scan buffer is limited to PageWords.
	const maxPtrCount = gc.PageWords
	for i := range n {
		if gc.SizeClassToSize[sizeClass] < gc.PageSize/2 {
			infos[i].size = uintptr(gc.SizeClassToSize[sizeClass])
		} else {
			infos[i].size = goarch.PtrSize * (minElemWords + uintptr(rnd.Uint64())%(uintptr(gc.SizeClassToSize[sizeClass])/goarch.PtrSize-minElemWords))
		}
		infos[i].ptrbytes = goarch.PtrSize * (8 + uintptr(rnd.Uint64())%uintptr((infos[i].size-8)/goarch.PtrSize))
		gcdataWords := (infos[i].ptrbytes + goarch.PtrBits - 1) / goarch.PtrBits
		ptrCount := 0
		for j := range gcdataWords {
			infos[i].gcdata[j] = uintptr(rnd.Uint64())
			ptrCount += sys.OnesCount64(uint64(infos[i].gcdata[j]))
			if ptrCount > maxPtrCount {
				for ptrCount > maxPtrCount {
					infos[i].gcdata[j] &= infos[i].gcdata[j] - 1
					ptrCount--
				}
				break
			}
		}
		if off := infos[i].ptrbytes % goarch.PtrBits; off != 0 {
			infos[i].gcdata[gcdataWords-1] &= (uintptr(1) << off) - 1
		}
	}
	return
}

func benchmarkScanSpanPacked(b *testing.B, nPages int, sizeClass int) {
	rnd := rand.New(rand.NewPCG(42, 42))

	// Construct a fake memory
	mem, free := makeMem(b, nPages)
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = uintptr(int(gc.PageSize) + i + 1)
	}

	// Construct a random pointer mask
	ptrs := make([]gc.PtrMask, nPages)
	for i := range ptrs {
		for j := range ptrs[i] {
			ptrs[i][j] = uintptr(rnd.Uint64())
		}
	}

	// Visit the pages in a random order
	pageOrder := rnd.Perm(nPages)

	// Create the scan buffer.
	buf := make([]uintptr, gc.PageWords)

	// Sweep from 0 marks to all marks. We'll use the same marks for each page
	// because I don't think that predictability matters.
	objBytes := uintptr(gc.SizeClassToSize[sizeClass])
	nObj := gc.PageSize / objBytes
	markOrder := rnd.Perm(int(nObj))
	const steps = 11
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)
		// Set frac marks.
		nMarks := int(float64(len(markOrder))*frac + 0.5)
		var objMarks gc.ObjMask
		for _, mark := range markOrder[:nMarks] {
			objMarks[mark/goarch.PtrBits] |= 1 << (mark % goarch.PtrBits)
		}
		greyClusters := 0
		for page := range ptrs {
			greyClusters += countGreyClusters(sizeClass, &objMarks, &ptrs[page])
		}

		// Report MB/s of how much memory they're actually hitting. This assumes
		// 64 byte cache lines (TODO: Should it assume 128 byte cache lines?)
		// and expands each access to the whole cache line. This is useful for
		// comparing against memory bandwidth.
		//
		// TODO: Add a benchmark that just measures single core memory bandwidth
		// for comparison. (See runtime memcpy benchmarks.)
		//
		// TODO: Should there be a separate measure where we don't expand to
		// cache lines?
		avgBytes := int64(greyClusters) * int64(cpu.CacheLineSize) / int64(len(ptrs))

		b.Run(fmt.Sprintf("pct=%d", int(100*frac)), func(b *testing.B) {
			b.Run("impl=Reference", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scan.ScanSpanPackedReference(unsafe.Pointer(&mem[gc.PageWords*page]), &buf[0], &objMarks, uintptr(sizeClass), &ptrs[page])
				}
			})
			b.Run("impl=Go", func(b *testing.B) {
				b.SetBytes(avgBytes)
				for i := range b.N {
					page := pageOrder[i%len(pageOrder)]
					scan.ScanSpanPackedGo(unsafe.Pointer(&mem[gc.PageWords*page]), &buf[0], &objMarks, uintptr(sizeClass), &ptrs[page])
				}
			})
			if scan.HasFastScanSpanPacked() {
				b.Run("impl=Platform", func(b *testing.B) {
					b.SetBytes(avgBytes)
					for i := range b.N {
						page := pageOrder[i%len(pageOrder)]
						scan.ScanSpanPacked(unsafe.Pointer(&mem[gc.PageWords*page]), &buf[0], &objMarks, uintptr(sizeClass), &ptrs[page])
					}
				})
			}
		})
	}
}

func benchmarkScanLargeObject(b *testing.B, sizeClass int) {
	rnd := rand.New(rand.NewPCG(42, 42))

	// Construct a fake memory
	mem, free := makeMem(b, int(gc.SizeClassToNPages[sizeClass]))
	defer free()
	for i := range mem {
		// Use values > heap.PageSize because a scan function can discard
		// pointers smaller than this.
		mem[i] = uintptr(int(gc.PageSize) + i + 1)
	}

	nObjects := int(gc.SizeClassToNPages[sizeClass]) * gc.PageSize / int(gc.SizeClassToSize[sizeClass])

	// Construct a random object info
	infos := makeObjectInfo(sizeClass, nObjects)

	// Visit the objects in a random order
	markOrder := rnd.Perm(nObjects)

	// Create the scan buffer.
	buf := make([]uintptr, gc.PageWords)

	const steps = 11
	for i := 0; i < steps; i++ {
		frac := float64(i) / float64(steps-1)

		b.Run(fmt.Sprintf("pct=%d", int(100*frac)), func(b *testing.B) {
			b.Run("impl=Reference", func(b *testing.B) {
				for i := range b.N {
					obj := markOrder[i%len(markOrder)]
					base := unsafe.Pointer(&mem[obj*int(gc.SizeClassToSize[sizeClass])/goarch.PtrSize])
					limit := uintptr(base) + uintptr(gc.SizeClassToSize[sizeClass])/infos[obj].size*infos[obj].size
					scan.ScanObjectLargeReference(base, &buf[0], infos[obj].ptrbytes, &infos[obj].gcdata[0], infos[obj].size-infos[obj].ptrbytes, limit)
				}
			})

			if scan.HasFastScanObjectLarge() {
				b.Run("impl=Platform", func(b *testing.B) {
					for i := range b.N {
						obj := markOrder[i%len(markOrder)]
						base := unsafe.Pointer(&mem[obj*int(gc.SizeClassToSize[sizeClass])/goarch.PtrSize])
						limit := uintptr(base) + uintptr(gc.SizeClassToSize[sizeClass])/infos[obj].size*infos[obj].size
						scan.ScanObjectLarge(base, &buf[0], infos[obj].ptrbytes, &infos[obj].gcdata[0], infos[obj].size-infos[obj].ptrbytes, limit)
					}
				})
			}
		})
	}
}

func countGreyClusters(sizeClass int, objMarks *gc.ObjMask, ptrMask *gc.PtrMask) int {
	clusters := 0
	lastCluster := -1

	expandBy := uintptr(gc.SizeClassToSize[sizeClass]) / goarch.PtrSize
	for word := range gc.PageWords {
		objI := uintptr(word) / expandBy
		if objMarks[objI/goarch.PtrBits]&(1<<(objI%goarch.PtrBits)) == 0 {
			continue
		}
		if ptrMask[word/goarch.PtrBits]&(1<<(word%goarch.PtrBits)) == 0 {
			continue
		}
		c := word * 8 / goarch.PtrBits
		if c != lastCluster {
			lastCluster = c
			clusters++
		}
	}
	return clusters
}

func BenchmarkScanMaxBandwidth(b *testing.B) {
	// Measure the theoretical "maximum" bandwidth of scanning by reproducing
	// the memory access pattern of a full page scan, but using memcpy as the
	// kernel instead of scanning.
	benchmarkCacheSizes(b, func(b *testing.B, heapPages int) {
		mem, free := makeMem(b, heapPages)
		defer free()
		for i := range mem {
			mem[i] = uintptr(int(gc.PageSize) + i + 1)
		}
		buf := make([]uintptr, gc.PageWords)

		// Visit the pages in a random order
		rnd := rand.New(rand.NewPCG(42, 42))
		pageOrder := rnd.Perm(heapPages)

		b.SetBytes(int64(gc.PageSize))

		b.ResetTimer()
		for i := range b.N {
			page := pageOrder[i%len(pageOrder)]
			copy(buf, mem[gc.PageWords*page:])
		}
	})
}
