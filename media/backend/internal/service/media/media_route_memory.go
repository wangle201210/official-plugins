// This file implements HotGo-compatible route memory backed by the host cache service.

package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/pluginservice/contract"
)

// Route memory cache constants.
const (
	routeMemoryNamespace = "route-memory"
	routeMemoryKeyFormat = "route_data:%s:%s"
	routeMemoryTTL       = 12 * time.Hour
)

// mediaCache defines the host cache operations used by transient media data.
type mediaCache interface {
	contract.CacheService
}

// RouteMemoryKeyInput defines one route-memory device/channel key.
type RouteMemoryKeyInput struct {
	DeviceCode  string // DeviceCode is the HotGo-compatible device code.
	ChannelCode string // ChannelCode is the HotGo-compatible channel code.
}

// RouteMemoryInput defines one route-memory write request.
type RouteMemoryInput struct {
	RouteMemoryKeyInput
	Data string // Data is the route payload stored in the host cache.
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
	cacheSvc, err := s.hostRouteMemoryCache()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaRouteSetFailed)
	}
	if _, err = cacheSvc.Set(ctx, routeMemoryNamespace, key, in.Data, routeMemoryTTL); err != nil {
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
	cacheSvc, err := s.hostRouteMemoryCache()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaRouteGetFailed)
	}
	item, found, err := cacheSvc.Get(ctx, routeMemoryNamespace, key)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMediaRouteGetFailed)
	}
	if !found || item == nil {
		return &RouteMemoryOutput{}, nil
	}
	return &RouteMemoryOutput{Data: item.Value}, nil
}

// DeleteRouteMemory removes HotGo-compatible route memory for one device channel.
func (s *serviceImpl) DeleteRouteMemory(ctx context.Context, in RouteMemoryKeyInput) error {
	key, err := s.routeMemoryCacheKey(in)
	if err != nil {
		return err
	}
	cacheSvc, err := s.hostRouteMemoryCache()
	if err != nil {
		return bizerr.WrapCode(err, CodeMediaRouteDeleteFailed)
	}
	if err = cacheSvc.Delete(ctx, routeMemoryNamespace, key); err != nil {
		return bizerr.WrapCode(err, CodeMediaRouteDeleteFailed)
	}
	return nil
}

// routeMemoryCacheKey builds the HotGo-compatible logical cache key for route memory.
func (s *serviceImpl) routeMemoryCacheKey(in RouteMemoryKeyInput) (string, error) {
	deviceCode, err := normalizeDeviceNodeKey(in.DeviceCode)
	if err != nil {
		return "", err
	}
	channelCode, err := normalizeDeviceNodeChannelID(in.ChannelCode)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(routeMemoryKeyFormat, deviceCode, channelCode), nil
}

// hostRouteMemoryCache returns the host-published cache adapter.
func (s *serviceImpl) hostRouteMemoryCache() (mediaCache, error) {
	if s == nil || s.cacheSvc == nil {
		return nil, gerror.New("media route memory cache service is unavailable")
	}
	return s.cacheSvc, nil
}
