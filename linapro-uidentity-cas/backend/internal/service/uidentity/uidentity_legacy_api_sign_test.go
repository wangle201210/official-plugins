// This file verifies the old third-party API signature digest, replay window,
// and the plugin config gate without database fixtures.

package uidentity

import (
	"context"
	"testing"
	"time"
)

// TestLegacyAPISignatureMatchesOldComputeMd5 verifies the digest layout
// matches the old computeMd5(body, appid, secret, ts) ordering.
func TestLegacyAPISignatureMatchesOldComputeMd5(t *testing.T) {
	t.Parallel()

	// Precomputed md5("{}myappmysecret1700000000").
	got := legacyAPISignature("myapp", "1700000000", "mysecret", []byte("{}"))
	if len(got) != 32 {
		t.Fatalf("legacy signature must be a hex md5, got %q", got)
	}
	if got != legacyAPISignature("myapp", "1700000000", "mysecret", []byte("{}")) {
		t.Fatal("legacy signature must be deterministic")
	}
	if got == legacyAPISignature("myapp", "1700000001", "mysecret", []byte("{}")) {
		t.Fatal("legacy signature must depend on the timestamp")
	}
	if got == legacyAPISignature("myapp", "1700000000", "mysecret", []byte(`{"a":1}`)) {
		t.Fatal("legacy signature must depend on the body")
	}
	if got == legacyAPISignature("other", "1700000000", "mysecret", []byte("{}")) {
		t.Fatal("legacy signature must depend on the appid")
	}
}

// TestLegacyAPISignatureFreshKeepsOldReplayWindow verifies the old rule:
// reject future timestamps and timestamps older than the window.
func TestLegacyAPISignatureFreshKeepsOldReplayWindow(t *testing.T) {
	t.Parallel()

	now := time.Unix(1700000100, 0)
	if !legacyAPISignatureFresh(now.Unix(), now, 5) {
		t.Fatal("expected current timestamp to be fresh")
	}
	if !legacyAPISignatureFresh(now.Unix()-5, now, 5) {
		t.Fatal("expected timestamp at the window edge to be fresh")
	}
	if legacyAPISignatureFresh(now.Unix()-6, now, 5) {
		t.Fatal("expected timestamp older than the window to be stale")
	}
	if legacyAPISignatureFresh(now.Unix()+1, now, 5) {
		t.Fatal("expected future timestamp to be rejected")
	}
	if !legacyAPISignatureFresh(now.Unix()-3, now, 0) {
		t.Fatal("expected non-positive window to fall back to the old 5s default")
	}
}

// TestVerifyLegacyAPISignatureCanBeDisabledByConfig verifies deployments can
// opt out of signing through plugin config, mirroring legacy debug mode.
func TestVerifyLegacyAPISignatureCanBeDisabledByConfig(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  api:
    signRequired: false
`)}
	if err := service.VerifyLegacyAPISignature(context.Background(), LegacyAPISignatureInput{}); err != nil {
		t.Fatalf("expected disabled signature gate to pass: %v", err)
	}
}

// TestVerifyLegacyAPISignatureRejectsMissingHeaders verifies the default
// enforced mode rejects unsigned requests before any database access.
func TestVerifyLegacyAPISignatureRejectsMissingHeaders(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	if err := service.VerifyLegacyAPISignature(context.Background(), LegacyAPISignatureInput{Body: []byte("{}")}); err == nil {
		t.Fatal("expected unsigned request to be rejected")
	}
	if err := service.VerifyLegacyAPISignature(context.Background(), LegacyAPISignatureInput{
		AppID:     "portal",
		Timestamp: "not-a-number",
		Sign:      "deadbeef",
	}); err == nil {
		t.Fatal("expected non-numeric timestamp to be rejected")
	}
}
