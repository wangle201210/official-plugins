//go:build cgo

// This file adapts the migrated HotGo watermark library for LinaPro water inputs.

package watermark

/*
#include <stdlib.h>
#include "watermark.h"
*/
import "C"
import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"unsafe"
)

const (
	// minWatermarkOutputSize keeps small JPEG inputs from under-allocating the
	// output buffer after text and overlay filters increase encoded size.
	minWatermarkOutputSize = 2 * 1024 * 1024
	// maxWatermarkOutputSize preserves the legacy upper bound used by the
	// migrated HotGo adapter to avoid unbounded FFmpeg output allocation.
	maxWatermarkOutputSize = 50 * 1024 * 1024
)

// DrawWatermarkJpeg invokes the migrated HotGo FFmpeg/C watermark pipeline for JPEG bytes.
func DrawWatermarkJpeg(_ context.Context, input []byte, config WatermarkConfig) ([]byte, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("watermark input is empty")
	}
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode input image: %v", err)
	}
	config, err = withDefaultFont(config)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	inputSize := len(input)
	filterDescr := makeFilterDescr(bounds, config)

	outputBuf := bufferPool.Get().([]byte)
	defer func() {
		outputBuf = outputBuf[:0]
		bufferPool.Put(outputBuf)
	}()

	outputSize := watermarkOutputBufferSize(inputSize, bounds)
	if cap(outputBuf) < outputSize {
		outputBuf = make([]byte, 0, outputSize)
	}
	outputBuf = outputBuf[:outputSize]

	cInput := (*C.uchar)(unsafe.Pointer(&input[0]))
	cFilterDescr := C.CString(filterDescr)
	defer C.free(unsafe.Pointer(cFilterDescr))
	cOutput := (*C.uchar)(unsafe.Pointer(&outputBuf[0]))
	cOutputSize := C.int(outputSize)

	ret := C.process_jpg_watermark(cInput, C.int(inputSize), cFilterDescr, cOutput, &cOutputSize)
	if ret != 0 {
		return nil, fmt.Errorf("process_jpg_watermark failed with code: %d (filter: %s)", ret, filterDescr)
	}
	if int(cOutputSize) <= 0 {
		return nil, fmt.Errorf("process_jpg_watermark returned zero output size")
	}
	actualOutput := make([]byte, cOutputSize)
	copy(actualOutput, outputBuf[:cOutputSize])
	return actualOutput, nil
}

// watermarkOutputBufferSize estimates the FFmpeg output buffer size from both
// encoded input size and pixel count because tiny highly-compressed inputs can
// expand after watermark filters are applied.
func watermarkOutputBufferSize(inputSize int, bounds image.Rectangle) int {
	size := inputSize * 4
	if inputSize > maxWatermarkOutputSize/4 {
		size = maxWatermarkOutputSize
	}
	pixelSize := int64(bounds.Dx()) * int64(bounds.Dy()) * 4
	if pixelSize > int64(size) {
		if pixelSize > int64(maxWatermarkOutputSize) {
			size = maxWatermarkOutputSize
		} else {
			size = int(pixelSize)
		}
	}
	if size < minWatermarkOutputSize {
		size = minWatermarkOutputSize
	}
	if size > maxWatermarkOutputSize {
		size = maxWatermarkOutputSize
	}
	return size
}
