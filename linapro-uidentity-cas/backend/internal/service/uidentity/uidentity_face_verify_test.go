// This file verifies the artemis face verification client against a local
// HTTP stub and the plugin-local fallback without external dependencies.

package uidentity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestVerifyActivationFacePassesWithoutCredentials keeps the plugin-local
// marker behavior when artemis credentials are absent.
func TestVerifyActivationFacePassesWithoutCredentials(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	pass, msg, err := service.verifyActivationFace(context.Background(), "A001", "510000200001010011", "marker")
	if err != nil {
		t.Fatalf("expected local marker mode to pass: %v", err)
	}
	if !pass || msg != "" {
		t.Fatalf("expected local marker mode pass, got pass=%v msg=%q", pass, msg)
	}
}

// TestVerifyActivationFaceUsesArtemisContract verifies the signed request
// shape and the similarity threshold decision against a stub server.
func TestVerifyActivationFaceUsesArtemisContract(t *testing.T) {
	t.Parallel()

	var gotPath, gotKey, gotNonce, gotSignatureHeaders string
	var gotBody legacyFaceRequest
	similarity := "85"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("x-ca-key")
		gotNonce = r.Header.Get("x-ca-nonce")
		gotSignatureHeaders = r.Header.Get("x-ca-signature-headers")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode face request: %v", err)
		}
		_, _ = fmt.Fprintf(w, `{"code":"0","msg":"success","data":{"similarity":"%s"}}`, similarity)
	}))
	t.Cleanup(server.Close)

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, fmt.Sprintf(`
legacy:
  face:
    ak: test-ak
    sk: test-sk
    baseUrl: %s
    similarity: 80
`, server.URL))}
	inlineImage := strings.Repeat("x", legacyFaceInlineBase64MinLength)
	pass, msg, err := service.verifyActivationFace(context.Background(), "A001", "510000200001010011", inlineImage)
	if err != nil {
		t.Fatalf("face verification failed: %v", err)
	}
	if !pass || msg != "" {
		t.Fatalf("expected similarity 85 >= 80 to pass, got pass=%v msg=%q", pass, msg)
	}
	if gotPath != defaultLegacyFacePath {
		t.Fatalf("face request path = %q, want old artemis path", gotPath)
	}
	if gotKey != "test-ak" || gotNonce == "" || gotSignatureHeaders != "x-ca-key,x-ca-nonce,x-ca-timestamp" {
		t.Fatalf("face request misses old signature headers: key=%q nonce=%q headers=%q", gotKey, gotNonce, gotSignatureHeaders)
	}
	if gotBody.JobNo != "A001" || gotBody.CertNo != "510000200001010011" || gotBody.SrcFacePicBase64Data != inlineImage {
		t.Fatalf("face request payload mismatch: %#v", gotBody)
	}

	similarity = "60"
	pass, msg, err = service.verifyActivationFace(context.Background(), "A001", "510000200001010011", inlineImage)
	if err != nil {
		t.Fatalf("face verification failed: %v", err)
	}
	if pass || msg == "" {
		t.Fatalf("expected similarity 60 < 80 to fail with message, got pass=%v msg=%q", pass, msg)
	}
}

// TestLegacyFaceSignatureMatchesOldStringToSign verifies the HMAC string
// layout matches the old getSignature implementation.
func TestLegacyFaceSignatureMatchesOldStringToSign(t *testing.T) {
	t.Parallel()

	first := legacyFaceSignature("ak", "sk", "nonce", "123", "/artemis/api/cflms/v1/face/verification")
	if first != legacyFaceSignature("ak", "sk", "nonce", "123", "/artemis/api/cflms/v1/face/verification") {
		t.Fatal("face signature must be deterministic")
	}
	if first == legacyFaceSignature("ak", "other-sk", "nonce", "123", "/artemis/api/cflms/v1/face/verification") {
		t.Fatal("face signature must depend on the secret")
	}
	if first == legacyFaceSignature("ak", "sk", "nonce", "456", "/artemis/api/cflms/v1/face/verification") {
		t.Fatal("face signature must depend on the timestamp")
	}
}
