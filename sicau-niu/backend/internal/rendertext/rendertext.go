// Package rendertext provides small deterministic text drawing helpers for plugin
// image renderers. It prefers an explicitly configured or common system CJK font
// and falls back to the bundled basic bitmap font when no font file is available.
package rendertext

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"strings"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// FontEnv is the optional environment variable used by renderers to locate a CJK
// capable TTF, OTF or TTC font file.
const FontEnv = "SICAU_NIU_RENDER_FONT"

var (
	loadOnce sync.Once
	loadFace font.Face
)

// DrawString draws text at the baseline point. It uses the best available CJK font
// when present and falls back to basic ASCII-safe text otherwise.
func DrawString(img draw.Image, x, y int, text string, c color.Color) {
	face := face()
	if face == basicfont.Face7x13 {
		text = ASCIISafe(text)
	}
	drawer := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.P(x, y),
	}
	drawer.DrawString(text)
}

// DrawStringBold draws a lightweight bold effect by painting the text with a small
// one-pixel offset.
func DrawStringBold(img draw.Image, x, y int, text string, c color.Color) {
	DrawString(img, x, y, text, c)
	DrawString(img, x+1, y, text, c)
	DrawString(img, x, y+1, text, c)
}

// ASCIISafe converts text into a stable string the fallback bitmap font can render.
func ASCIISafe(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return "-"
	}
	var b strings.Builder
	for _, r := range text {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			b.WriteByte(' ')
		case r >= 32 && r <= 126:
			b.WriteRune(r)
		default:
			b.WriteByte('?')
		}
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	if out == "" {
		return "-"
	}
	return out
}

func face() font.Face {
	loadOnce.Do(func() {
		loadFace = basicfont.Face7x13
		for _, path := range candidateFontPaths() {
			face, err := loadFontFace(path)
			if err == nil {
				loadFace = face
				return
			}
		}
	})
	return loadFace
}

func candidateFontPaths() []string {
	paths := []string{}
	if path := strings.TrimSpace(os.Getenv(FontEnv)); path != "" {
		paths = append(paths, path)
	}
	paths = append(paths,
		"/System/Library/Fonts/STHeiti Medium.ttc",
		"/System/Library/Fonts/Hiragino Sans GB.ttc",
		"/System/Library/Fonts/Supplemental/Songti.ttc",
		"/System/Library/Fonts/STHeiti Light.ttc",
		"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc",
		"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
		"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
	)
	return paths
}

func loadFontFace(path string) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parsed, err := opentype.Parse(data)
	if err != nil {
		collection, collectionErr := opentype.ParseCollection(data)
		if collectionErr != nil {
			return nil, err
		}
		parsed, err = collection.Font(0)
		if err != nil {
			return nil, err
		}
	}
	return opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    18,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}
