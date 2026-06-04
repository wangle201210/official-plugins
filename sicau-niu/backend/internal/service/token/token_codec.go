// token_codec.go implements signing and verification of the compact
// `header.payload.signature` player session token using standard-library
// HMAC-SHA256 and base64url(JSON) encoding. The header is a fixed constant; the
// payload carries the player ID and expiry. Verification is constant-time over
// the signature and rejects tampered or expired tokens.

package token

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/bizerr"
)

// tokenHeaderSegment is the fixed, pre-encoded token header declaring the
// HMAC-SHA256 algorithm and the plugin token type. Keeping it constant avoids
// re-encoding on every sign call and pins the accepted algorithm.
const tokenHeaderSegment = "eyJhbGciOiJIUzI1NiIsInR5cCI6IlNJQ0FVLU5JVS1QTEFZRVIifQ"

// tokenPayload is the JSON body embedded in the token. Field names are kept
// short to bound token size.
type tokenPayload struct {
	PlayerID int64 `json:"pid"` // PlayerID is the authenticated player ID.
	ExpireAt int64 `json:"exp"` // ExpireAt is the expiry instant as Unix seconds.
}

// Sign issues a signed player session token for playerID.
func (s *serviceImpl) Sign(ctx context.Context, playerID int64) (string, error) {
	if playerID <= 0 {
		return "", bizerr.WrapCode(gerror.New("player id must be positive"), CodeTokenSignFailed)
	}

	payload := tokenPayload{
		PlayerID: playerID,
		ExpireAt: s.now().Add(s.ttl).Unix(),
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeTokenSignFailed)
	}

	payloadSegment := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := tokenHeaderSegment + "." + payloadSegment
	signatureSegment := s.sign(signingInput)
	return signingInput + "." + signatureSegment, nil
}

// Verify validates token signature and expiry and returns the embedded player ID.
func (s *serviceImpl) Verify(_ context.Context, token string) (int64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != tokenHeaderSegment {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := s.sign(signingInput)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}
	var payload tokenPayload
	if err = json.Unmarshal(payloadBytes, &payload); err != nil {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}
	if payload.PlayerID <= 0 {
		return 0, bizerr.NewCode(CodeTokenInvalid)
	}
	if time.Unix(payload.ExpireAt, 0).Before(s.now()) {
		return 0, bizerr.NewCode(CodeTokenExpired)
	}
	return payload.PlayerID, nil
}

// sign computes the base64url-encoded HMAC-SHA256 signature for signingInput.
func (s *serviceImpl) sign(signingInput string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
