// certrender_test.go covers the default certificate renderer: it encodes a
// decodable field-bearing PNG for valid data, personalizes distinct holders and
// rejects nil data. These are pure-logic tests with no database dependency.

package certrender

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"testing"

	"lina-core/pkg/bizerr"
)

// TestRenderEncodesDecodablePNG verifies the default renderer produces non-empty
// bytes that decode as a PNG image of the expected certificate size.
func TestRenderEncodesDecodablePNG(t *testing.T) {
	renderer := New()
	data := &CertData{
		Nickname:    "Certificate Student",
		HonorName:   "120th Anniversary Certificate",
		HonorCode:   "CERT-001",
		CampusBadge: "SICAU 120",
	}
	out, err := renderer.Render(context.Background(), data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected non-empty certificate PNG bytes")
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("expected decodable PNG, got decode error: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != certWidth || bounds.Dy() != certHeight {
		t.Fatalf("expected %dx%d image, got %dx%d", certWidth, certHeight, bounds.Dx(), bounds.Dy())
	}
	if countCertColor(img, certTextInk) < 300 {
		t.Fatal("expected certificate to contain visible text ink pixels")
	}
	if countCertColor(img, certPanel) < 90000 {
		t.Fatal("expected certificate to contain a readable content panel")
	}
}

// TestRenderDistinctPerHolder verifies two different holders' certificates differ,
// so the rendered image is personalized rather than a fixed placeholder.
func TestRenderDistinctPerHolder(t *testing.T) {
	renderer := New()
	base := CertData{HonorName: "Certificate", HonorCode: "CERT-001"}

	a := base
	a.Nickname = "Holder A"
	b := base
	b.Nickname = "Holder B"

	outA, err := renderer.Render(context.Background(), &a)
	if err != nil {
		t.Fatalf("render A failed: %v", err)
	}
	outB, err := renderer.Render(context.Background(), &b)
	if err != nil {
		t.Fatalf("render B failed: %v", err)
	}
	if bytes.Equal(outA, outB) {
		t.Fatal("distinct holders produced identical certificates")
	}
}

// TestRenderNilDataRejected verifies nil certificate data is rejected with the
// render business error code.
func TestRenderNilDataRejected(t *testing.T) {
	_, err := New().Render(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil certificate data")
	}
	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != CodeCertRenderFailed.RuntimeCode() {
		t.Fatalf("expected CodeCertRenderFailed, got %v", err)
	}
}

func countCertColor(img image.Image, want color.RGBA) int {
	bounds := img.Bounds()
	count := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.RGBAModel.Convert(img.At(x, y)) == want {
				count++
			}
		}
	}
	return count
}
