// This file builds FFmpeg watermark filter descriptions and declares cgo
// linkage for the migrated HotGo watermark static library.

package watermark

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lwatermark
#cgo pkg-config: libavformat libavcodec libavutil libavfilter libswscale x264

#include <stdlib.h>
#include <string.h>
#include "watermark.h"
*/
import "C"
import (
	"encoding/base64"
	"fmt"
	"image"
	"log"
	"os"
	"strings"
	"time"
	"unsafe"
)

const (
	textFilterDescrFormat    = "drawtext=text='%s':x=%s:y=%s:fontsize=%d:fontcolor=%s%s"
	imageFilterDescrFormat   = "movie=%s,scale=%d:%d,lut=a='val*%f'"
	overlayFilterDescrFormat = "[in][layer%d]overlay=x=0:y=0[out]"
)

func makeFilterDescr(bounds image.Rectangle, config WatermarkConfig, variables map[string]string) string {
	filterParts := make([]string, 0)

	width := bounds.Dx()
	height := bounds.Dy()

	// 第一步：处理图片水印（如果有）
	currentLabel := "watermark"
	if config.Base64 != "" && width > 0 && height > 0 {
		// 确保使用绝对路径或相对路径正确
		os.MkdirAll("./img", 0755)
		filename := "./img/" + MD5(config.Base64) + ".png"
		data, err := base64.StdEncoding.DecodeString(config.Base64)
		if err != nil {
			log.Printf("base64 decode error: %v\n", err)
			return ""
		}
		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			log.Printf("write file error: %v\n", err)
			return ""
		}
		imageFilter := fmt.Sprintf(imageFilterDescrFormat, filename, width, height, config.Opacity)
		filterParts = append(filterParts, fmt.Sprintf("%s[%s]", imageFilter, currentLabel))
	} else {
		text := config.Order
		text = strings.ReplaceAll(text, "{{date-1}}", time.Now().Format("2006-01-02"))
		text = strings.ReplaceAll(text, "{{date-2}}", time.Now().Format("2006-01-02 15:04:05"))
		for k, v := range variables {
			text = strings.ReplaceAll(text, "{{"+k+"}}", v)
		}
		if text != "" && width > 0 && height > 0 {
			os.MkdirAll("./img", 0755)
			filename := "./img/" + MD5(text+config.Font+fmt.Sprint(config.FontSize)+config.Color+fmt.Sprint(config.Rotate)) + ".png"
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				if err := generateTiledTextImage(text, config.Font, config.FontSize, config.Color, width, height, config.Rotate, filename); err != nil {
					log.Printf("generate tiled text image error: %v\n", err)
					return ""
				}
			} else if err != nil {
				log.Printf("stat image file error: %v\n", err)
				return ""
			}
			imageFilter := fmt.Sprintf(imageFilterDescrFormat, filename, width, height, config.Opacity)
			filterParts = append(filterParts, fmt.Sprintf("%s[%s]", imageFilter, currentLabel))
		}
	}

	if len(filterParts) > 0 {
		filterParts = append(filterParts, fmt.Sprintf("[in][%s]overlay=x=0:y=0[out]", currentLabel))
	} else {
		filterParts = append(filterParts, "[in]copy[out]")
	}

	return strings.Join(filterParts, ";")
}

func DrawWatermark(input []byte, config WatermarkConfig, variables map[string]string) (output []byte, err error) {
	bounds := image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{config.Width, config.Height},
	}
	inputSize := len(input)

	if variables == nil {
		variables = map[string]string{}
	}

	filterDescr := makeFilterDescr(bounds, config, variables)
	log.Printf("filterDescr: %s\n", filterDescr)
	// 分配输出缓冲区（初始大小为输入大小的 2 倍，以应对可能的增长）
	outputSize := inputSize * 2
	if outputSize > 50*1024*1024 {
		outputSize = 50 * 1024 * 1024 // 最大 50MB
	}
	outputBuf := make([]byte, outputSize)

	// 调用 C 函数
	cInput := (*C.uchar)(unsafe.Pointer(&input[0]))
	cFilterDescr := C.CString(filterDescr)
	defer C.free(unsafe.Pointer(cFilterDescr))
	cOutput := (*C.uchar)(unsafe.Pointer(&outputBuf[0]))
	cOutputSize := C.int(outputSize)

	ret := C.process_jpg_watermark(cInput, C.int(inputSize), cFilterDescr, cOutput, &cOutputSize)
	if ret != 0 {
		// 尝试获取 FFmpeg 错误信息
		return nil, fmt.Errorf("process_jpg_watermark failed with code: %d (filter: %s)", ret, filterDescr)
	}

	if int(cOutputSize) <= 0 {
		return nil, fmt.Errorf("process_jpg_watermark returned zero output size")
	}

	// 返回实际输出数据
	return outputBuf[:int(cOutputSize)], nil
}
