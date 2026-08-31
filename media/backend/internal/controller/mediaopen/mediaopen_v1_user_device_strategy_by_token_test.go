// This file tests the HotGo-compatible Tieta user response mapping.

package mediaopen

import (
	"encoding/json"
	"testing"

	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// TestBuildCompatTietaUserInfoReturnsOnlyIdCustomerNameAndPhone verifies the narrowed response projection.
func TestBuildCompatTietaUserInfoReturnsOnlyIdCustomerNameAndPhone(t *testing.T) {
	user := buildCompatTietaUserInfo(&mediasvc.TietaUser{
		Id:           13,
		CustomerName: "公安",
		Mobile:       "18213268117",
	})
	if user == nil || user.Id != 13 || user.CustomerName != "公安" || user.Phone != "18213268117" {
		t.Fatalf("expected id, customerName and phone projection, got %+v", user)
	}

	payload, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal Tieta user info: %v", err)
	}
	var fields map[string]any
	if err = json.Unmarshal(payload, &fields); err != nil {
		t.Fatalf("unmarshal Tieta user info: %v", err)
	}
	if len(fields) != 3 || fields["id"] != float64(13) || fields["customerName"] != "公安" || fields["phone"] != "18213268117" {
		t.Fatalf("expected JSON to contain only id, customerName and phone, got %s", payload)
	}
}
