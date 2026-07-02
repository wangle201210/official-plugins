//go:build cgo

// This file verifies LinaPro's cgo adapter around the migrated HotGo watermark library.

package watermark

import (
	"image"
	"os"
	"runtime"
	"testing"
)

// TestMigratedHotGoWatermarkLibraryProducesJpeg verifies the migrated HotGo
// cgo/FFmpeg library is callable from the LinaPro water plugin.
func TestMigratedHotGoWatermarkLibraryProducesJpeg(t *testing.T) {
	input, err := os.ReadFile("input.jpg")
	if err != nil {
		t.Fatalf("read migrated hotgo input fixture: %v", err)
	}

	output, err := DrawWatermarkJpeg(nil, input, WatermarkConfig{
		Order:    "LinaPro 水印测试",
		FontSize: 32,
		Color:    "white",
		Opacity:  0.15,
	})
	if err != nil {
		t.Fatalf("draw watermark with migrated hotgo library: %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected non-empty watermark output")
	}
	if string(output[:2]) != "\xff\xd8" {
		t.Fatalf("expected JPEG output from migrated library, got header %x", output[:2])
	}
}

// TestWatermarkOutputBufferSizeHandlesTinyInputs verifies small compressed
// images still receive enough output space after watermark filters expand them.
func TestWatermarkOutputBufferSizeHandlesTinyInputs(t *testing.T) {
	got := watermarkOutputBufferSize(1024, image.Rect(0, 0, 120, 80))
	if got != minWatermarkOutputSize {
		t.Fatalf("expected minimum output size %d, got %d", minWatermarkOutputSize, got)
	}
}

// TestMigratedHotGoWatermarkStaticLibraryPresent verifies the current platform
// has the prebuilt FFmpeg/C library required by the cgo adapter.
func TestMigratedHotGoWatermarkStaticLibraryPresent(t *testing.T) {
	libraryName := "lib" + runtime.GOARCH + "_watermark.a"
	info, err := os.Stat(libraryName)
	if err != nil {
		t.Fatalf("stat migrated hotgo static library %s: %v", libraryName, err)
	}
	if info.Size() == 0 {
		t.Fatalf("expected migrated hotgo static library %s to be non-empty", libraryName)
	}
}
