// This file tests the HotGo-compatible Tieta user response mapping.

package mediaopen

import (
	"encoding/json"
	"testing"

	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TestBuildCompatTietaUserInfoReturnsOnlyCustomerNameAndPhone verifies the narrowed response projection.
func TestBuildCompatTietaUserInfoReturnsOnlyCustomerNameAndPhone(t *testing.T) {
	user := buildCompatTietaUserInfo(&mediasvc.TietaUser{
		CustomerName: "公安",
		Mobile:       "18213268117",
	})
	if user == nil || user.CustomerName != "公安" || user.Phone != "18213268117" {
		t.Fatalf("expected customerName and phone projection, got %+v", user)
	}

	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal Tieta user info: %v", err)
	}
	var fields map[string]any
	if err = json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal Tieta user info: %v", err)
	}
	if len(fields) != 2 || fields["customerName"] != "公安" || fields["phone"] != "18213268117" {
		t.Fatalf("expected JSON to contain only customerName and phone, got %s", payload)
	}
}
