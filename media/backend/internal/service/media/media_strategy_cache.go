// This file implements short-lived effective strategy caching for media lookups.

package media

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/logger"
	"lina-core/pkg/plugin/capability/cachecap"
)

// Effective strategy cache constants.
const (
	strategyResolveCacheNamespace  = "strategy-resolve"
	strategyResolveCacheVersionKey = "version"
	strategyResolveCacheKeyPrefix  = "resolve:"
	strategyResolveCacheTTL        = 30 * time.Second
	strategyResolveCacheVersionTTL = 24 * time.Hour
)

// ResolveStrategy resolves the effective strategy for one tenant/device pair.
func (s *serviceImpl) ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	if err := validateMediaTablesReady(ctx); err != nil {
		return nil, err
	}
	normalized := normalizeResolveStrategyInput(in)
	cacheSvc, version, cacheOK := s.strategyResolveCacheVersion(ctx)
	if cacheOK {
		if cached, ok := cachedResolvedStrategy(ctx, cacheSvc, version, normalized); ok {
			return cached, nil
		}
	}
	resolved, err := s.resolveStrategyFromStore(ctx, normalized)
	if err != nil {
		return nil, err
	}
	if cacheOK {
		cacheResolvedStrategy(ctx, cacheSvc, version, normalized, resolved)
	}
	return resolved, nil
}

// normalizeResolveStrategyInput trims one strategy lookup input.
func normalizeResolveStrategyInput(in ResolveStrategyInput) ResolveStrategyInput {
	return ResolveStrategyInput{
		TenantId: strings.TrimSpace(in.TenantId),
		DeviceId: strings.TrimSpace(in.DeviceId),
	}
}

// resolveStrategyFromStore performs the authoritative DB-backed strategy lookup.
func (s *serviceImpl) resolveStrategyFromStore(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	if in.TenantId != "" && in.DeviceId != "" {
		strategy, err := s.strategyFromTenantDeviceBinding(ctx, in.TenantId, in.DeviceId)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceTenantDevice, strategy), nil
		}
	}
	if in.DeviceId != "" {
		strategy, err := s.strategyFromDeviceBinding(ctx, in.DeviceId)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceDevice, strategy), nil
		}
	}
	if in.TenantId != "" {
		strategy, err := s.strategyFromTenantBinding(ctx, in.TenantId)
		if err != nil {
			return nil, err
		}
		if strategy != nil {
			return buildResolveOutput(StrategySourceTenant, strategy), nil
		}
	}
	strategy, err := s.globalStrategy(ctx)
	if err != nil {
		return nil, err
	}
	if strategy != nil {
		return buildResolveOutput(StrategySourceGlobal, strategy), nil
	}
	return buildResolveOutput(StrategySourceNone, nil), nil
}

// cachedResolvedStrategy returns one cached effective strategy result when present.
func cachedResolvedStrategy(
	ctx context.Context,
	cacheSvc mediaCache,
	version int64,
	in ResolveStrategyInput,
) (*ResolveStrategyOutput, bool) {
	key := strategyResolveCacheKey(version, in)
	item, found, err := cacheSvc.Get(ctx, strategyResolveCacheNamespace, key)
	if err != nil {
		logger.Warningf(ctx, "读取媒体策略解析缓存失败: %v", err)
		return nil, false
	}
	if !found || item == nil || item.ValueKind != cachecap.CacheValueKindString || strings.TrimSpace(item.Value) == "" {
		return nil, false
	}
	var out ResolveStrategyOutput
	if err = json.Unmarshal([]byte(item.Value), &out); err != nil {
		logger.Warningf(ctx, "解析媒体策略解析缓存失败: %v", err)
		return nil, false
	}
	return &out, true
}

// cacheResolvedStrategy stores one effective strategy result for short-lived reuse.
func cacheResolvedStrategy(
	ctx context.Context,
	cacheSvc mediaCache,
	version int64,
	in ResolveStrategyInput,
	out *ResolveStrategyOutput,
) {
	if out == nil {
		return
	}
	payload, err := json.Marshal(out)
	if err != nil {
		logger.Warningf(ctx, "序列化媒体策略解析缓存失败: %v", err)
		return
	}
	if _, err = cacheSvc.Set(
		ctx,
		strategyResolveCacheNamespace,
		strategyResolveCacheKey(version, in),
		string(payload),
		strategyResolveCacheTTL,
	); err != nil {
		logger.Warningf(ctx, "写入媒体策略解析缓存失败: %v", err)
	}
}

// strategyResolveCacheVersion returns the current shared strategy cache version.
func (s *serviceImpl) strategyResolveCacheVersion(ctx context.Context) (mediaCache, int64, bool) {
	if s == nil || s.cacheSvc == nil {
		return nil, 0, false
	}
	item, found, err := s.cacheSvc.Get(ctx, strategyResolveCacheNamespace, strategyResolveCacheVersionKey)
	if err != nil {
		logger.Warningf(ctx, "读取媒体策略解析缓存版本失败: %v", err)
		return nil, 0, false
	}
	if found && item != nil && item.ValueKind == cachecap.CacheValueKindInt {
		return s.cacheSvc, item.IntValue, true
	}
	item, err = s.cacheSvc.Incr(
		ctx,
		strategyResolveCacheNamespace,
		strategyResolveCacheVersionKey,
		0,
		strategyResolveCacheVersionTTL,
	)
	if err != nil {
		logger.Warningf(ctx, "读取媒体策略解析缓存版本失败: %v", err)
		return nil, 0, false
	}
	if item == nil || item.ValueKind != cachecap.CacheValueKindInt {
		return nil, 0, false
	}
	return s.cacheSvc, item.IntValue, true
}

// invalidateStrategyResolveCache bumps the shared version so older entries are bypassed.
func (s *serviceImpl) invalidateStrategyResolveCache(ctx context.Context) {
	if s == nil || s.cacheSvc == nil {
		return
	}
	if _, err := s.cacheSvc.Incr(
		ctx,
		strategyResolveCacheNamespace,
		strategyResolveCacheVersionKey,
		1,
		strategyResolveCacheVersionTTL,
	); err != nil {
		logger.Warningf(ctx, "失效媒体策略解析缓存失败: %v", err)
	}
}

// strategyResolveCacheKey builds a readable key for one versioned tenant/device lookup.
func strategyResolveCacheKey(version int64, in ResolveStrategyInput) string {
	return strategyResolveCacheKeyPrefix +
		"tenant:" + strategyResolveCacheSegment(in.TenantId) +
		":device:" + strategyResolveCacheSegment(in.DeviceId) +
		":version:" + strconv.FormatInt(version, 10)
}

// strategyResolveCacheSegment keeps cache keys readable while avoiding blank segments.
func strategyResolveCacheSegment(value string) string {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "_"
	}
	return strings.NewReplacer(" ", "_", "\t", "_", "\n", "_", "\r", "_").Replace(normalized)
}
