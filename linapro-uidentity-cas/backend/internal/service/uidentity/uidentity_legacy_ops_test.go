// This file tests dependency-light legacy operation helpers.

package uidentity

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitLegacyBase64Payload(t *testing.T) {
	mimeType, payload, err := splitLegacyBase64Payload("data:image/png;base64,QUJD")
	if err != nil {
		t.Fatalf("splitLegacyBase64Payload returned error: %v", err)
	}
	if mimeType != "image/png" || payload != "QUJD" {
		t.Fatalf("unexpected split result: %q %q", mimeType, payload)
	}
}

func TestLegacySafeFilename(t *testing.T) {
	got := legacySafeFilename("../bad:name.png")
	if got != "bad_name.png" {
		t.Fatalf("unexpected safe filename: %s", got)
	}
}

func TestReadTailLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "today.log")
	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatalf("write temp log: %v", err)
	}
	lines, exists, truncated, err := readTailLines(context.Background(), path, 2)
	if err != nil {
		t.Fatalf("readTailLines returned error: %v", err)
	}
	if !exists || !truncated {
		t.Fatalf("expected existing truncated log, exists=%v truncated=%v", exists, truncated)
	}
	if len(lines) != 2 || lines[0] != "b" || lines[1] != "c" {
		t.Fatalf("unexpected tail lines: %#v", lines)
	}
}
