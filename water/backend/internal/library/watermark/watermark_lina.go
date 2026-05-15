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

	outputSize := inputSize * 2
	if outputSize > 50*1024*1024 {
		outputSize = 50 * 1024 * 1024
	}
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
