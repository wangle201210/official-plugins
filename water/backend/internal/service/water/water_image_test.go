// This file tests watermark image helpers.

package water

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
)

// TestDecodeImageDataURLRejectsInvalidBase64 verifies invalid input is rejected.
func TestDecodeImageDataURLRejectsInvalidBase64(t *testing.T) {
	if _, err := decodeImageDataURL("data:image/png;base64,%%%"); err == nil {
		t.Fatal("expected invalid base64 error")
	}
}

// TestEnsurePNGDataURL verifies an input image is converted to PNG data URL.
func TestEnsurePNGDataURL(t *testing.T) {
	dataURL, err := ensurePNGDataURL(testPNGBytes(t))
	if err != nil {
		t.Fatalf("ensure png data url failed: %v", err)
	}
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		t.Fatalf("expected png data url prefix, got %q", dataURL)
	}
	if _, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(dataURL, prefix)); err != nil {
		t.Fatalf("expected valid base64 output: %v", err)
	}
}

// TestBase64ToMD5PicStoresWatermarkImage verifies HotGo-compatible base64
// watermark images are materialized as files for FFmpeg's movie filter.
func TestBase64ToMD5PicStoresWatermarkImage(t *testing.T) {
	path, err := base64ToMD5Pic(encodePNGDataURL(testPNGBytes(t)), t.TempDir())
	if err != nil {
		t.Fatalf("store base64 watermark image failed: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read stored watermark image failed: %v", err)
	}
	if _, err := png.Decode(bytes.NewReader(content)); err != nil {
		t.Fatalf("expected stored watermark image to decode: %v", err)
	}
}

// testPNGBytes creates a deterministic PNG image.
func testPNGBytes(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 120, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 120; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 30, G: 80, B: 160, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test png failed: %v", err)
	}
	return buf.Bytes()
}
