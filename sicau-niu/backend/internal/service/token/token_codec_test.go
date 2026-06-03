// token_codec_test.go verifies the player session token sign/verify round trip,
// expiry handling, tamper rejection, malformed-token rejection and constructor
// validation. All assertions are pure-logic and require no database. Each test
// constructs its own service instance and, when it overrides the injectable
// clock, restores it via the local instance scope so tests stay self-contained
// and order independent.

package token

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// signPayloadForTest encodes a fresh payload segment for the given player ID,
// reusing the production payload encoding so the forged-payload test exercises
// the real verification path. It lives in the test file only.
func (s *serviceImpl) signPayloadForTest(t *testing.T, playerID int64) string {
	t.Helper()
	payloadBytes, err := json.Marshal(tokenPayload{
		PlayerID: playerID,
		ExpireAt: s.now().Add(s.ttl).Unix(),
	})
	if err != nil {
		t.Fatalf("marshal forged payload failed: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(payloadBytes)
}

// newTestService builds a token service with a fixed secret and TTL for tests.
// It fails the test when construction returns an unexpected error.
func newTestService(t *testing.T, ttl time.Duration) *serviceImpl {
	t.Helper()
	svc, err := New(Config{Secret: "unit-test-secret", TTL: ttl})
	if err != nil {
		t.Fatalf("construct token service failed: %v", err)
	}
	impl, ok := svc.(*serviceImpl)
	if !ok {
		t.Fatalf("expected *serviceImpl, got %T", svc)
	}
	return impl
}

// TestSignVerifyRoundTripReturnsSamePlayerID verifies a signed token verifies
// back to the same player ID.
func TestSignVerifyRoundTripReturnsSamePlayerID(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Hour)

	const playerID int64 = 4242
	token, err := svc.Sign(ctx, playerID)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("expected header.payload.signature token, got %q", token)
	}

	got, err := svc.Verify(ctx, token)
	if err != nil {
		t.Fatalf("verify token failed: %v", err)
	}
	if got != playerID {
		t.Fatalf("expected player ID %d, got %d", playerID, got)
	}
}

// TestSignRejectsNonPositivePlayerID verifies signing a non-positive player ID
// returns CodeTokenSignFailed and an empty token.
func TestSignRejectsNonPositivePlayerID(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Hour)

	for _, playerID := range []int64{0, -1} {
		token, err := svc.Sign(ctx, playerID)
		if err == nil {
			t.Fatalf("expected error signing player ID %d", playerID)
		}
		if token != "" {
			t.Fatalf("expected empty token on sign failure, got %q", token)
		}
		bizErr, ok := bizerr.As(err)
		if !ok {
			t.Fatalf("expected structured business error, got %T", err)
		}
		if bizErr.RuntimeCode() != CodeTokenSignFailed.RuntimeCode() {
			t.Fatalf("expected %s, got %s", CodeTokenSignFailed.RuntimeCode(), bizErr.RuntimeCode())
		}
	}
}

// TestVerifyRejectsExpiredToken verifies an expired token fails with
// CodeTokenExpired. The token is signed in the past via the injectable clock so
// the test does not sleep.
func TestVerifyRejectsExpiredToken(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Minute)

	// Pin signing time one hour in the past so the one-minute TTL has lapsed.
	pastInstant := time.Now().Add(-time.Hour)
	svc.now = func() time.Time { return pastInstant }
	token, err := svc.Sign(ctx, 7)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}

	// Restore the clock to wall time for verification.
	svc.now = time.Now
	_, err = svc.Verify(ctx, token)
	if err == nil {
		t.Fatal("expected expired token to fail verification")
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if bizErr.RuntimeCode() != CodeTokenExpired.RuntimeCode() {
		t.Fatalf("expected %s, got %s", CodeTokenExpired.RuntimeCode(), bizErr.RuntimeCode())
	}
}

// TestVerifyRejectsTamperedSignature verifies that mutating the signature
// segment makes verification fail with CodeTokenInvalid.
func TestVerifyRejectsTamperedSignature(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Hour)

	token, err := svc.Sign(ctx, 11)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected three token segments, got %d", len(parts))
	}
	// Flip the last signature character to a different base64url symbol.
	sig := []byte(parts[2])
	if sig[len(sig)-1] == 'A' {
		sig[len(sig)-1] = 'B'
	} else {
		sig[len(sig)-1] = 'A'
	}
	tampered := parts[0] + "." + parts[1] + "." + string(sig)

	_, err = svc.Verify(ctx, tampered)
	if err == nil {
		t.Fatal("expected tampered token to fail verification")
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if bizErr.RuntimeCode() != CodeTokenInvalid.RuntimeCode() {
		t.Fatalf("expected %s, got %s", CodeTokenInvalid.RuntimeCode(), bizErr.RuntimeCode())
	}
}

// TestVerifyRejectsTamperedPayload verifies that mutating the payload segment so
// the embedded player ID changes is rejected because the signature no longer
// matches the altered payload.
func TestVerifyRejectsTamperedPayload(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Hour)

	token, err := svc.Sign(ctx, 11)
	if err != nil {
		t.Fatalf("sign token failed: %v", err)
	}
	parts := strings.Split(token, ".")

	// Re-sign a different payload but keep the original signature: the HMAC over
	// the new signing input will no longer equal the attached signature.
	forgedPayload := svc.signPayloadForTest(t, 99)
	forged := parts[0] + "." + forgedPayload + "." + parts[2]

	_, err = svc.Verify(ctx, forged)
	if err == nil {
		t.Fatal("expected payload-tampered token to fail verification")
	}
	bizErr, ok := bizerr.As(err)
	if !ok {
		t.Fatalf("expected structured business error, got %T", err)
	}
	if bizErr.RuntimeCode() != CodeTokenInvalid.RuntimeCode() {
		t.Fatalf("expected %s, got %s", CodeTokenInvalid.RuntimeCode(), bizErr.RuntimeCode())
	}
}

// TestVerifyRejectsEmptyAndMalformedTokens verifies blank and structurally
// invalid tokens fail with CodeTokenInvalid.
func TestVerifyRejectsEmptyAndMalformedTokens(t *testing.T) {
	ctx := context.Background()
	svc := newTestService(t, time.Hour)

	cases := []string{
		"",
		"   ",
		"onlyonesegment",
		"two.segments",
		"a.b.c.d",
		"wrong-header.payload.signature",
		tokenHeaderSegment + ".!!!notbase64!!!.signature",
	}
	for _, token := range cases {
		_, err := svc.Verify(ctx, token)
		if err == nil {
			t.Fatalf("expected malformed token %q to fail verification", token)
		}
		bizErr, ok := bizerr.As(err)
		if !ok {
			t.Fatalf("token %q: expected structured business error, got %T", token, err)
		}
		if bizErr.RuntimeCode() != CodeTokenInvalid.RuntimeCode() {
			t.Fatalf("token %q: expected %s, got %s", token, CodeTokenInvalid.RuntimeCode(), bizErr.RuntimeCode())
		}
	}
}

// TestNewRejectsInvalidConfig verifies the constructor returns configuration
// bizerrs for an empty secret or a non-positive TTL.
func TestNewRejectsInvalidConfig(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "empty secret", cfg: Config{Secret: "", TTL: time.Hour}, want: CodeTokenSecretMissing.RuntimeCode()},
		{name: "zero ttl", cfg: Config{Secret: "secret", TTL: 0}, want: CodeTokenTTLInvalid.RuntimeCode()},
		{name: "negative ttl", cfg: Config{Secret: "secret", TTL: -time.Second}, want: CodeTokenTTLInvalid.RuntimeCode()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := New(tc.cfg)
			if err == nil {
				t.Fatal("expected constructor error")
			}
			if svc != nil {
				t.Fatalf("expected nil service on failure, got %#v", svc)
			}
			bizErr, ok := bizerr.As(err)
			if !ok {
				t.Fatalf("expected structured business error, got %T", err)
			}
			if bizErr.RuntimeCode() != tc.want {
				t.Fatalf("expected %s, got %s", tc.want, bizErr.RuntimeCode())
			}
		})
	}
}
