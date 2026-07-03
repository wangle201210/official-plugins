// This file verifies watermark text image font fallback behavior.

package watermark

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenerateTiledTextImageUsesFallbackFont verifies text rendering no longer
// depends on the cgo adapter materializing a package-owned default font.
func TestGenerateTiledTextImageUsesFallbackFont(t *testing.T) {
	cases := []struct {
		name string
		font string
	}{
		{name: "empty font"},
		{name: "unavailable explicit font", font: "/not-found/font.ttc"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			outputPath := filepath.Join(t.TempDir(), "watermark.png")
			err := generateTiledTextImage("LinaPro 水印测试", tc.font, 24, "#ffffff", 160, 90, -30, outputPath)
			if err != nil {
				t.Fatalf("generate tiled text image: %v", err)
			}
			info, err := os.Stat(outputPath)
			if err != nil {
				t.Fatalf("stat generated image: %v", err)
			}
			if info.Size() == 0 {
				t.Fatal("expected generated image to be non-empty")
			}
		})
	}
}
