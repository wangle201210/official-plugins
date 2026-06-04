//go:build !cgo

// This file provides the non-CGO fallback for the LinaPro watermark adapter.

package watermark

import (
	"context"
	"fmt"
)

// DrawWatermarkJpeg reports that the migrated HotGo C/FFmpeg pipeline is unavailable without CGO.
func DrawWatermarkJpeg(_ context.Context, _ []byte, _ WatermarkConfig) ([]byte, error) {
	return nil, fmt.Errorf("watermark cgo/FFmpeg library is unavailable because CGO is disabled")
}
