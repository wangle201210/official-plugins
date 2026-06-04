// This file resolves media strategies and parses watermark strategy YAML.

package water

import (
	"context"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	mediastrategy "lina-plugin-media/backend/provider/strategy"
	"lina-plugin-water/backend/internal/library/watermark"
)

// strategyYAML is a tolerant projection of media strategy YAML.
type strategyYAML struct {
	Watermark       *watermarkConfig `json:"watermark" yaml:"watermark"` // Watermark is the nested Lina strategy node.
	watermarkConfig `yaml:",inline"` // Inline fields keep hotgo root-level strategy compatibility.
}

// parseWatermarkStrategy parses watermark configuration from strategy YAML.
func parseWatermarkStrategy(strategyBody string) (*watermarkConfig, error) {
	body := strings.TrimSpace(strategyBody)
	if body == "" {
		return nil, nil
	}
	var parsed strategyYAML
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyParseFailed)
	}

	cfg := parsed.Watermark
	if cfg == nil && hasRootWatermarkConfig(parsed.watermarkConfig) {
		cfg = &parsed.watermarkConfig
	}
	if cfg == nil {
		return nil, nil
	}
	normalized, err := normalizeWatermarkConfig(*cfg)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

// resolveStrategy delegates effective media strategy lookup to the media plugin provider.
func (s *serviceImpl) resolveStrategy(ctx context.Context, tenantID string, deviceID string) (*resolvedStrategy, error) {
	if s == nil || s.strategyResolver == nil {
		return nil, bizerr.NewCode(CodeWaterMediaResolverUnavailable)
	}
	strategy, err := s.strategyResolver.ResolveStrategy(ctx, mediastrategy.ResolveStrategyInput{
		TenantId: strings.TrimSpace(tenantID),
		DeviceId: strings.TrimSpace(deviceID),
	})
	if err != nil {
		return nil, err
	}
	return buildResolvedStrategy(strategy), nil
}

// buildResolvedStrategy converts the media provider output into service output.
func buildResolvedStrategy(strategy *mediastrategy.ResolveStrategyOutput) *resolvedStrategy {
	source := StrategySourceNone
	if strategy != nil {
		source = normalizeStrategySource(strategy.Source)
	}
	sourceLabel := strategySourceLabel(source)
	if strategy != nil && strings.TrimSpace(strategy.SourceLabel) != "" {
		sourceLabel = strings.TrimSpace(strategy.SourceLabel)
	}
	out := &resolvedStrategy{
		Matched:     strategy != nil && strategy.Matched,
		Source:      source,
		SourceLabel: sourceLabel,
	}
	if strategy != nil {
		out.StrategyId = strategy.StrategyId
		out.StrategyName = strategy.StrategyName
		out.Strategy = strategy.Strategy
	}
	return out
}

// normalizeStrategySource constrains media source text to water's known enum values.
func normalizeStrategySource(source string) StrategySource {
	switch StrategySource(strings.TrimSpace(source)) {
	case StrategySourceTenantDevice:
		return StrategySourceTenantDevice
	case StrategySourceDevice:
		return StrategySourceDevice
	case StrategySourceTenant:
		return StrategySourceTenant
	case StrategySourceGlobal:
		return StrategySourceGlobal
	default:
		return StrategySourceNone
	}
}

// strategySourceLabel returns the Chinese label for one strategy source.
func strategySourceLabel(source StrategySource) string {
	switch source {
	case StrategySourceTenantDevice:
		return "租户设备策略"
	case StrategySourceDevice:
		return "设备策略"
	case StrategySourceTenant:
		return "租户策略"
	case StrategySourceGlobal:
		return "全局策略"
	default:
		return "未匹配"
	}
}

// hasRootWatermarkConfig reports whether the root YAML object looks like a hotgo watermark config.
func hasRootWatermarkConfig(cfg watermarkConfig) bool {
	return cfg.Enabled ||
		strings.TrimSpace(cfg.Text) != "" ||
		strings.TrimSpace(cfg.Image) != "" ||
		strings.TrimSpace(cfg.Base64) != "" ||
		cfg.FontSize > 0 ||
		cfg.Opacity > 0 ||
		cfg.PosX > 0 ||
		cfg.PosY > 0 ||
		cfg.Align != ""
}

// normalizeWatermarkConfig fills defaults and normalizes bounded fields.
func normalizeWatermarkConfig(cfg watermarkConfig) (watermarkConfig, error) {
	cfg.Text = strings.TrimSpace(cfg.Text)
	cfg.Font = strings.TrimSpace(cfg.Font)
	cfg.Color = strings.TrimSpace(cfg.Color)
	cfg.Image = strings.TrimSpace(cfg.Image)
	cfg.Base64 = strings.TrimSpace(cfg.Base64)
	cfg.Align = watermarkAlignment(strings.TrimSpace(string(cfg.Align)))
	if cfg.Base64 != "" && cfg.Image == "" {
		imagePath, err := base64ToMD5Pic(cfg.Base64, "")
		if err != nil {
			return cfg, err
		}
		cfg.Image = imagePath
	}
	if cfg.FontSize <= 0 {
		cfg.FontSize = defaultFontSize
	}
	if cfg.Color == "" {
		cfg.Color = "#ffffff"
	}
	if cfg.Opacity <= 0 {
		cfg.Opacity = defaultWatermarkOpacity
	}
	if cfg.Opacity > 1 {
		cfg.Opacity = 1
	}
	return cfg, nil
}

// watermarkAlignment accepts both hotgo numeric and Lina named alignment values.
type watermarkAlignment string

// UnmarshalYAML decodes alignment from YAML strings or integers.
func (a *watermarkAlignment) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		*a = watermarkAlignment(strings.TrimSpace(value.Value))
		return nil
	default:
		*a = ""
		return nil
	}
}

// normalizedAlignment converts named or HotGo numeric alignment values.
func normalizedAlignment(align watermarkAlignment) string {
	value := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(string(align)), "_", ""))
	value = strings.ReplaceAll(value, "-", "")
	switch value {
	case "1", "left":
		return "left"
	case "2", "center", "centre":
		return "center"
	case "3", "right":
		return "right"
	case "4", "top":
		return "top"
	case "5", "bottom":
		return "bottom"
	case "6", "topleft":
		return "topleft"
	case "7", "topright":
		return "topright"
	case "8", "bottomleft":
		return "bottomleft"
	case "9", "bottomright":
		return "bottomright"
	default:
		return "topleft"
	}
}

// ToHotGoAlignment converts named or numeric Lina strategy values to the migrated HotGo enum.
func (a watermarkAlignment) ToHotGoAlignment() watermark.Alignment {
	switch normalizedAlignment(a) {
	case "left":
		return watermark.AlignmentLeft
	case "center":
		return watermark.AlignmentCenter
	case "right":
		return watermark.AlignmentRight
	case "top":
		return watermark.AlignmentTop
	case "bottom":
		return watermark.AlignmentBottom
	case "topleft":
		return watermark.AlignmentTopLeft
	case "topright":
		return watermark.AlignmentTopRight
	case "bottomleft":
		return watermark.AlignmentBottomLeft
	case "bottomright":
		return watermark.AlignmentBottomRight
	default:
		return watermark.AlignmentNothing
	}
}
