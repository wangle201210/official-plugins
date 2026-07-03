package watermark

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// cjkGlyphProbeText is used to select a TTC sub-font that actually covers Chinese glyphs.
const cjkGlyphProbeText = "中文测试汉字"

var darwinFontPaths = []string{
	"/System/Library/Fonts/PingFang.ttc",
	"/System/Library/Fonts/STHeiti Light.ttc",
	"/System/Library/Fonts/STHeiti Medium.ttc",
	"/Library/Fonts/Arial Unicode.ttf",
}

var linuxFontPaths = []string{
	// Ubuntu: fonts-noto-cjk
	"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
	"/usr/share/fonts/opentype/noto/NotoSansCJK-VF.ttf",
	"/usr/share/fonts/truetype/noto/NotoSansCJK-Regular.ttc",
	"/usr/share/fonts/opentype/noto/NotoSerifCJK-Regular.ttc",
	// Ubuntu: fonts-wqy-*
	"/usr/share/fonts/truetype/wqy/wqy-microhei.ttc",
	"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
	// Ubuntu: legacy CJK fallbacks
	"/usr/share/fonts/truetype/droid/DroidSansFallbackFull.ttf",
	"/usr/share/fonts/truetype/arphic/uming.ttc",
	"/usr/share/fonts/truetype/arphic/ukai.ttc",
	// Ubuntu: fonts-ubuntu / general sans
	"/usr/share/fonts/truetype/ubuntu/Ubuntu-R.ttf",
	"/usr/share/fonts/truetype/ubuntu/Ubuntu[wdth,wght].ttf",
	"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
	"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
	// Adobe Source Han (manual installs)
	"/usr/share/fonts/opentype/adobe-source-han-sans/SourceHanSansCN-Regular.otf",
	"/usr/share/fonts/truetype/adobe-source-han-sans/SourceHanSansCN-Regular.otf",
	"/usr/local/share/fonts/NotoSansCJK-Regular.ttc",
}

func parseTextColor(colorStr string) (color.RGBA, error) {
	colorStr = strings.TrimSpace(colorStr)
	if colorStr == "" {
		return color.RGBA{255, 255, 255, 255}, nil
	}

	switch strings.ToLower(colorStr) {
	case "white":
		return color.RGBA{255, 255, 255, 255}, nil
	case "black":
		return color.RGBA{0, 0, 0, 255}, nil
	case "red":
		return color.RGBA{255, 0, 0, 255}, nil
	case "green":
		return color.RGBA{0, 255, 0, 255}, nil
	case "blue":
		return color.RGBA{0, 0, 255, 255}, nil
	case "yellow":
		return color.RGBA{255, 255, 0, 255}, nil
	case "gray", "grey":
		return color.RGBA{128, 128, 128, 255}, nil
	}

	hex := colorStr
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = "#" + hex[2:]
	}
	if !strings.HasPrefix(hex, "#") {
		hex = "#" + hex
	}

	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 && len(hex) != 8 {
		return color.RGBA{}, fmt.Errorf("invalid color: %s", colorStr)
	}

	value, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return color.RGBA{}, fmt.Errorf("invalid color: %s", colorStr)
	}

	if len(hex) == 6 {
		return color.RGBA{
			R: uint8(value >> 16),
			G: uint8(value >> 8),
			B: uint8(value),
			A: 255,
		}, nil
	}

	return color.RGBA{
		R: uint8(value >> 24),
		G: uint8(value >> 16),
		B: uint8(value >> 8),
		A: uint8(value),
	}, nil
}

func loadFontFaceFromFile(fontPath string, fontSize float64) (font.Face, error) {
	fontData, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	face, trueTypeErr := loadTrueTypeFontFace(fontData, fontSize)
	if trueTypeErr == nil {
		return face, nil
	}
	face, openTypeErr := loadOpenTypeFontFace(fontData, fontSize)
	if openTypeErr == nil {
		return face, nil
	}
	return nil, errors.Join(trueTypeErr, openTypeErr)
}

func loadOpenTypeFontFace(fontData []byte, fontSize float64) (font.Face, error) {
	var otFont *opentype.Font
	if collection, err := opentype.ParseCollection(fontData); err == nil {
		var err error
		otFont, err = collection.Font(0)
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		otFont, err = opentype.Parse(fontData)
		if err != nil {
			return nil, err
		}
	}

	return opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func loadTrueTypeFontFace(fontData []byte, fontSize float64) (font.Face, error) {
	ttFont, err := loadTrueTypeFont(fontData)
	if err != nil {
		return nil, err
	}
	return truetype.NewFace(ttFont, &truetype.Options{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	}), nil
}

// loadTrueTypeFont loads a single TTF or selects a CJK-capable face from a TTC.
func loadTrueTypeFont(fontData []byte) (*truetype.Font, error) {
	offsets, ok := trueTypeCollectionOffsets(fontData)
	if !ok {
		return truetype.Parse(fontData)
	}

	var (
		bestFont  *truetype.Font
		bestScore = -1
		firstErr  error
	)
	for _, offset := range offsets {
		if int(offset) >= len(fontData) {
			if firstErr == nil {
				firstErr = fmt.Errorf("trueType collection font offset %d out of range", offset)
			}
			continue
		}
		extractedFont, err := extractTrueTypeCollectionFont(fontData, offset)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		candidate, err := truetype.Parse(extractedFont)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		score := trueTypeGlyphCoverageScore(candidate, cjkGlyphProbeText)
		if score > bestScore {
			bestFont = candidate
			bestScore = score
		}
		if score == len([]rune(cjkGlyphProbeText)) {
			return candidate, nil
		}
	}
	if bestFont != nil {
		return bestFont, nil
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return nil, fmt.Errorf("trueType collection has no fonts")
}

// trueTypeCollectionOffsets returns embedded font offsets from a TTC header.
func trueTypeCollectionOffsets(fontData []byte) ([]uint32, bool) {
	const ttcHeaderSize = 12
	if len(fontData) < ttcHeaderSize || string(fontData[:4]) != "ttcf" {
		return nil, false
	}
	fontCount := binary.BigEndian.Uint32(fontData[8:12])
	offsetTableEnd := ttcHeaderSize + int(fontCount)*4
	if fontCount == 0 || offsetTableEnd > len(fontData) {
		return nil, false
	}
	offsets := make([]uint32, 0, fontCount)
	for offsetPos := ttcHeaderSize; offsetPos < offsetTableEnd; offsetPos += 4 {
		offsets = append(offsets, binary.BigEndian.Uint32(fontData[offsetPos:offsetPos+4]))
	}
	return offsets, true
}

// extractTrueTypeCollectionFont rewrites one TTC sub-font into standalone TTF bytes.
func extractTrueTypeCollectionFont(fontData []byte, fontOffset uint32) ([]byte, error) {
	const tableDirectorySize = 12
	start := int(fontOffset)
	if start < 0 || start+tableDirectorySize > len(fontData) {
		return nil, fmt.Errorf("trueType collection font offset %d out of range", fontOffset)
	}

	numTables := int(binary.BigEndian.Uint16(fontData[start+4 : start+6]))
	if numTables == 0 {
		return nil, fmt.Errorf("trueType collection font at offset %d has no tables", fontOffset)
	}
	directoryLen := tableDirectorySize + numTables*16
	if start+directoryLen > len(fontData) {
		return nil, fmt.Errorf("trueType collection font directory at offset %d out of range", fontOffset)
	}

	tableDataOffset := align4(directoryLen)
	tableData := make([]byte, tableDataOffset)
	copy(tableData[:directoryLen], fontData[start:start+directoryLen])

	for i := 0; i < numTables; i++ {
		recordPos := tableDirectorySize + i*16
		sourceOffset := binary.BigEndian.Uint32(tableData[recordPos+8 : recordPos+12])
		length := binary.BigEndian.Uint32(tableData[recordPos+12 : recordPos+16])
		sourceStart := int(sourceOffset)
		sourceEnd := sourceStart + int(length)
		if sourceStart < 0 || sourceEnd < sourceStart || sourceEnd > len(fontData) {
			return nil, fmt.Errorf("trueType table %d range out of collection bounds", i)
		}

		destinationOffset := align4(len(tableData))
		if destinationOffset > len(tableData) {
			tableData = append(tableData, make([]byte, destinationOffset-len(tableData))...)
		}
		binary.BigEndian.PutUint32(tableData[recordPos+8:recordPos+12], uint32(destinationOffset))
		tableData = append(tableData, fontData[sourceStart:sourceEnd]...)
	}

	return tableData, nil
}

func align4(value int) int {
	return (value + 3) &^ 3
}

// trueTypeGlyphCoverageScore counts how many probe runes resolve to real glyphs.
func trueTypeGlyphCoverageScore(ttFont *truetype.Font, probe string) int {
	score := 0
	for _, r := range probe {
		if ttFont.Index(r) != 0 {
			score++
		}
	}
	return score
}

func discoverLinuxFonts() []string {
	if runtime.GOOS != "linux" {
		return nil
	}

	roots := []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	keywords := []string{
		"notosanscjk", "notoserifcjk", "wqy", "droidsansfallback",
		"uming", "ukai", "sourcehansans", "sourcehanserif",
	}

	found := make([]string, 0)
	seen := make(map[string]struct{})

	for _, root := range roots {
		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".ttf" && ext != ".ttc" && ext != ".otf" {
				return nil
			}

			base := strings.ToLower(filepath.Base(path))
			for _, keyword := range keywords {
				if strings.Contains(base, keyword) {
					if _, ok := seen[path]; !ok {
						seen[path] = struct{}{}
						found = append(found, path)
					}
					break
				}
			}
			return nil
		})
	}

	return found
}

func systemFontPaths() []string {
	paths := make([]string, 0)
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths, darwinFontPaths...)
	case "linux":
		paths = append(paths, linuxFontPaths...)
		paths = append(paths, discoverLinuxFonts()...)
	default:
		paths = append(paths, linuxFontPaths...)
		paths = append(paths, darwinFontPaths...)
	}
	return paths
}

func loadEmbeddedFontFace(fontSize float64) (font.Face, error) {
	otFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(otFont, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func loadDefaultFontFace(fontSize float64) (font.Face, error) {
	for _, path := range systemFontPaths() {
		face, err := loadFontFaceFromFile(path, fontSize)
		if err == nil {
			return face, nil
		}
	}
	return loadEmbeddedFontFace(fontSize)
}

func loadFontFace(fontPath string, fontSize float64) (font.Face, error) {
	fontPath = strings.TrimSpace(fontPath)
	if fontPath != "" {
		face, err := loadFontFaceFromFile(fontPath, fontSize)
		if err == nil {
			return face, nil
		}
		return nil, fmt.Errorf("load configured font %q: %w", fontPath, err)
	}
	return loadDefaultFontFace(fontSize)
}

func measureTextBlock(face font.Face, lines []string) (int, int) {
	drawer := &font.Drawer{Face: face}
	maxWidth := 0
	for _, line := range lines {
		if line == "" {
			continue
		}
		advance := drawer.MeasureString(line)
		width := (advance + fixed.I(1) - 1).Ceil()
		if width > maxWidth {
			maxWidth = width
		}
	}

	lineHeight := face.Metrics().Height.Ceil()
	if lineHeight <= 0 {
		lineHeight = int(face.Metrics().Ascent.Ceil() + face.Metrics().Descent.Ceil())
	}
	if lineHeight <= 0 {
		lineHeight = 1
	}

	blockHeight := lineHeight * len(lines)
	if blockHeight <= 0 {
		blockHeight = lineHeight
	}

	return maxWidth, blockHeight
}

func drawTextTile(face font.Face, lines []string, textColor color.RGBA, blockWidth, blockHeight int) *image.RGBA {
	tile := image.NewRGBA(image.Rect(0, 0, blockWidth, blockHeight))
	drawer := &font.Drawer{
		Dst:  tile,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	lineHeight := face.Metrics().Height.Ceil()
	if lineHeight <= 0 {
		lineHeight = 1
	}
	ascent := face.Metrics().Ascent.Ceil()

	for i, line := range lines {
		if line == "" {
			continue
		}
		drawer.Dot = fixed.Point26_6{
			X: fixed.I(0),
			Y: fixed.I(ascent + i*lineHeight),
		}
		drawer.DrawString(line)
	}

	return tile
}

func rotateRGBA(src *image.RGBA, angleDeg int) *image.RGBA {
	angleDeg = angleDeg % 360
	if angleDeg < 0 {
		angleDeg += 360
	}
	if angleDeg == 0 {
		return src
	}

	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	cx := float64(w-1) / 2
	cy := float64(h-1) / 2

	rad := float64(angleDeg) * math.Pi / 180
	cosA := math.Cos(rad)
	sinA := math.Sin(rad)

	corners := [][2]float64{{0, 0}, {float64(w - 1), 0}, {float64(w - 1), float64(h - 1)}, {0, float64(h - 1)}}
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64
	for _, p := range corners {
		rx, ry := rotatePoint(p[0], p[1], cx, cy, cosA, sinA)
		minX = math.Min(minX, rx)
		minY = math.Min(minY, ry)
		maxX = math.Max(maxX, rx)
		maxY = math.Max(maxY, ry)
	}

	dstW := int(math.Ceil(maxX-minX)) + 1
	dstH := int(math.Ceil(maxY-minY)) + 1
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	ncx := cx - minX
	ncy := cy - minY

	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			sx := (float64(x)-ncx)*cosA + (float64(y)-ncy)*sinA + cx
			sy := -(float64(x)-ncx)*sinA + (float64(y)-ncy)*cosA + cy
			if sx < 0 || sy < 0 || sx >= float64(w) || sy >= float64(h) {
				continue
			}
			dst.Set(x, y, src.At(int(sx+0.5), int(sy+0.5)))
		}
	}

	return dst
}

func rotatePoint(x, y, cx, cy, cosA, sinA float64) (float64, float64) {
	dx, dy := x-cx, y-cy
	return cx + dx*cosA - dy*sinA, cy + dx*sinA + dy*cosA
}

func blendOver(dst *image.RGBA, src *image.RGBA, x, y int) {
	srcBounds := src.Bounds()
	for sy := 0; sy < srcBounds.Dy(); sy++ {
		for sx := 0; sx < srcBounds.Dx(); sx++ {
			_, _, _, a := src.At(sx, sy).RGBA()
			if a == 0 {
				continue
			}
			dx, dy := x+sx, y+sy
			if dx < 0 || dy < 0 || dx >= dst.Bounds().Dx() || dy >= dst.Bounds().Dy() {
				continue
			}
			dst.Set(dx, dy, src.At(sx, sy))
		}
	}
}

func generateTiledTextImage(text, fontPath string, fontSize int, colorStr string, width, height, rotate int, outputPath string) error {
	if text == "" {
		return fmt.Errorf("text is empty")
	}
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid image size: %dx%d", width, height)
	}
	if fontSize < 8 {
		fontSize = 8
	}

	textColor, err := parseTextColor(colorStr)
	if err != nil {
		return err
	}
	textColor.A = 255

	face, err := loadFontFace(fontPath, float64(fontSize))
	if err != nil {
		return fmt.Errorf("load watermark font: %w", err)
	}
	defer face.Close()

	lines := strings.Split(text, "\n")
	if len(lines) == 0 {
		lines = []string{text}
	}

	blockWidth, blockHeight := measureTextBlock(face, lines)
	if blockWidth <= 0 {
		blockWidth = fontSize
	}

	tile := drawTextTile(face, lines, textColor, blockWidth, blockHeight)
	rotatedTile := rotateRGBA(tile, rotate)
	tileW := rotatedTile.Bounds().Dx()
	tileH := rotatedTile.Bounds().Dy()

	gapX := tileW / 2
	gapY := tileH / 2
	if gapX < fontSize/2 {
		gapX = fontSize / 2
	}
	if gapY < fontSize/2 {
		gapY = fontSize / 2
	}
	stepX := tileW + gapX
	stepY := tileH + gapY
	if stepX <= 0 {
		stepX = 1
	}
	if stepY <= 0 {
		stepY = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y += stepY {
		for x := 0; x < width; x += stepX {
			blendOver(img, rotatedTile, x, y)
		}
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	return nil
}
