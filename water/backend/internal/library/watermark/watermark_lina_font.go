//go:build cgo

// This file provides LinaPro's default CJK font for the migrated watermark renderer.

package watermark

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	// defaultChineseFontEmbedPath is the embedded HotGo font asset path.
	defaultChineseFontEmbedPath = "fonts/STHeiti Medium.ttc"
	// defaultChineseFontFileName is the materialized font filename passed to FFmpeg.
	defaultChineseFontFileName = "STHeiti-Medium.ttc"
)

var (
	//go:embed "fonts/STHeiti Medium.ttc"
	defaultChineseFontFS embed.FS

	defaultChineseFontMu   sync.Mutex
	defaultChineseFontPath string
)

// withDefaultFont fills the default CJK-capable font when a text watermark omits font.
func withDefaultFont(config WatermarkConfig) (WatermarkConfig, error) {
	if strings.TrimSpace(config.TextSetting.Text) == "" || strings.TrimSpace(config.TextSetting.Font) != "" {
		return config, nil
	}
	fontPath, err := DefaultFontPath()
	if err != nil {
		return config, err
	}
	config.TextSetting.Font = fontPath
	return config, nil
}

// DefaultFontPath materializes and returns the embedded CJK-capable watermark font path.
func DefaultFontPath() (string, error) {
	defaultChineseFontMu.Lock()
	defer defaultChineseFontMu.Unlock()

	if defaultChineseFontPath != "" {
		if _, err := os.Stat(defaultChineseFontPath); err == nil {
			return defaultChineseFontPath, nil
		}
	}

	fontPath, err := materializeDefaultChineseFont()
	if err != nil {
		return "", err
	}
	defaultChineseFontPath = fontPath
	return defaultChineseFontPath, nil
}

// materializeDefaultChineseFont writes the embedded font to a stable temporary path for FFmpeg.
func materializeDefaultChineseFont() (string, error) {
	fontBytes, err := defaultChineseFontFS.ReadFile(defaultChineseFontEmbedPath)
	if err != nil {
		return "", fmt.Errorf("read embedded watermark font: %w", err)
	}

	fontDir := filepath.Join(os.TempDir(), "media", "fonts")
	if err := os.MkdirAll(fontDir, 0o755); err != nil {
		return "", fmt.Errorf("create watermark font directory: %w", err)
	}

	fontPath := filepath.Join(fontDir, defaultChineseFontFileName)
	if existing, err := os.ReadFile(fontPath); err == nil && bytes.Equal(existing, fontBytes) {
		return fontPath, nil
	}
	if err := os.WriteFile(fontPath, fontBytes, 0o644); err != nil {
		return "", fmt.Errorf("write watermark font: %w", err)
	}
	return fontPath, nil
}
