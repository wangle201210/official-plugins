//go:build !cgo

package watermark

import (
	"context"
	"strings"
	"testing"
)

// TestDrawWatermarkJpegWithoutCGOReportsUnavailable verifies non-CGO builds keep the public adapter symbol.
func TestDrawWatermarkJpegWithoutCGOReportsUnavailable(t *testing.T) {
	_, err := DrawWatermarkJpeg(context.Background(), []byte("jpeg"), WatermarkConfig{})
	if err == nil {
		t.Fatal("expected non-CGO watermark adapter to report unavailable")
	}
	if !strings.Contains(err.Error(), "CGO is disabled") {
		t.Fatalf("expected CGO disabled error, got %v", err)
	}
}
