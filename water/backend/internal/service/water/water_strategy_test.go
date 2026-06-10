// This file tests watermark strategy parsing.

package water

import "testing"

// TestParseWatermarkStrategySnapshotNode verifies Lina snapshot watermark YAML parsing.
func TestParseWatermarkStrategySnapshotNode(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true
snapshot_watermark:
  text: 园区安防
  fontSize: 48
  color: "#00ff88"
  align: bottomRight
  opacity: 0.5`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected watermark config")
	}
	if cfg.Text != "园区安防" {
		t.Fatalf("expected text to roundtrip, got %q", cfg.Text)
	}
	if cfg.FontSize != 48 {
		t.Fatalf("expected font size 48, got %d", cfg.FontSize)
	}
	if normalizedAlignment(cfg.Align) != "bottomright" {
		t.Fatalf("expected bottomright alignment, got %q", normalizedAlignment(cfg.Align))
	}
}

// TestParseWatermarkStrategyNumericAlign verifies numeric alignment inside snapshot watermark YAML.
func TestParseWatermarkStrategyNumericAlign(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`snapshot_watermark:
  text: 热点水印
  fontSize: 64
  align: 9
  opacity: 0.15`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected snapshot watermark config")
	}
	if normalizedAlignment(cfg.Align) != "bottomright" {
		t.Fatalf("expected numeric alignment to map to bottomright, got %q", normalizedAlignment(cfg.Align))
	}
	if cfg.Opacity != 0.15 {
		t.Fatalf("expected opacity 0.15, got %f", cfg.Opacity)
	}
}

// TestParseWatermarkStrategyDefaultOpacity verifies omitted opacity uses the snapshot watermark default.
func TestParseWatermarkStrategyDefaultOpacity(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`snapshot_watermark:
  text: 默认透明度`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected snapshot watermark config")
	}
	if cfg.Opacity != 0.15 {
		t.Fatalf("expected default opacity 0.15, got %f", cfg.Opacity)
	}
}

// TestParseWatermarkStrategyMissing verifies strategies without snapshot watermark are skipped.
func TestParseWatermarkStrategyMissing(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil watermark config, got %+v", cfg)
	}
}

// TestParseWatermarkStrategyIgnoresGenericWatermark verifies only screenshot-specific nodes are accepted.
func TestParseWatermarkStrategyIgnoresGenericWatermark(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`watermark:
  enabled: true
  text: 通用水印`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected generic watermark node to be ignored, got %+v", cfg)
	}
}
