// This file loads remote media strategy resolver configuration for water.

package water

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/plugin/capability/plugincap"
)

// RemoteStrategyResolverConfig defines the HTTP settings for remote media strategy lookup.
type RemoteStrategyResolverConfig struct {
	BaseURL string        // BaseURL is the remote LinaPro media cluster HTTP base URL.
	APIKey  string        // APIKey is sent as X-Inner-Api-Key when non-empty.
	Timeout time.Duration // Timeout bounds one remote strategy lookup request.
}

// LoadRemoteStrategyResolverConfig reads water plugin-scoped runtime configuration.
func LoadRemoteStrategyResolverConfig(
	ctx context.Context,
	configSvc plugincap.ConfigService,
) (RemoteStrategyResolverConfig, error) {
	if configSvc == nil {
		return RemoteStrategyResolverConfig{Timeout: defaultStrategyResolverTimeout}, nil
	}
	baseURL, err := configSvc.String(ctx, "mediaStrategy.baseUrl", "")
	if err != nil {
		return RemoteStrategyResolverConfig{}, err
	}
	apiKey, err := configSvc.String(ctx, "mediaStrategy.apiKey", "media")
	if err != nil {
		return RemoteStrategyResolverConfig{}, err
	}
	timeout, err := configSvc.Duration(ctx, "mediaStrategy.timeout", defaultStrategyResolverTimeout)
	if err != nil {
		return RemoteStrategyResolverConfig{}, err
	}
	if timeout <= 0 {
		timeout = defaultStrategyResolverTimeout
	}
	return RemoteStrategyResolverConfig{
		BaseURL: strings.TrimSpace(baseURL),
		APIKey:  strings.TrimSpace(apiKey),
		Timeout: timeout,
	}, nil
}
