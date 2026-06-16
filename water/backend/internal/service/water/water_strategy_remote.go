// This file implements the remote mediaopen HTTP strategy resolver.

package water

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/logger"
)

// remoteStrategyResolver resolves media strategies through the mediaopen API.
type remoteStrategyResolver struct {
	baseURL *url.URL     // baseURL is the remote LinaPro media cluster base URL.
	apiKey  string       // apiKey is sent to the mediaopen inner API key gate.
	client  *http.Client // client owns request timeout and transport settings.
}

// disabledStrategyResolver reports a missing remote strategy configuration at call time.
type disabledStrategyResolver struct{}

// NewRemoteStrategyResolver creates a remote mediaopen strategy resolver.
func NewRemoteStrategyResolver(config RemoteStrategyResolverConfig) (StrategyResolver, error) {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		return disabledStrategyResolver{}, nil
	}
	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		if err == nil {
			err = gerror.New("remote media strategy base URL must include scheme and host")
		}
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverConfigInvalid)
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultStrategyResolverTimeout
	}
	return &remoteStrategyResolver{
		baseURL: parsedURL,
		apiKey:  strings.TrimSpace(config.APIKey),
		client:  &http.Client{Timeout: timeout},
	}, nil
}

// ResolveStrategy always fails because the remote resolver is not configured.
func (disabledStrategyResolver) ResolveStrategy(context.Context, ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	return nil, bizerr.NewCode(CodeWaterMediaResolverUnavailable)
}

// ResolveStrategy resolves one media strategy through HTTP without importing the media plugin.
func (r *remoteStrategyResolver) ResolveStrategy(
	ctx context.Context,
	in ResolveStrategyInput,
) (*ResolveStrategyOutput, error) {
	if r == nil || r.baseURL == nil || r.client == nil {
		return nil, bizerr.NewCode(CodeWaterMediaResolverUnavailable)
	}
	targetURL, err := r.strategyURL(in)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverConfigInvalid)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverConfigInvalid)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "linapro-water/1.0")
	if r.apiKey != "" {
		req.Header.Set(strategyResolverAPIKeyHeader, r.apiKey)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverRequestFailed)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Warningf(ctx, "关闭媒体策略解析响应体失败: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxStrategyResolverResponseBytes))
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverResponseInvalid)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, bizerr.WrapCode(
			gerror.Newf("remote media strategy API returned status %d", resp.StatusCode),
			CodeWaterStrategyResolverRequestFailed,
		)
	}
	out, err := decodeRemoteStrategyResponse(body)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// strategyURL builds the remote mediaopen strategy resolution URL.
func (r *remoteStrategyResolver) strategyURL(in ResolveStrategyInput) (string, error) {
	target, err := url.JoinPath(r.baseURL.String(), strategyResolverPath)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	query.Set("tenantId", strings.TrimSpace(in.TenantId))
	query.Set("deviceId", strings.TrimSpace(in.DeviceId))
	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

// remoteStrategyEnvelope mirrors the host unified response wrapper.
type remoteStrategyEnvelope struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	ErrorCode string          `json:"errorCode"`
}

// decodeRemoteStrategyResponse decodes both unified and direct response payloads.
func decodeRemoteStrategyResponse(body []byte) (*ResolveStrategyOutput, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, bizerr.NewCode(CodeWaterStrategyResolverResponseInvalid)
	}
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverResponseInvalid)
	}
	if _, ok := probe["data"]; ok {
		var envelope remoteStrategyEnvelope
		if err := json.Unmarshal(trimmed, &envelope); err != nil {
			return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverResponseInvalid)
		}
		if envelope.Code != 0 {
			return nil, bizerr.WrapCode(
				gerror.Newf("remote media strategy API returned code %d", envelope.Code),
				CodeWaterStrategyResolverRequestFailed,
			)
		}
		trimmed = bytes.TrimSpace(envelope.Data)
	}
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nil, nil
	}
	var out ResolveStrategyOutput
	if err := json.Unmarshal(trimmed, &out); err != nil {
		return nil, bizerr.WrapCode(err, CodeWaterStrategyResolverResponseInvalid)
	}
	return &out, nil
}
