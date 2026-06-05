// Package posterrender is the replaceable activation-poster PNG output seam for
// the sicau-niu C3 capability. It defines the stable PosterRenderer contract that
// the activation service depends on and ships a default implementation that
// composes the key poster fields into a shareable PNG. The rendering seam keeps the
// activation closure decoupled from image production so a richer brand template can
// replace the default renderer without changing service logic.
package posterrender

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/rendertext"
)

// PosterData carries the composed activation-poster fields handed to the renderer.
// It is a plain value object owned by this seam so the renderer never depends on
// the activation service internals.
type PosterData struct {
	// Nickname is the player nickname.
	Nickname string
	// IdentityType is the raw player identity type.
	IdentityType string
	// NiuCode is the activated cattle serial code.
	NiuCode string
	// OrderNo is the player's arrival order for the cattle, starting at 1.
	OrderNo int
	// Quote is the selected random school-history quote; may be empty.
	Quote string
	// CampusBadge is the campus anniversary badge text; may be empty.
	CampusBadge string
}

// PosterRenderer renders activation-poster composition data into a PNG image.
type PosterRenderer interface {
	// Render encodes data into a PNG byte slice. It returns CodePosterRenderFailed
	// when encoding fails. The default implementation produces a deterministic
	// field-bearing poster image; richer brand templates replace it behind this seam.
	Render(ctx context.Context, data *PosterData) (png []byte, err error)
}

// Interface compliance assertion for the default poster renderer implementation.
var _ PosterRenderer = (*basicRenderer)(nil)

// basicRenderer is the default poster renderer. It draws a complete poster layout
// with visible text fields and a stable per-player accent, while keeping the
// implementation dependency-light and fully deterministic.
type basicRenderer struct{}

// New creates the default basic poster renderer.
func New() PosterRenderer {
	return &basicRenderer{}
}

// Poster canvas dimensions, frame inset and accent band height, in pixels. The
// portrait canvas matches a share-friendly poster aspect ratio.
const (
	posterWidth  = 600
	posterHeight = 800
	posterInset  = 28
	posterAccent = 120
)

var (
	posterBackground = color.RGBA{R: 0x0f, G: 0x24, B: 0x17, A: 0xff}
	posterPanel      = color.RGBA{R: 0xfb, G: 0xf7, B: 0xea, A: 0xff}
	posterGold       = color.RGBA{R: 0xd8, G: 0xb2, B: 0x4a, A: 0xff}
	posterDarkGold   = color.RGBA{R: 0x91, G: 0x68, B: 0x23, A: 0xff}
	posterTextInk    = color.RGBA{R: 0x19, G: 0x26, B: 0x1e, A: 0xff}
	posterMutedInk   = color.RGBA{R: 0x68, G: 0x70, B: 0x58, A: 0xff}
	posterWhite      = color.RGBA{R: 0xff, G: 0xfb, B: 0xef, A: 0xff}
)

// Render encodes a portrait poster PNG personalized by the player's data. The
// background is campus green, an inner content panel carries the player, code,
// order, identity and quote fields, and a hash-derived accent makes every poster
// visually distinct. The bytes are returned to the caller for base64 delivery.
func (r *basicRenderer) Render(ctx context.Context, data *PosterData) ([]byte, error) {
	if data == nil {
		return nil, bizerr.NewCode(CodePosterRenderFailed)
	}

	img := image.NewRGBA(image.Rect(0, 0, posterWidth, posterHeight))
	accent := posterAccentColor(data)

	fillPosterRect(img, img.Bounds(), posterBackground)
	fillPosterRect(img, image.Rect(0, 0, posterWidth, posterAccent), accent)
	fillPosterRect(img, image.Rect(38, 142, posterWidth-38, posterHeight-36), posterPanel)
	strokePosterRect(img, image.Rect(posterInset, posterInset, posterWidth-posterInset, posterHeight-posterInset), 3, posterGold)
	strokePosterRect(img, image.Rect(52, 156, posterWidth-52, posterHeight-52), 2, posterDarkGold)

	drawPosterTextBold(img, 52, 54, "SICAU NIU HUNT", posterWhite)
	drawPosterText(img, 52, 78, "ANNIVERSARY CAMPUS GAME", posterWhite)
	drawPosterText(img, 52, 98, fitPosterText(data.CampusBadge, 36), posterWhite)
	drawPosterHashBars(img, image.Rect(416, 46, 546, 92), data.Nickname+"|"+data.NiuCode, posterGold)

	drawPosterText(img, 78, 196, "PLAYER", posterMutedInk)
	drawPosterTextBold(img, 78, 224, fitPosterText(data.Nickname, 48), posterTextInk)

	drawPosterText(img, 78, 274, "NIU CODE", posterMutedInk)
	drawPosterTextBold(img, 78, 302, fitPosterText(data.NiuCode, 42), posterTextInk)

	drawPosterText(img, 78, 352, "ARRIVAL ORDER", posterMutedInk)
	drawPosterTextBold(img, 78, 380, "#"+strconv.Itoa(data.OrderNo), posterTextInk)

	drawPosterText(img, 310, 352, "IDENTITY", posterMutedInk)
	drawPosterTextBold(img, 310, 380, fitPosterText(data.IdentityType, 26), posterTextInk)

	drawPosterDivider(img, 78, 424, 444)
	drawPosterText(img, 78, 466, "CAMPUS QUOTE", posterMutedInk)
	for i, line := range wrapPosterText(data.Quote, 54, 4) {
		drawPosterText(img, 78, 496+i*24, line, posterTextInk)
	}

	fillPosterRect(img, image.Rect(78, 642, 522, 704), color.RGBA{R: 0x17, G: 0x35, B: 0x25, A: 0xff})
	drawPosterTextBold(img, 104, 678, "120TH ANNIVERSARY", posterGold)
	drawPosterText(img, 104, 698, "SICHUAN AGRICULTURAL UNIVERSITY", posterWhite)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, bizerr.WrapCode(err, CodePosterRenderFailed)
	}
	return buf.Bytes(), nil
}

// posterAccentColor derives a stable accent colour from the player nickname and
// cattle code so each player's poster is visually distinct. The seed is an inline
// FNV-1a hash over the player identity.
func posterAccentColor(data *PosterData) color.RGBA {
	sum := posterFNV1a(data.Nickname + "|" + data.NiuCode)
	return color.RGBA{
		R: byte(0x80 + sum%0x80),
		G: byte(0x80 + (sum>>8)%0x80),
		B: byte(0x80 + (sum>>16)%0x80),
		A: 0xff,
	}
}

// posterFNV1a computes the 32-bit FNV-1a hash of s inline, avoiding a hash.Hash
// whose Write return values would need discarding.
func posterFNV1a(s string) uint32 {
	const (
		offset = 2166136261
		prime  = 16777619
	)
	hash := uint32(offset)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime
	}
	return hash
}

func fillPosterRect(img draw.Image, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect.Intersect(img.Bounds()), image.NewUniform(c), image.Point{}, draw.Src)
}

func strokePosterRect(img draw.Image, rect image.Rectangle, width int, c color.Color) {
	fillPosterRect(img, image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+width), c)
	fillPosterRect(img, image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y), c)
	fillPosterRect(img, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y), c)
	fillPosterRect(img, image.Rect(rect.Max.X-width, rect.Min.Y, rect.Max.X, rect.Max.Y), c)
}

func drawPosterText(img draw.Image, x, y int, text string, c color.Color) {
	rendertext.DrawString(img, x, y, text, c)
}

func drawPosterTextBold(img draw.Image, x, y int, text string, c color.Color) {
	rendertext.DrawStringBold(img, x, y, text, c)
}

func drawPosterDivider(img draw.Image, x, y, width int) {
	fillPosterRect(img, image.Rect(x, y, x+width, y+2), posterDarkGold)
	fillPosterRect(img, image.Rect(x, y+8, x+width, y+9), posterGold)
}

func drawPosterHashBars(img draw.Image, rect image.Rectangle, seed string, c color.Color) {
	sum := posterFNV1a(seed)
	fillPosterRect(img, rect, color.RGBA{R: 0xff, G: 0xfb, B: 0xef, A: 0xff})
	for i := 0; i < 12; i++ {
		barHeight := 10 + int((sum>>uint((i%4)*8))&0x1f)
		x := rect.Min.X + 8 + i*9
		fillPosterRect(img, image.Rect(x, rect.Max.Y-8-barHeight, x+5, rect.Max.Y-8), c)
	}
	strokePosterRect(img, rect, 2, posterDarkGold)
}

func fitPosterText(text string, maxChars int) string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if text == "" {
		text = "-"
	}
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return string(runes[:maxChars])
	}
	return string(runes[:maxChars-3]) + "..."
}

func wrapPosterText(text string, maxChars, maxLines int) []string {
	text = strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
	if text == "" {
		text = "-"
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{"-"}
	}

	lines := make([]string, 0, maxLines)
	current := ""
	for _, word := range words {
		if len([]rune(word)) > maxChars {
			word = fitPosterText(word, maxChars)
		}
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) <= maxChars {
			current += " " + word
			continue
		}
		lines = append(lines, current)
		current = word
		if len(lines) == maxLines {
			return lines
		}
	}
	if current != "" && len(lines) < maxLines {
		lines = append(lines, current)
	}
	return lines
}
