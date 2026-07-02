// This file resolves media strategies and parses snapshot watermark strategy YAML.

package water

import (
	"context"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/cachecap"
)

// strategyYAML is a projection of media strategy YAML for snapshot watermark rules.
type strategyYAML struct {
	SnapshotWatermark *watermarkConfig `json:"snapshot_watermark" yaml:"snapshot_watermark"` // SnapshotWatermark is the screenshot watermark strategy node.
}

// parseWatermarkStrategy parses snapshot watermark configuration from strategy YAML.
func parseWatermarkStrategy(strategyBody string) (*watermarkConfig, error) {
	body := strings.TrimSpace(strategyBody)
	if body == "" {
		return nil, nil
	}
	var parsed strategyYAML
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyParseFailed)
	}

	cfg := parsed.SnapshotWatermark
	if cfg == nil {
		return nil, nil
	}
	normalized, err := normalizeWatermarkConfig(*cfg)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

// resolveStrategy delegates effective media strategy lookup to the configured resolver.
func (s *serviceImpl) resolveStrategy(ctx context.Context, tenantID string, deviceID string) (*resolvedStrategy, error) {
	if s == nil || s.strategyResolver == nil {
		return nil, bizerr.NewCode(CodeWaterMediaResolverUnavailable)
	}
	in := ResolveStrategyInput{
		TenantId: strings.TrimSpace(tenantID),
		DeviceId: strings.TrimSpace(deviceID),
	}
	if cached, ok := s.cachedResolveStrategy(ctx, in); ok {
		return buildResolvedStrategy(cached), nil
	}
	strategy, err := s.strategyResolver.ResolveStrategy(ctx, in)
	if err != nil {
		return nil, err
	}
	if strategy == nil {
		strategy = &ResolveStrategyOutput{
			Matched:     false,
			Source:      string(StrategySourceNone),
			SourceLabel: strategySourceLabel(StrategySourceNone),
		}
	}
	s.cacheResolveStrategy(ctx, in, strategy)
	return buildResolvedStrategy(strategy), nil
}

// cachedResolveStrategy returns one cached remote strategy projection when available.
func (s *serviceImpl) cachedResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, bool) {
	if s == nil || s.strategyCache == nil {
		return nil, false
	}
	item, found, err := s.strategyCache.Get(ctx, strategyResolveCacheNamespace, strategyResolveCacheKey(in))
	if err != nil {
		logger.Warningf(ctx, "读取水印策略解析缓存失败: %v", err)
		return nil, false
	}
	if !found || item == nil || item.ValueKind != cachecap.CacheValueKindString || strings.TrimSpace(item.Value) == "" {
		return nil, false
	}
	var out ResolveStrategyOutput
	if err = json.Unmarshal([]byte(item.Value), &out); err != nil {
		logger.Warningf(ctx, "解析水印策略解析缓存失败: %v", err)
		return nil, false
	}
	return &out, true
}

// cacheResolveStrategy stores one remote strategy projection for short-lived reuse.
func (s *serviceImpl) cacheResolveStrategy(ctx context.Context, in ResolveStrategyInput, out *ResolveStrategyOutput) {
	if s == nil || s.strategyCache == nil || out == nil {
		return
	}
	payload, err := json.Marshal(out)
	if err != nil {
		logger.Warningf(ctx, "序列化水印策略解析缓存失败: %v", err)
		return
	}
	if _, err = s.strategyCache.Set(
		ctx,
		strategyResolveCacheNamespace,
		strategyResolveCacheKey(in),
		string(payload),
		strategyResolveCacheTTL,
	); err != nil {
		logger.Warningf(ctx, "写入水印策略解析缓存失败: %v", err)
	}
}

// strategyResolveCacheKey builds a readable tenant/device key for remote resolution results.
func strategyResolveCacheKey(in ResolveStrategyInput) string {
	return strategyResolveCacheKeyPrefix +
		"tenant:" + strategyResolveCacheSegment(in.TenantId) +
		":device:" + strategyResolveCacheSegment(in.DeviceId)
}

// strategyResolveCacheSegment keeps cache keys readable while preserving empty values.
func strategyResolveCacheSegment(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "_"
	}
	return strings.NewReplacer(" ", "_", "\t", "_", "\n", "_", "\r", "_").Replace(normalized)
}

// buildResolvedStrategy converts the media strategy projection into service output.
func buildResolvedStrategy(strategy *ResolveStrategyOutput) *resolvedStrategy {
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
