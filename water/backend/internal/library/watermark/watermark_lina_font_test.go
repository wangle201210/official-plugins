//go:build cgo

package watermark

import (
	"os"
	"strings"
	"testing"
)

// TestWithDefaultFontMaterializesChineseFont verifies Chinese text watermarks get a CJK-capable font.
func TestWithDefaultFontMaterializesChineseFont(t *testing.T) {
	config, err := withDefaultFont(WatermarkConfig{
		TextSetting: TextSetting{
			Text: "LinaPro 水印测试",
		},
	})
	if err != nil {
		t.Fatalf("materialize default watermark font: %v", err)
	}
	if !strings.HasSuffix(config.TextSetting.Font, defaultChineseFontFileName) {
		t.Fatalf("expected default Chinese font path, got %q", config.TextSetting.Font)
	}
	content, err := os.ReadFile(config.TextSetting.Font)
	if err != nil {
		t.Fatalf("read materialized watermark font: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected materialized watermark font content")
	}
}

// TestWithDefaultFontKeepsExplicitFont verifies caller-provided fonts remain authoritative.
func TestWithDefaultFontKeepsExplicitFont(t *testing.T) {
	config, err := withDefaultFont(WatermarkConfig{
		TextSetting: TextSetting{
			Text: "LinaPro 水印测试",
			Font: "/custom/font.ttc",
		},
	})
	if err != nil {
		t.Fatalf("normalize explicit watermark font: %v", err)
	}
	if config.TextSetting.Font != "/custom/font.ttc" {
		t.Fatalf("expected explicit font to be preserved, got %q", config.TextSetting.Font)
	}
}
