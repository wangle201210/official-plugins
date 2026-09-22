// This file verifies customer-code tenant identity across Tieta responses and cached users.

package media

import (
	"testing"

	"github.com/gogf/gf/v2/encoding/gjson"

	"lina-core/pkg/bizerr"
)

// TestBuildTietaUserUsesCustomerCode verifies raw customer IDs never become media tenant keys.
func TestBuildTietaUserUsesCustomerCode(t *testing.T) {
	for _, tc := range []struct {
		name       string
		payload    string
		wantTenant string
	}{
		{name: "different ID and code", payload: `{"id":13,"customerId":"12345","customerCode":" tenant-code "}`, wantTenant: "tenant-code"},
		{name: "code without ID", payload: `{"id":13,"customerCode":"tenant-code"}`, wantTenant: "tenant-code"},
		{name: "missing code", payload: `{"id":13,"customerId":"12345"}`},
		{name: "blank code", payload: `{"id":13,"customerId":"12345","customerCode":"  "}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var info tietaUserInfo
			if err := gjson.Unmarshal([]byte(tc.payload), &info); err != nil {
				t.Fatalf("decode upstream identity: %v", err)
			}
			user := buildTietaUser(&info)
			if user.TenantId != tc.wantTenant || user.CustomerCode != tc.wantTenant {
				t.Fatalf("expected customer-code identity %q, got %+v", tc.wantTenant, user)
			}
			tenantID, err := resolveTietaTenantID("", user)
			if tc.wantTenant == "" {
				structured, ok := bizerr.As(err)
				if !ok || structured.RuntimeCode() != "MEDIA_TIETA_TENANT_MISSING" {
					t.Fatalf("expected missing tenant rejection without customerId fallback, got %v", err)
				}
				return
			}
			if err != nil || tenantID != tc.wantTenant {
				t.Fatalf("resolve customer-code tenant: tenant=%q err=%v", tenantID, err)
			}
			if _, err := resolveTietaTenantID("12345", user); err == nil {
				t.Fatal("expected customerId override to be rejected")
			}
		})
	}
}

// TestAuthenticateCachedTietaTokenDerivesTenantFromCustomerCode verifies shared cached snapshots use the same identity mapping.
func TestAuthenticateCachedTietaTokenDerivesTenantFromCustomerCode(t *testing.T) {
	for _, code := range []string{" tenant-code ", ""} {
		t.Run(code, func(t *testing.T) {
			var (
				ctx      = t.Context()
				cacheSvc = newMemoryRouteMemoryCache()
				client   = &fakeTietaClient{}
			)
			restore := replaceMediaTietaClient(t, client)
			t.Cleanup(restore)
			setCachedTietaUser(ctx, cacheSvc, "cached-token", &TietaUser{
				Id: 13, CustomerCode: code, TenantId: "12345",
			})
			for range 2 {
				user, err := authenticateCachedTietaToken(ctx, cacheSvc, newTestMediaConfig(), "cached-token")
				if err != nil {
					t.Fatalf("read cached user: %v", err)
				}
				if code != "" && user.TenantId != "tenant-code" {
					t.Fatalf("expected customer-code tenant on cache hit, got %+v", user)
				}
				if code == "" {
					if _, err := resolveTietaTenantID("", user); err == nil {
						t.Fatal("expected missing cached customer code to reject tenant access")
					}
				}
			}
			if len(client.tokens) != 0 {
				t.Fatalf("cache hits must not add upstream requests, got %v", client.tokens)
			}
		})
	}
}

// TestMockTietaUserUsesCustomerCode verifies local mock identity follows the production tenant contract.
func TestMockTietaUserUsesCustomerCode(t *testing.T) {
	for _, token := range []string{"mock", "tenant-code"} {
		user := mockTietaUser(token)
		if user.CustomerCode == "" || user.CustomerCode != user.TenantId {
			t.Fatalf("expected mock customer code to equal tenant key, got %+v", user)
		}
	}
}
