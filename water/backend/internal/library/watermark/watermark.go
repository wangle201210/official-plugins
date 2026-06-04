package watermark

/*
#cgo amd64,linux LDFLAGS: -L. -lamd64_watermark
#cgo amd64,darwin LDFLAGS: -L. -lamd64_watermark
#cgo arm64,linux LDFLAGS: -L. -larm64_watermark
#cgo arm64,darwin LDFLAGS: -L. -larm64_watermark
#cgo pkg-config: libavformat libavcodec libavutil libavfilter libswscale x264

#include <stdlib.h>
#include "watermark.h"
*/
import "C"
import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"log"
	"strconv"
	"strings"
	"sync"
	"unsafe"
)

const (
	textFilterDescrFormat    = "drawtext=text='%s':x=%s:y=%s:fontsize=%d:fontcolor=%s%s"
	imageFilterDescrFormat   = "movie=%s,scale=%d:%d,lut=a='val*%f'"
	overlayFilterDescrFormat = "[in][layer%d]overlay=x=0:y=0[out]"
)

// 缓冲池，复用内存减少GC压力
var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 2*1024*1024) // 2MB初始缓冲区，专为7MB图片优化
	},
}

func makeFilterDescr(bounds image.Rectangle, config WatermarkConfig) string {
	filterParts := make([]string, 0)

	width := bounds.Dx()
	height := bounds.Dy()

	// 第一步：处理图片水印（如果有）
	currentLabel := "watermark"
	if config.ImageSetting.Image != "" && width > 0 && height > 0 {
		// 确保使用绝对路径或相对路径正确
		imagePath := config.ImageSetting.Image
		imageFilter := fmt.Sprintf(imageFilterDescrFormat,
			imagePath, width, height, config.ImageSetting.Opacity)
		filterParts = append(filterParts, fmt.Sprintf("%s[%s]", imageFilter, currentLabel))
	} else {
		// 如果没有图片水印，创建一个透明背景
		// 使用 color 滤镜创建透明背景
		filterParts = append(filterParts, fmt.Sprintf("color=c=black@0:size=%dx%d[%s]", width, height, currentLabel))
	}

	// 第二步：在图片水印上叠加文本水印
	if config.TextSetting.Text != "" {
		var fontDescr string
		if config.TextSetting.FontSize < 8 {
			config.TextSetting.FontSize = 8
		}

		if config.TextSetting.Font != "" {
			fontDescr = fmt.Sprintf(":fontfile=%s", config.TextSetting.Font)
		}

		if strings.Contains(config.TextSetting.Text, "\n") {
			lines := strings.Split(config.TextSetting.Text, "\n")
			for i, line := range lines {
				var x, y string
				if config.TextSetting.PosX > 0 && config.TextSetting.PosY > 0 {
					// 使用绝对位置，行间距为字体大小的 1.5 倍
					lineSpacing := int(float64(config.TextSetting.FontSize) * 1.5)
					x = strconv.Itoa(config.TextSetting.PosX)
					y = strconv.Itoa(config.TextSetting.PosY + i*lineSpacing)
				} else {
					// 使用对齐方式计算位置
					x, y = calculateTextPosition(i, len(lines),
						config.TextSetting.FontSize, config.TextSetting.Align)
				}

				nextLabel := fmt.Sprintf("txt%d", i+1)
				descr := fmt.Sprintf(textFilterDescrFormat,
					line, x, y, config.TextSetting.FontSize, config.TextSetting.Color, fontDescr)
				filterParts = append(filterParts, fmt.Sprintf("[%s]%s[%s]", currentLabel, descr, nextLabel))
				currentLabel = nextLabel
			}
		} else {
			var x, y string
			if config.TextSetting.PosX > 0 && config.TextSetting.PosY > 0 {
				x = strconv.Itoa(config.TextSetting.PosX)
				y = strconv.Itoa(config.TextSetting.PosY)
			} else {
				x, y = calculateTextPosition(0, 1,
					config.TextSetting.FontSize, config.TextSetting.Align)
			}

			nextLabel := "txt1"
			descr := fmt.Sprintf(textFilterDescrFormat,
				config.TextSetting.Text, x, y, config.TextSetting.FontSize,
				config.TextSetting.Color, fontDescr)
			filterParts = append(filterParts, fmt.Sprintf("[%s]%s[%s]", currentLabel, descr, nextLabel))
			currentLabel = nextLabel
		}
	}

	// 第三步：将处理后的水印叠加到原图上
	// 注意：overlay 的第一个输入是主图，第二个输入是叠加层
	// 如果水印尺寸与原图不同，overlay 会自动处理
	filterParts = append(filterParts, fmt.Sprintf("[in][%s]overlay=x=0:y=0[out]", currentLabel))

	return strings.Join(filterParts, ";")
}

func calculateTextPosition(idx int, total int, fontSize int, align Alignment) (string, string) {
	var x, y string
	// 行间距：字体大小的 1.5 倍，确保文字不重叠
	lineSpacing := int(float64(fontSize) * 1.5)

	switch align {
	case AlignmentLeft:
		// 居中垂直，然后根据行索引调整
		centerOffset := lineSpacing * (total - 1) / 2
		y = fmt.Sprintf("(h-th-20)/2 - %d + %d", centerOffset, idx*lineSpacing)
	case AlignmentRight:
		centerOffset := lineSpacing * (total - 1) / 2
		y = fmt.Sprintf("(h-th-20)/2 - %d + %d", centerOffset, idx*lineSpacing)
		x = "w-tw-10"
	case AlignmentCenter:
		centerOffset := lineSpacing * (total - 1) / 2
		y = fmt.Sprintf("(h-th-20)/2 - %d + %d", centerOffset, idx*lineSpacing)
		x = "(w-tw-20)/2"
	case AlignmentTop:
		y = fmt.Sprintf("10 + %d", idx*lineSpacing)
		x = "(w-tw-20)/2"
	case AlignmentBottom:
		// 从底部向上排列
		y = fmt.Sprintf("h-th-20 - %d", (total-idx-1)*lineSpacing)
		x = "(w-tw-20)/2"
	case AlignmentTopLeft:
		y = fmt.Sprintf("10 + %d", idx*lineSpacing)
		x = "10"
	case AlignmentTopRight:
		y = fmt.Sprintf("10 + %d", idx*lineSpacing)
		x = "w-tw-10"
	case AlignmentBottomLeft:
		y = fmt.Sprintf("h-th-20 - %d", (total-idx-1)*lineSpacing)
		x = "10"
	case AlignmentBottomRight:
		y = fmt.Sprintf("h-th-20 - %d", (total-idx-1)*lineSpacing)
		x = "w-tw-10"
	default:
		x = "10"
		y = fmt.Sprintf("10 + %d", idx*lineSpacing)
	}
	return x, y
}

func DrawWatermark(ctx context.Context, input []byte, config WatermarkConfig) (output []byte, err error) {
	// 从 JPEG 字节数组中解码获取图片尺寸（用于生成 filter 描述）
	img, err := png.Decode(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("failed to decode input JPEG: %v", err)
	}
	bounds := img.Bounds()
	// x := g.Cfg().MustGet(ctx, "tieta.snapX", 1920).Int()
	// y := g.Cfg().MustGet(ctx, "tieta.snapY", 1080).Int()
	// bounds := image.Rectangle{
	// 	Min: image.Point{0, 0},
	// 	Max: image.Point{x, y},
	// }
	inputSize := len(input)

	// 创建过滤器描述
	filterDescr := makeFilterDescr(bounds, config)
	log.Printf("filterDescr: %s\n", filterDescr)
	// 智能分配输出缓冲区大小
	// 从缓冲池获取缓冲区，减少内存分配
	outputBuf := bufferPool.Get().([]byte)
	defer func() {
		// 重置长度但保留容量，放回缓冲池
		outputBuf = outputBuf[:0]
		bufferPool.Put(outputBuf)
	}()

	// 恢复原有的内存分配逻辑，但使用缓冲池
	outputSize := inputSize * 2
	if outputSize > 50*1024*1024 {
		outputSize = 50 * 1024 * 1024 // 最大 50MB
	}

	// 确保缓冲区足够大，如果不够则重新分配
	if cap(outputBuf) < outputSize {
		outputBuf = make([]byte, 0, outputSize)
	}
	outputBuf = outputBuf[:outputSize]

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

	// 复制实际输出数据到新切片，避免缓冲池污染
	actualOutput := make([]byte, cOutputSize)
	copy(actualOutput, outputBuf[:cOutputSize])
	return actualOutput, nil
}
