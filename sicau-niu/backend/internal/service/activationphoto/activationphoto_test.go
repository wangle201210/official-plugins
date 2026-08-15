// activationphoto_test.go covers the image standardization guards: the header
// pixel budget that rejects decompression bombs before any decode, the single
// upfront downscale, the bounded output size and the transcoding concurrency
// gate. These are pure-logic tests with no database dependency.

package activationphoto

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// photoBytes encodes a smooth, camera-like JPEG so the fixture compresses the way
// a real photo does instead of behaving like incompressible noise.
func photoBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	random := rand.New(rand.NewSource(7))
	for y := 0; y < height; y++ {
		vertical := float64(y) / float64(height)
		for x := 0; x < width; x++ {
			horizontal := float64(x) / float64(width)
			shade := 0.5 + 0.5*math.Sin(horizontal*3)*math.Cos(vertical*2)
			noise := float64(random.Intn(7) - 3)
			img.Set(x, y, color.NRGBA{
				R: clampByte(120 + 90*vertical + 20*shade + noise),
				G: clampByte(140 + 70*vertical + 25*shade + noise),
				B: clampByte(190 - 40*vertical + 15*shade + noise),
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode fixture jpeg failed: %v", err)
	}
	return buf.Bytes()
}

func clampByte(value float64) uint8 {
	switch {
	case value < 0:
		return 0
	case value > 255:
		return 255
	default:
		return uint8(value)
	}
}

// bombPNG returns a tiny PNG whose header declares an enormous canvas. Decoding it
// would allocate width*height*4 bytes, so the fixture proves the guard runs on the
// header rather than after a decode.
func bombPNG(t *testing.T, width, height uint32) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatalf("encode seed png failed: %v", err)
	}
	data := buf.Bytes()
	// IHDR data starts after the 8-byte signature, the 4-byte length and the 4-byte
	// chunk type; width and height are its first two big-endian uint32 fields.
	const ihdrData = 8 + 4 + 4
	binary.BigEndian.PutUint32(data[ihdrData:], width)
	binary.BigEndian.PutUint32(data[ihdrData+4:], height)
	// The CRC covers the chunk type and its 13 data bytes.
	crc := crc32.ChecksumIEEE(data[ihdrData-4 : ihdrData+13])
	binary.BigEndian.PutUint32(data[ihdrData+13:], crc)
	return data
}

// TestStandardizeRejectsOversizedHeaderBeforeDecode verifies a small file that
// declares more than the pixel budget is rejected, quickly and without allocating
// the declared canvas.
func TestStandardizeRejectsOversizedHeaderBeforeDecode(t *testing.T) {
	// 20000x20000 would need 1.6 GiB as NRGBA while the file itself stays tiny.
	content := bombPNG(t, 20000, 20000)
	if len(content) > 4096 {
		t.Fatalf("expected a tiny bomb fixture, got %d bytes", len(content))
	}

	started := time.Now()
	_, err := standardizeImage(context.Background(), content, "image/png", "bomb.png")
	elapsed := time.Since(started)

	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != CodePhotoTooLarge.RuntimeCode() {
		t.Fatalf("expected CodePhotoTooLarge, got %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("rejection took %v, which suggests the image was decoded first", elapsed)
	}
}

// TestStandardizeAcceptsCameraSizedPhoto verifies a normal camera photo is scaled
// down once and encoded within the stored object budget.
func TestStandardizeAcceptsCameraSizedPhoto(t *testing.T) {
	out, err := standardizeImage(context.Background(), photoBytes(t, 4000, 3000), "image/jpeg", "photo.jpg")
	if err != nil {
		t.Fatalf("standardize failed: %v", err)
	}
	if int64(len(out)) > maxPhotoBytes {
		t.Fatalf("standardized object is %d bytes, over the %d byte budget", len(out), maxPhotoBytes)
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("standardized object is not decodable: %v", err)
	}
	if maxInt(config.Width, config.Height) > targetLongEdge {
		t.Fatalf("expected the long edge scaled to %d, got %dx%d", targetLongEdge, config.Width, config.Height)
	}
}

// TestStandardizeRejectsUndecodableInput verifies non-image payloads are rejected
// as invalid rather than reaching the transcoder.
func TestStandardizeRejectsUndecodableInput(t *testing.T) {
	_, err := standardizeImage(context.Background(), []byte("not an image at all"), "image/png", "x.png")
	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != CodePhotoInvalid.RuntimeCode() {
		t.Fatalf("expected CodePhotoInvalid, got %v", err)
	}
}

// TestScaleToLongEdgeLeavesSmallImages verifies images already inside the budget
// skip the resize entirely, so small uploads pay no scaling cost.
func TestScaleToLongEdgeLeavesSmallImages(t *testing.T) {
	small := image.NewNRGBA(image.Rect(0, 0, 800, 600))
	if scaled := scaleToLongEdge(small, targetLongEdge); scaled != image.Image(small) {
		t.Fatal("expected an image within the budget to be returned untouched")
	}
	large := image.NewNRGBA(image.Rect(0, 0, 4000, 3000))
	scaled := scaleToLongEdge(large, targetLongEdge)
	if got := maxInt(scaled.Bounds().Dx(), scaled.Bounds().Dy()); got != targetLongEdge {
		t.Fatalf("expected the long edge scaled to %d, got %d", targetLongEdge, got)
	}
	if ratio := float64(scaled.Bounds().Dx()) / float64(scaled.Bounds().Dy()); math.Abs(ratio-4.0/3.0) > 0.01 {
		t.Fatalf("expected the aspect ratio preserved, got %.3f", ratio)
	}
}

// TestTranscodeSlotRejectsWhenContextDone verifies a request whose context is
// already cancelled is rejected with the busy code instead of queueing.
func TestTranscodeSlotRejectsWhenContextDone(t *testing.T) {
	for i := 0; i < maxConcurrentTranscodes; i++ {
		transcodeSlots <- struct{}{}
	}
	t.Cleanup(func() {
		for i := 0; i < maxConcurrentTranscodes; i++ {
			<-transcodeSlots
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	release, err := acquireTranscodeSlot(ctx)
	if release != nil {
		release()
	}
	bizErr, ok := bizerr.As(err)
	if !ok || bizErr.RuntimeCode() != CodePhotoBusy.RuntimeCode() {
		t.Fatalf("expected CodePhotoBusy, got %v", err)
	}
}
