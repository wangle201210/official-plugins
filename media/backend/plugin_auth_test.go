// This file verifies the media plugin authentication helpers that sit between
// the host middleware chain and media-owned Tieta fallback context.

package backend

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/bizctxcap"
)

// mediaStaticBizCtx returns one fixed plugin-visible context for auth helper tests.
type mediaStaticBizCtx struct {
	current bizctxcap.CurrentContext
}

// Current returns the configured business context snapshot.
func (s mediaStaticBizCtx) Current(context.Context) bizctxcap.CurrentContext {
	return s.current
}

// TestMediaBizCtxOverlayPrefersTietaContext verifies fallback identity is visible to media services.
func TestMediaBizCtxOverlayPrefersTietaContext(t *testing.T) {
	base := mediaStaticBizCtx{current: bizctxcap.CurrentContext{
		UserID:   1,
		Username: "host-user",
		TenantID: 10,
	}}
	overlay := mediaBizCtxWithTietaOverlay(base)
	tietaCtx := bizctxcap.WithCurrentContext(context.Background(), bizctxcap.CurrentContext{
		UserID:   13,
		Username: "wj530",
		TenantID: 0,
	})
	tietaCtx = context.WithValue(tietaCtx, mediaTietaCurrentContextKey{}, bizctxcap.CurrentFromContext(tietaCtx))

	current := overlay.Current(tietaCtx)
	if current.UserID != 13 || current.Username != "wj530" || !current.PlatformBypass {
		t.Fatalf("expected Tieta context to override host context, got %#v", current)
	}

	current = overlay.Current(context.Background())
	if current.UserID != 1 || current.Username != "host-user" || current.TenantID != 10 {
		t.Fatalf("expected host context fallback, got %#v", current)
	}
}
