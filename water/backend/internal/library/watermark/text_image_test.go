package watermark

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadTrueTypeFontSelectsCJKCollectionFace verifies TTC collections choose
// a sub-font with real Chinese glyph mappings instead of the first face.
func TestLoadTrueTypeFontSelectsCJKCollectionFace(t *testing.T) {
	fontData, err := os.ReadFile(filepath.Join("fonts", "STHeiti Medium.ttc"))
	if err != nil {
		t.Fatalf("read embedded STHeiti font: %v", err)
	}
	if offsets, ok := trueTypeCollectionOffsets(fontData); !ok || len(offsets) == 0 {
		t.Fatalf("expected embedded STHeiti to be a TTC collection, ok=%t offsets=%v", ok, offsets)
	}
	ttFont, err := loadTrueTypeFont(fontData)
	if err != nil {
		t.Fatalf("load embedded STHeiti TrueType font: %v", err)
	}
	if score := trueTypeGlyphCoverageScore(ttFont, cjkGlyphProbeText); score != len([]rune(cjkGlyphProbeText)) {
		t.Fatalf("expected selected STHeiti face to cover probe glyphs, score=%d", score)
	}
}

// TestLoadFontFaceFromFileSupportsEmbeddedSTHeitiChineseGlyph verifies the
// embedded STHeiti TTC can render CJK glyphs on Linux containers without
// relying on system fonts.
func TestLoadFontFaceFromFileSupportsEmbeddedSTHeitiChineseGlyph(t *testing.T) {
	face, err := loadFontFaceFromFile(filepath.Join("fonts", "STHeiti Medium.ttc"), 32)
	if err != nil {
		t.Fatalf("load embedded STHeiti font face: %v", err)
	}
	defer face.Close()

	advance, ok := face.GlyphAdvance('中')
	if !ok || advance <= 0 {
		t.Fatalf("expected embedded STHeiti to provide Chinese glyph advance, ok=%t advance=%v", ok, advance)
	}
	bounds, _, ok := face.GlyphBounds('中')
	if !ok || bounds.Empty() {
		t.Fatalf("expected embedded STHeiti to provide Chinese glyph bounds, ok=%t bounds=%v", ok, bounds)
	}
}

// TestLoadFontFaceRejectsInvalidExplicitFont verifies configured fonts are not
// silently replaced with a fallback that might miss CJK glyphs.
func TestLoadFontFaceRejectsInvalidExplicitFont(t *testing.T) {
	_, err := loadFontFace(filepath.Join(t.TempDir(), "missing-font.ttf"), 32)
	if err == nil {
		t.Fatal("expected explicit missing font to return an error")
	}
}
