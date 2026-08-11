// Package certrender is the replaceable electronic-certificate PNG output seam for
// the sicau-niu certificate honors. It defines the stable CertRenderer contract the
// honor service depends on and ships a default implementation that composes the
// certificate fields into a personalized PNG. The seam keeps the honor service
// decoupled from image production so richer brand templates can replace the default
// renderer without changing business logic.
package certrender

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/rendertext"
)

// CertData carries the composed certificate fields handed to the renderer. It is a
// plain value object owned by this seam so the renderer never depends on the honor
// service internals.
type CertData struct {
	// Nickname is the certificate holder's nickname.
	Nickname string
	// HonorName is the certificate honor display name.
	HonorName string
	// HonorCode is the certificate honor unique code.
	HonorCode string
	// CampusBadge is the campus anniversary badge text; may be empty.
	CampusBadge string
}

// CertRenderer renders certificate composition data into a PNG image.
type CertRenderer interface {
	// Render encodes data into a PNG byte slice. It returns CodeCertRenderFailed
	// when encoding fails. The default implementation produces a field-bearing
	// certificate image whose accent color is derived from the holder's data.
	Render(ctx context.Context, data *CertData) (png []byte, err error)
}

// Interface compliance assertion for the default certificate renderer.
var _ CertRenderer = (*basicRenderer)(nil)

// basicRenderer is the default certificate renderer. It draws a complete
// certificate layout with visible text fields and a stable per-holder accent, while
// keeping the implementation dependency-light and fully deterministic.
type basicRenderer struct{}

// New creates the default basic certificate renderer.
func New() CertRenderer {
	return &basicRenderer{}
}

// Certificate canvas dimensions and frame inset, in pixels.
const (
	certWidth  = 600
	certHeight = 400
	frameInset = 24
	accentBand = 40
)

var (
	certBackground = color.RGBA{R: 0x0f, G: 0x24, B: 0x17, A: 0xff}
	certPanel      = color.RGBA{R: 0xfb, G: 0xf7, B: 0xea, A: 0xff}
	certGold       = color.RGBA{R: 0xd8, G: 0xb2, B: 0x4a, A: 0xff}
	certDarkGold   = color.RGBA{R: 0x91, G: 0x68, B: 0x23, A: 0xff}
	certTextInk    = color.RGBA{R: 0x19, G: 0x26, B: 0x1e, A: 0xff}
	certMutedInk   = color.RGBA{R: 0x68, G: 0x70, B: 0x58, A: 0xff}
	certWhite      = color.RGBA{R: 0xff, G: 0xfb, B: 0xef, A: 0xff}
)

// Render encodes a certificate PNG personalized by the holder's data. The
// background is campus green, an inner certificate panel carries the holder, honor,
// code and badge fields, and a hash-derived accent makes every certificate visually
// distinct.
func (r *basicRenderer) Render(ctx context.Context, data *CertData) ([]byte, error) {
	if data == nil {
		return nil, bizerr.NewCode(CodeCertRenderFailed)
	}

	img := image.NewRGBA(image.Rect(0, 0, certWidth, certHeight))
	accent := holderAccent(data)

	fillCertRect(img, img.Bounds(), certBackground)
	fillCertRect(img, image.Rect(0, 0, certWidth, accentBand), accent)
	fillCertRect(img, image.Rect(46, 70, certWidth-46, certHeight-34), certPanel)
	strokeCertRect(img, image.Rect(frameInset, frameInset, certWidth-frameInset, certHeight-frameInset), 3, certGold)
	strokeCertRect(img, image.Rect(60, 84, certWidth-60, certHeight-48), 2, certDarkGold)

	drawCertTextBold(img, 46, 30, "SICAU NIU CERTIFICATE", certWhite)
	drawCertText(img, 386, 30, fitCertText(data.CampusBadge, 26), certWhite)

	drawCertText(img, 86, 124, "AWARDED TO", certMutedInk)
	drawCertTextBold(img, 86, 152, fitCertText(data.Nickname, 40), certTextInk)

	drawCertDivider(img, 86, 180, 428)
	drawCertText(img, 86, 224, "HONOR", certMutedInk)
	drawCertTextBold(img, 86, 252, fitCertText(data.HonorName, 52), certTextInk)

	drawCertText(img, 86, 306, "CERTIFICATE CODE", certMutedInk)
	drawCertTextBold(img, 86, 334, fitCertText(data.HonorCode, 42), certTextInk)

	drawCertSeal(img, 460, 260, data.Nickname+"|"+data.HonorCode, accent)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, bizerr.WrapCode(err, CodeCertRenderFailed)
	}
	return buf.Bytes(), nil
}

// holderAccent derives a stable accent color from the holder nickname and honor
// code so each holder's certificate is visually distinct. The seed is an inline
// FNV-1a hash over the holder identity; the hue spread keeps the band readable
// against the dark background.
func holderAccent(data *CertData) color.RGBA {
	sum := fnv1a(data.Nickname + "|" + data.HonorCode)
	return color.RGBA{
		R: byte(0x80 + sum%0x80),
		G: byte(0x80 + (sum>>8)%0x80),
		B: byte(0x80 + (sum>>16)%0x80),
		A: 0xff,
	}
}

// fnv1a computes the 32-bit FNV-1a hash of s inline, avoiding a hash.Hash whose
// Write return values would need discarding.
func fnv1a(s string) uint32 {
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

func fillCertRect(img draw.Image, rect image.Rectangle, c color.Color) {
	draw.Draw(img, rect.Intersect(img.Bounds()), image.NewUniform(c), image.Point{}, draw.Src)
}

func strokeCertRect(img draw.Image, rect image.Rectangle, width int, c color.Color) {
	fillCertRect(img, image.Rect(rect.Min.X, rect.Min.Y, rect.Max.X, rect.Min.Y+width), c)
	fillCertRect(img, image.Rect(rect.Min.X, rect.Max.Y-width, rect.Max.X, rect.Max.Y), c)
	fillCertRect(img, image.Rect(rect.Min.X, rect.Min.Y, rect.Min.X+width, rect.Max.Y), c)
	fillCertRect(img, image.Rect(rect.Max.X-width, rect.Min.Y, rect.Max.X, rect.Max.Y), c)
}

func drawCertText(img draw.Image, x, y int, text string, c color.Color) {
	rendertext.DrawString(img, x, y, text, c)
}

func drawCertTextBold(img draw.Image, x, y int, text string, c color.Color) {
	rendertext.DrawStringBold(img, x, y, text, c)
}

func drawCertDivider(img draw.Image, x, y, width int) {
	fillCertRect(img, image.Rect(x, y, x+width, y+2), certDarkGold)
	fillCertRect(img, image.Rect(x, y+8, x+width, y+9), certGold)
}

func drawCertSeal(img draw.Image, x, y int, seed string, c color.Color) {
	fillCertRect(img, image.Rect(x, y, x+74, y+74), certBackground)
	strokeCertRect(img, image.Rect(x+5, y+5, x+69, y+69), 2, certGold)
	sum := fnv1a(seed)
	for i := 0; i < 9; i++ {
		barHeight := 12 + int((sum>>uint((i%4)*8))&0x1f)
		fillCertRect(img, image.Rect(x+13+i*6, y+58-barHeight, x+17+i*6, y+58), c)
	}
	drawCertText(img, x+18, y+69, "SEAL", certWhite)
}

func fitCertText(text string, maxChars int) string {
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
