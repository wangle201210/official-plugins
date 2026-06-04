// Package strategy publishes media-owned strategy resolution contracts for source plugins under backend/.
package strategy

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/contract"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// Resolver defines the stable media strategy lookup contract for dependent plugins.
type Resolver interface {
	// ResolveStrategy resolves the effective media strategy for one tenant/device pair.
	ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error)
}

// ResolveStrategyInput defines the public media strategy lookup input.
type ResolveStrategyInput struct {
	TenantId string // TenantId is the media tenant ID.
	DeviceId string // DeviceId is the GB device ID.
}

// ResolveStrategyOutput defines the public media strategy lookup output.
type ResolveStrategyOutput struct {
	Matched      bool   // Matched reports whether a strategy matched.
	Source       string // Source is the matching strategy source.
	SourceLabel  string // SourceLabel is the source label prepared by media.
	StrategyId   int64  // StrategyId is the matched strategy ID.
	StrategyName string // StrategyName is the matched strategy name.
	Strategy     string // Strategy is the matched YAML strategy body.
}

// resolver adapts the internal media service to the public strategy contract.
type resolver struct {
	mediaSvc mediasvc.Service // mediaSvc owns the authoritative strategy resolution implementation.
}

// NewResolver creates a media strategy resolver backed by the media service.
func NewResolver(bizCtxSvc contract.BizCtxService, cacheSvc contract.CacheService) (Resolver, error) {
	mediaSvc, err := mediasvc.New(bizCtxSvc, cacheSvc)
	if err != nil {
		return nil, err
	}
	return NewResolverFromService(mediaSvc)
}

// NewResolverFromService adapts an already constructed media service into the public resolver contract.
func NewResolverFromService(mediaSvc mediasvc.Service) (Resolver, error) {
	if mediaSvc == nil {
		return nil, gerror.New("media strategy resolver requires media service")
	}
	return &resolver{mediaSvc: mediaSvc}, nil
}

// ResolveStrategy resolves one effective strategy without exposing media internal DAO or entity types.
func (r *resolver) ResolveStrategy(ctx context.Context, in ResolveStrategyInput) (*ResolveStrategyOutput, error) {
	if r == nil || r.mediaSvc == nil {
		return nil, gerror.New("media strategy resolver is not initialized")
	}
	out, err := r.mediaSvc.ResolveStrategy(ctx, mediasvc.ResolveStrategyInput{
		TenantId: in.TenantId,
		DeviceId: in.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, nil
	}
	return &ResolveStrategyOutput{
		Matched:      out.Matched,
		Source:       out.Source,
		SourceLabel:  out.SourceLabel,
		StrategyId:   out.StrategyId,
		StrategyName: out.StrategyName,
		Strategy:     out.Strategy,
	}, nil
}
