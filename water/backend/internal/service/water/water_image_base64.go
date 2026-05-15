// This file implements image base64 and PNG data URL helpers.

package water

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/crypto/gmd5"

	"lina-core/pkg/bizerr"
)

// decodeImageDataURL decodes base64 image text with or without a data URL prefix.
func decodeImageDataURL(value string) ([]byte, error) {
	clean := strings.TrimSpace(value)
	if clean == "" {
		return nil, bizerr.NewCode(CodeWaterImageRequired)
	}
	if strings.HasPrefix(clean, "data:image/") {
		comma := strings.Index(clean, ",")
		if comma < 0 {
			return nil, bizerr.NewCode(CodeWaterImageBase64Invalid)
		}
		clean = clean[comma+1:]
	}
	decoded, err := base64.StdEncoding.DecodeString(clean)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(clean)
	}
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterImageBase64Invalid)
	}
	return decoded, nil
}

// encodePNGDataURL encodes image bytes with HotGo's data URL prefix.
func encodePNGDataURL(img []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(img)
}

// ensurePNGDataURL converts image bytes to PNG data URL.
func ensurePNGDataURL(input []byte) (string, error) {
	img, _, err := image.Decode(bytes.NewReader(input))
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageDecodeFailed)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageEncodeFailed)
	}
	return encodePNGDataURL(buf.Bytes()), nil
}

// base64ToMD5Pic stores a base64 image in the same MD5 path layout used by HotGo.
func base64ToMD5Pic(base64Image string, targetDir string) (string, error) {
	cleanBase64 := strings.TrimSpace(base64Image)
	if cleanBase64 == "" {
		return "", bizerr.NewCode(CodeWaterImageRequired)
	}
	if strings.HasPrefix(cleanBase64, "data:image/") {
		commaIndex := strings.Index(cleanBase64, ",")
		if commaIndex != -1 {
			cleanBase64 = cleanBase64[commaIndex+1:]
		}
	}
	hash, err := gmd5.EncryptString(cleanBase64)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageBase64Invalid)
	}
	dirPath := "/tmp/media/pic/md5"
	if strings.TrimSpace(targetDir) != "" {
		dirPath = targetDir
	}
	path := filepath.Join(dirPath, fmt.Sprintf("%s.png", hash))
	if _, err = os.Stat(path); err == nil {
		return path, nil
	}
	if err = os.MkdirAll(dirPath, 0755); err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageEncodeFailed)
	}
	imgData, err := base64.StdEncoding.DecodeString(cleanBase64)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageBase64Invalid)
	}
	if err = os.WriteFile(path, imgData, 0644); err != nil {
		return "", bizerr.WrapCode(err, CodeWaterImageEncodeFailed)
	}
	return path, nil
}
