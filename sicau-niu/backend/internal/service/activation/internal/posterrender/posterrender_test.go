// posterrender_test.go covers the basic poster renderer: it encodes a decodable
// PNG for valid data and rejects nil data. These are pure-logic tests with no
// database dependency.

package posterrender

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestRenderEncodesDecodablePNG verifies the default renderer produces bytes that
// decode as a PNG image of the expected placeholder size.
func TestRenderEncodesDecodablePNG(t *testing.T) {
	renderer := New()
	data := &PosterData{
		Nickname:     "川农牛同学",
		IdentityType: "student",
		NiuCode:      "NIU-001",
		OrderNo:      1,
		Quote:        "任重道远",
		CampusBadge:  "校庆",
	}
	out, err := renderer.Render(context.Background(), data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("expected decodable PNG, got decode error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != placeholderEdge || bounds.Dy() != placeholderEdge {
		t.Fatalf("expected %dx%d image, got %dx%d", placeholderEdge, placeholderEdge, bounds.Dx(), bounds.Dy())
	}
}

// TestRenderNilDataRejected verifies nil poster data is rejected with the render
// business error code.
func TestRenderNilDataRejected(t *testing.T) {
	_, err := New().Render(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil poster data")
	}
	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != CodePosterRenderFailed.RuntimeCode() {
		t.Fatalf("expected CodePosterRenderFailed, got %v", err)
	}
}
