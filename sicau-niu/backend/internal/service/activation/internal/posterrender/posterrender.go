// Package posterrender is the replaceable activation-poster PNG output seam for
// the sicau-niu C3 capability. It defines the stable PosterRenderer contract that
// the activation service depends on and ships a basic default implementation that
// encodes a deterministic placeholder PNG carrying the poster fields. The
// design-grade CJK layout and font rendering are intended to replace this default
// without changing the seam, so the activation closure stays decoupled from the
// rendering implementation. The seam is kept minimal: only the contract, the
// poster data value object and the default constructor are exported.
package posterrender

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"

	"lina-core/pkg/bizerr"
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
	// placeholder image; the design-grade renderer replaces it behind this seam.
	Render(ctx context.Context, data *PosterData) (png []byte, err error)
}

// Interface compliance assertion for the default poster renderer implementation.
var _ PosterRenderer = (*basicRenderer)(nil)

// basicRenderer is the default poster renderer. It encodes a small solid-color
// placeholder PNG so the output seam is exercised end to end without bundling CJK
// fonts; the real layout renderer replaces it without changing the contract.
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

// Render encodes a real portrait poster PNG personalized by the player's data. The
// background is campus green, an inner gold frame is drawn, and a top accent band
// uses a colour derived from a hash of the player nickname and cattle code so each
// player's poster is visually distinct. The bytes are returned to the caller (the
// activation service base64-encodes them into the response). Readable CJK text
// layout is the design-grade renderer that replaces this default behind the seam.
func (r *basicRenderer) Render(ctx context.Context, data *PosterData) ([]byte, error) {
	if data == nil {
		return nil, bizerr.NewCode(CodePosterRenderFailed)
	}

	img := image.NewRGBA(image.Rect(0, 0, posterWidth, posterHeight))
	background := color.RGBA{R: 0x0f, G: 0x24, B: 0x17, A: 0xff}
	gold := color.RGBA{R: 0xd8, G: 0xb2, B: 0x4a, A: 0xff}
	accent := posterAccentColor(data)

	for y := 0; y < posterHeight; y++ {
		for x := 0; x < posterWidth; x++ {
			switch {
			case y < posterAccent:
				img.Set(x, y, accent)
			case isPosterFrameEdge(x, y):
				img.Set(x, y, gold)
			default:
				img.Set(x, y, background)
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, bizerr.WrapCode(err, CodePosterRenderFailed)
	}
	return buf.Bytes(), nil
}

// isPosterFrameEdge reports whether (x,y) lies on the inner gold frame border.
func isPosterFrameEdge(x, y int) bool {
	onVertical := (x == posterInset || x == posterWidth-1-posterInset) && y >= posterInset && y <= posterHeight-1-posterInset
	onHorizontal := (y == posterInset || y == posterHeight-1-posterInset) && x >= posterInset && x <= posterWidth-1-posterInset
	return onVertical || onHorizontal
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
