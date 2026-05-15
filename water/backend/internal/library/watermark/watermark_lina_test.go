//go:build cgo

package watermark

import (
	"os"
	"strings"
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
		TextSetting: TextSetting{
			Text:     "LinaPro 水印测试",
			FontSize: 32,
			Color:    "white",
			Align:    AlignmentBottomRight,
		},
		ImageSetting: ImageSetting{
			Image:   "background.png",
			Opacity: 0.15,
		},
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

// TestMigratedHotGoWatermarkSourceUnmodified verifies the critical C source
// still contains HotGo's FFmpeg entrypoint name.
func TestMigratedHotGoWatermarkSourceUnmodified(t *testing.T) {
	content, err := os.ReadFile("watermark.c")
	if err != nil {
		t.Fatalf("read migrated watermark.c: %v", err)
	}
	if !strings.Contains(string(content), "int process_jpg_watermark") {
		t.Fatal("expected migrated watermark.c to keep HotGo process_jpg_watermark entrypoint")
	}
}
