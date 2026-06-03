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

// placeholderEdge is the side length in pixels of the placeholder poster image.
const placeholderEdge = 64

// Render encodes a deterministic placeholder PNG. The poster fields are carried
// in the response DTO for the frontend; this default output keeps the PNG seam
// stable until the design-grade renderer replaces it.
func (r *basicRenderer) Render(ctx context.Context, data *PosterData) ([]byte, error) {
	if data == nil {
		return nil, bizerr.NewCode(CodePosterRenderFailed)
	}
	img := image.NewRGBA(image.Rect(0, 0, placeholderEdge, placeholderEdge))
	background := color.RGBA{R: 0x1f, G: 0x6f, B: 0x3f, A: 0xff}
	for y := 0; y < placeholderEdge; y++ {
		for x := 0; x < placeholderEdge; x++ {
			img.Set(x, y, background)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, bizerr.WrapCode(err, CodePosterRenderFailed)
	}
	return buf.Bytes(), nil
}
