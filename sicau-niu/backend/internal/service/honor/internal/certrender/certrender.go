// Package certrender is the replaceable electronic-certificate PNG output seam for
// the sicau-niu certificate honors. It defines the stable CertRenderer contract the
// honor service depends on and ships a basic default implementation that encodes a
// real certificate-frame PNG whose accent colour is derived from the holder's data,
// so every holder's certificate is visually distinct. Unlike the activation poster
// seam, the encoded bytes ARE returned to the caller. The design-grade CJK layout
// and font rendering are intended to replace this default without changing the
// seam, so the honor service stays decoupled from the rendering implementation. The
// seam is kept minimal: only the contract, the certificate data value object and the
// default constructor are exported.
package certrender

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"

	"lina-core/pkg/bizerr"
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
	// when encoding fails. The default implementation produces a certificate-frame
	// image whose accent colour is derived from the holder's data so different
	// holders get distinct certificates; the design-grade renderer replaces it
	// behind this seam.
	Render(ctx context.Context, data *CertData) (png []byte, err error)
}

// Interface compliance assertion for the default certificate renderer.
var _ CertRenderer = (*basicRenderer)(nil)

// basicRenderer is the default certificate renderer. It draws a campus-coloured
// frame with a per-holder accent band so the output seam is exercised end to end
// and every holder's certificate is visually distinct, without bundling CJK fonts;
// the real layout renderer replaces it without changing the contract.
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

// Render encodes a certificate-frame PNG personalized by the holder's data. The
// background is campus green, an inner gold frame is drawn, and a top accent band
// uses a colour derived from a hash of the holder nickname and honor code so each
// holder's certificate is distinct. Readable CJK text layout is the design-grade
// renderer that replaces this default behind the seam.
func (r *basicRenderer) Render(ctx context.Context, data *CertData) ([]byte, error) {
	if data == nil {
		return nil, bizerr.NewCode(CodeCertRenderFailed)
	}

	img := image.NewRGBA(image.Rect(0, 0, certWidth, certHeight))
	background := color.RGBA{R: 0x0f, G: 0x24, B: 0x17, A: 0xff}
	gold := color.RGBA{R: 0xd8, G: 0xb2, B: 0x4a, A: 0xff}
	accent := holderAccent(data)

	for y := 0; y < certHeight; y++ {
		for x := 0; x < certWidth; x++ {
			switch {
			case y < accentBand:
				img.Set(x, y, accent)
			case isFrameEdge(x, y):
				img.Set(x, y, gold)
			default:
				img.Set(x, y, background)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, bizerr.WrapCode(err, CodeCertRenderFailed)
	}
	return buf.Bytes(), nil
}

// isFrameEdge reports whether (x,y) lies on the inner gold frame border.
func isFrameEdge(x, y int) bool {
	onVertical := (x == frameInset || x == certWidth-1-frameInset) && y >= frameInset && y <= certHeight-1-frameInset
	onHorizontal := (y == frameInset || y == certHeight-1-frameInset) && x >= frameInset && x <= certWidth-1-frameInset
	return onVertical || onHorizontal
}

// holderAccent derives a stable accent colour from the holder nickname and honor
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
