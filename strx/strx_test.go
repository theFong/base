// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package strx

import (
	"testing"
)

func TestOr(t *testing.T) {
	if v := Or(); v != "" {
		t.Errorf("got %q", v)
	}
	if v := Or("", "backup"); v != "backup" {
		t.Errorf("got %q", v)
	}
	if v := Or("primary", ""); v != "primary" {
		t.Errorf("got %q", v)
	}
	if v := Or("primary", "backup"); v != "primary" {
		t.Errorf("got %q", v)
	}
}

func TestYes(t *testing.T) {
	if !Yes("yes") {
		t.Error("expected true")
	}
	if Yes("no") {
		t.Error("expected no")
	}

	if !Yes("  true") {
		t.Error("expected true")
	}
	if Yes("false") {
		t.Error("expected no")
	}

	if Yes("") {
		t.Error("expected no")
	}
}
