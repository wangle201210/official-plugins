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

// TestRenderEncodesDecodablePNG verifies the default renderer produces non-empty
// bytes that decode as a PNG image of the expected portrait poster size.
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
	if len(out) == 0 {
		t.Fatal("expected non-empty poster PNG bytes")
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("expected decodable PNG, got decode error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != posterWidth || bounds.Dy() != posterHeight {
		t.Fatalf("expected %dx%d image, got %dx%d", posterWidth, posterHeight, bounds.Dx(), bounds.Dy())
	}
}

// TestRenderDistinctPerPlayer verifies two different players' posters differ, so
// the rendered image is personalized rather than a fixed placeholder.
func TestRenderDistinctPerPlayer(t *testing.T) {
	renderer := New()
	base := PosterData{NiuCode: "NIU-001", OrderNo: 1}

	a := base
	a.Nickname = "玩家甲"
	b := base
	b.Nickname = "玩家乙"

	outA, err := renderer.Render(context.Background(), &a)
	if err != nil {
		t.Fatalf("render A failed: %v", err)
	}
	outB, err := renderer.Render(context.Background(), &b)
	if err != nil {
		t.Fatalf("render B failed: %v", err)
	}
	if bytes.Equal(outA, outB) {
		t.Fatal("distinct players produced identical posters")
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
