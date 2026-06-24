// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build arm64

package scan_test

import (
	"internal/runtime/gc/scan"
	"testing"
)

func TestScanObjectLargeSVE(t *testing.T) {
	if !scan.CanSVE() {
		t.Skip("no SVE")
	}
	testScanObjectLarge(t, scan.ScanObjectLarge)
}
