// This file implements HotGo-compatible route memory backed by plugin Redis.

package media

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/gogf/gf/contrib/nosql/redis/v2"
	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// Route memory cache constants.
const (
	routeMemoryKeyFormat = "route_data:%s:%s"
	routeMemoryTTL       = 12 * time.Hour

	routeMemoryRedisAddressKey               = "media.routeMemory.redis.address"        // routeMemoryRedisAddressKey overrides the route-memory Redis endpoint.
	routeMemoryRedisDBKey                    = "media.routeMemory.redis.db"             // routeMemoryRedisDBKey selects the plugin-specific route-memory Redis DB.
	routeMemoryRedisPasswordKey              = "media.routeMemory.redis.password"       // routeMemoryRedisPasswordKey authenticates plugin-specific route-memory Redis.
	routeMemoryRedisConnectTimeoutKey        = "media.routeMemory.redis.connectTimeout" // routeMemoryRedisConnectTimeoutKey bounds route-memory Redis connection setup.
	routeMemoryRedisReadTimeoutKey           = "media.routeMemory.redis.readTimeout"    // routeMemoryRedisReadTimeoutKey bounds route-memory Redis reads.
	routeMemoryRedisWriteTimeoutKey          = "media.routeMemory.redis.writeTimeout"   // routeMemoryRedisWriteTimeoutKey bounds route-memory Redis writes.
	routeMemoryClusterRedisAddressKey        = "cluster.redis.address"                  // routeMemoryClusterRedisAddressKey reuses host cluster Redis when no plugin endpoint is set.
	routeMemoryClusterRedisDBKey             = "cluster.redis.db"                       // routeMemoryClusterRedisDBKey reuses host cluster Redis DB.
	routeMemoryClusterRedisPasswordKey       = "cluster.redis.password"                 // routeMemoryClusterRedisPasswordKey reuses host cluster Redis password.
	routeMemoryClusterRedisConnectTimeoutKey = "cluster.redis.connectTimeout"           // routeMemoryClusterRedisConnectTimeoutKey reuses host cluster Redis connect timeout.
	routeMemoryClusterRedisReadTimeoutKey    = "cluster.redis.readTimeout"              // routeMemoryClusterRedisReadTimeoutKey reuses host cluster Redis read timeout.
	routeMemoryClusterRedisWriteTimeoutKey   = "cluster.redis.writeTimeout"             // routeMemoryClusterRedisWriteTimeoutKey reuses host cluster Redis write timeout.
	routeMemoryGoFrameRedisAddressKey        = "redis.default.address"                  // routeMemoryGoFrameRedisAddressKey supports standard GoFrame Redis config fallback.
	routeMemoryGoFrameRedisDBKey             = "redis.default.db"                       // routeMemoryGoFrameRedisDBKey supports standard GoFrame Redis DB fallback.
	routeMemoryGoFrameRedisPasswordKey       = "redis.default.pass"                     // routeMemoryGoFrameRedisPasswordKey supports standard GoFrame Redis password fallback.

	routeMemoryDefaultConnectTimeout = 3 * time.Second
	routeMemoryDefaultReadTimeout    = 2 * time.Second
	routeMemoryDefaultWriteTimeout   = 2 * time.Second
)

var defaultRouteMemoryRedis = struct {
	sync.Mutex
	signature string
	client    *gredis.Redis
}{}

// routeMemoryStore defines the plugin-owned route memory persistence boundary.
type routeMemoryStore interface {
	// Set writes one route memory value with the given TTL.
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	// Get reads one route memory value; a missing key returns an empty string.
	Get(ctx context.Context, key string) (string, error)
	// Delete removes one route memory value.
	Delete(ctx context.Context, key string) error
}

// defaultRouteMemoryStore stores route memory in the configured Redis backend.
type defaultRouteMemoryStore struct{}

// RouteMemoryKeyInput defines one route-memory device/channel key.
type RouteMemoryKeyInput struct {
	DeviceCode  string // DeviceCode is the HotGo-compatible device code.
	ChannelCode string // ChannelCode is the HotGo-compatible channel code.
}

// RouteMemoryInput defines one route-memory write request.
type RouteMemoryInput struct {
	RouteMemoryKeyInput
	Data string // Data is the route payload stored in Redis.
}

// RouteMemoryOutput defines one route-memory read result.
type RouteMemoryOutput struct {
	Data string // Data is the stored route payload or empty when missing.
}

// SetRouteMemory stores HotGo-compatible route memory for one device channel.
func (s *serviceImpl) SetRouteMemory(ctx context.Context, in RouteMemoryInput) error {
	key, err := s.routeMemoryCacheKey(in.RouteMemoryKeyInput)
	if err != nil {
		return err
	}
	if strings.TrimSpace(in.Data) == "" {
		return bizerr.NewCode(CodeMediaRouteDataRequired)
	}
	if err = s.routeMemoryStore.Set(ctx, key, in.Data, routeMemoryTTL); err != nil {
		return bizerr.WrapCode(err, CodeMediaRouteSetFailed)
	}
	return nil
}

// GetRouteMemory reads HotGo-compatible route memory for one device channel.
func (s *serviceImpl) GetRouteMemory(ctx context.Context, in RouteMemoryKeyInput) (*RouteMemoryOutput, error) {
	key, err := s.routeMemoryCacheKey(in)
	if err != nil {
		return nil, err
	}
	data, err := s.routeMemoryStore.Get(ctx, key)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaRouteGetFailed)
	}
	return &RouteMemoryOutput{Data: data}, nil
}

// DeleteRouteMemory removes HotGo-compatible route memory for one device channel.
func (s *serviceImpl) DeleteRouteMemory(ctx context.Context, in RouteMemoryKeyInput) error {
	key, err := s.routeMemoryCacheKey(in)
	if err != nil {
		return err
	}
	if err = s.routeMemoryStore.Delete(ctx, key); err != nil {
		return bizerr.WrapCode(err, CodeMediaRouteDeleteFailed)
	}
	return nil
}

// routeMemoryCacheKey builds the HotGo-compatible Redis key for route memory.
func (s *serviceImpl) routeMemoryCacheKey(in RouteMemoryKeyInput) (string, error) {
	deviceCode, err := normalizeDeviceCode(in.DeviceCode)
	if err != nil {
		return "", err
	}
	channelCode, err := normalizeChannelCode(in.ChannelCode)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(routeMemoryKeyFormat, deviceCode, channelCode), nil
}

// Set writes one route memory value into Redis with an explicit TTL.
func (s defaultRouteMemoryStore) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	client, err := routeMemoryRedisClient(ctx)
	if err != nil {
		return err
	}
	ttlMillis := ttl.Milliseconds()
	_, err = client.Set(ctx, key, value, gredis.SetOption{
		TTLOption: gredis.TTLOption{PX: gconv.PtrInt64(ttlMillis)},
	})
	return err
}

// Get reads one route memory value from Redis; missing keys return an empty string.
func (s defaultRouteMemoryStore) Get(ctx context.Context, key string) (string, error) {
	client, err := routeMemoryRedisClient(ctx)
	if err != nil {
		return "", err
	}
	value, err := client.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if routeMemoryValueMissing(value) {
		return "", nil
	}
	return value.String(), nil
}

// Delete removes one route memory value from Redis.
func (s defaultRouteMemoryStore) Delete(ctx context.Context, key string) error {
	client, err := routeMemoryRedisClient(ctx)
	if err != nil {
		return err
	}
	_, err = client.Del(ctx, key)
	return err
}

// routeMemoryRedisClient returns the lazily initialized Redis client used by route memory.
func routeMemoryRedisClient(ctx context.Context) (*gredis.Redis, error) {
	config, signature, err := routeMemoryRedisConfig(ctx)
	if err != nil {
		return nil, err
	}
	defaultRouteMemoryRedis.Lock()
	defer defaultRouteMemoryRedis.Unlock()
	if defaultRouteMemoryRedis.client != nil && defaultRouteMemoryRedis.signature == signature {
		return defaultRouteMemoryRedis.client, nil
	}
	if defaultRouteMemoryRedis.client != nil {
		if closeErr := defaultRouteMemoryRedis.client.Close(ctx); closeErr != nil {
			logger.Warningf(ctx, "关闭 media 路由记忆 Redis 客户端失败: %v", closeErr)
		}
	}
	client, err := gredis.New(config)
	if err != nil {
		defaultRouteMemoryRedis.client = nil
		defaultRouteMemoryRedis.signature = ""
		return nil, err
	}
	defaultRouteMemoryRedis.client = client
	defaultRouteMemoryRedis.signature = signature
	return client, nil
}

// routeMemoryRedisConfig resolves plugin-specific Redis config with cluster Redis fallback.
func routeMemoryRedisConfig(ctx context.Context) (*gredis.Config, string, error) {
	address, source := routeMemoryConfigString(ctx, routeMemoryRedisAddressKey, "")
	db := 0
	password := ""
	connectTimeout := routeMemoryDefaultConnectTimeout
	readTimeout := routeMemoryDefaultReadTimeout
	writeTimeout := routeMemoryDefaultWriteTimeout
	if address == "" {
		address, source = routeMemoryConfigString(ctx, routeMemoryClusterRedisAddressKey, "")
	}
	if address == "" {
		address, source = routeMemoryConfigString(ctx, routeMemoryGoFrameRedisAddressKey, "")
	}
	if address == "" {
		return nil, "", gerror.New("media route memory redis address is not configured")
	}
	switch source {
	case routeMemoryRedisAddressKey:
		db = routeMemoryConfigInt(ctx, routeMemoryRedisDBKey, db)
		password, _ = routeMemoryConfigString(ctx, routeMemoryRedisPasswordKey, "")
		connectTimeout = routeMemoryConfigDuration(ctx, routeMemoryRedisConnectTimeoutKey, connectTimeout)
		readTimeout = routeMemoryConfigDuration(ctx, routeMemoryRedisReadTimeoutKey, readTimeout)
		writeTimeout = routeMemoryConfigDuration(ctx, routeMemoryRedisWriteTimeoutKey, writeTimeout)
	case routeMemoryClusterRedisAddressKey:
		db = routeMemoryConfigInt(ctx, routeMemoryClusterRedisDBKey, db)
		password, _ = routeMemoryConfigString(ctx, routeMemoryClusterRedisPasswordKey, "")
		connectTimeout = routeMemoryConfigDuration(ctx, routeMemoryClusterRedisConnectTimeoutKey, connectTimeout)
		readTimeout = routeMemoryConfigDuration(ctx, routeMemoryClusterRedisReadTimeoutKey, readTimeout)
		writeTimeout = routeMemoryConfigDuration(ctx, routeMemoryClusterRedisWriteTimeoutKey, writeTimeout)
	case routeMemoryGoFrameRedisAddressKey:
		db = routeMemoryConfigInt(ctx, routeMemoryGoFrameRedisDBKey, db)
		password, _ = routeMemoryConfigString(ctx, routeMemoryGoFrameRedisPasswordKey, "")
	}
	signature := fmt.Sprintf(
		"%s|%s|%d|%s|%s|%s|%s",
		source,
		address,
		db,
		password,
		connectTimeout,
		readTimeout,
		writeTimeout,
	)
	return &gredis.Config{
		Address:      address,
		Db:           db,
		Pass:         password,
		DialTimeout:  connectTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}, signature, nil
}

// routeMemoryConfigString reads one string config value and returns its source key.
func routeMemoryConfigString(ctx context.Context, key string, defaultValue string) (string, string) {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		logger.Warningf(ctx, "读取 media 路由记忆配置失败 key=%s err=%v", key, err)
		return defaultValue, ""
	}
	if configValueMissing(value) {
		return defaultValue, ""
	}
	trimmed := strings.TrimSpace(value.String())
	if trimmed == "" {
		return defaultValue, ""
	}
	return trimmed, key
}

// routeMemoryConfigInt reads one integer config value.
func routeMemoryConfigInt(ctx context.Context, key string, defaultValue int) int {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		logger.Warningf(ctx, "读取 media 路由记忆配置失败 key=%s err=%v", key, err)
		return defaultValue
	}
	if configValueMissing(value) {
		return defaultValue
	}
	return value.Int()
}

// routeMemoryConfigDuration reads one duration config value.
func routeMemoryConfigDuration(ctx context.Context, key string, defaultValue time.Duration) time.Duration {
	value, err := g.Cfg().Get(ctx, key)
	if err != nil {
		logger.Warningf(ctx, "读取 media 路由记忆配置失败 key=%s err=%v", key, err)
		return defaultValue
	}
	if configValueMissing(value) {
		return defaultValue
	}
	raw := strings.TrimSpace(value.String())
	if raw == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(raw)
	if err != nil || duration <= 0 {
		logger.Warningf(ctx, "media 路由记忆 duration 配置无效 key=%s value=%s err=%v", key, raw, err)
		return defaultValue
	}
	return duration
}

// routeMemoryValueMissing reports whether one Redis value should be treated as absent.
func routeMemoryValueMissing(value *gvar.Var) bool {
	return value == nil || value.IsNil()
}

// normalizeDeviceCode validates one route-memory device code.
func normalizeDeviceCode(deviceCode string) (string, error) {
	trimmed := strings.TrimSpace(deviceCode)
	if trimmed == "" {
		return "", bizerr.NewCode(CodeMediaDeviceIDRequired)
	}
	if len([]rune(trimmed)) > 64 {
		return "", bizerr.NewCode(CodeMediaDeviceIDTooLong)
	}
	return trimmed, nil
}

// normalizeChannelCode validates one route-memory channel code.
func normalizeChannelCode(channelCode string) (string, error) {
	trimmed := strings.TrimSpace(channelCode)
	if trimmed == "" {
		return "", bizerr.NewCode(CodeMediaChannelIDRequired)
	}
	if len([]rune(trimmed)) > 64 {
		return "", bizerr.NewCode(CodeMediaChannelIDTooLong)
	}
	return trimmed, nil
}
