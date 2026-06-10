// This file implements the old uidentity/admin third-party API signature
// contract: md5(body + appid + secret + ts) sent as request headers and bound
// to the applications table secret with a short anti-replay window.

package uidentity

import (
	"context"
	"crypto/md5"
	"crypto/subtle"
	"fmt"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
)

const (
	configKeyLegacyAPISignRequired      = "legacy.api.signRequired"
	configKeyLegacyAPISignWindowSeconds = "legacy.api.signWindowSeconds"

	// defaultLegacyAPISignWindowSeconds matches the old APIMiddleWare 5-second
	// replay window.
	defaultLegacyAPISignWindowSeconds = 5
)

// VerifyLegacyAPISignature enforces the old appid/ts/sign header contract for
// legacy third-party user self-service routes.
func (s *serviceImpl) VerifyLegacyAPISignature(ctx context.Context, in LegacyAPISignatureInput) error {
	required, err := s.configSvc.Bool(ctx, configKeyLegacyAPISignRequired, true)
	if err != nil {
		return err
	}
	if !required {
		return nil
	}
	appID := strings.TrimSpace(in.AppID)
	tsValue := strings.TrimSpace(in.Timestamp)
	sign := strings.TrimSpace(in.Sign)
	if appID == "" || tsValue == "" || sign == "" {
		return bizerr.NewCode(CodeAPISignatureInvalid)
	}
	ts, err := strconv.ParseInt(tsValue, 10, 64)
	if err != nil {
		return bizerr.NewCode(CodeAPISignatureInvalid)
	}
	app, err := s.runtimeApplicationByClientID(ctx, appID)
	if err != nil {
		return err
	}
	expected := legacyAPISignature(appID, tsValue, app.SecretKey, in.Body)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(sign)) != 1 {
		return bizerr.NewCode(CodeAPISignatureInvalid)
	}
	window, err := s.configSvc.Int(ctx, configKeyLegacyAPISignWindowSeconds, defaultLegacyAPISignWindowSeconds)
	if err != nil {
		return err
	}
	if !legacyAPISignatureFresh(ts, time.Now(), window) {
		return bizerr.NewCode(CodeAPISignatureExpired)
	}
	return nil
}

// legacyAPISignature computes the old md5(body + appid + secret + ts) digest.
func legacyAPISignature(appID string, ts string, secret string, body []byte) string {
	h := md5.New()
	h.Write(body)
	h.Write([]byte(appID))
	h.Write([]byte(secret))
	h.Write([]byte(ts))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// legacyAPISignatureFresh applies the old replay rule: the timestamp must not
// be in the future and must be at most windowSeconds old.
func legacyAPISignatureFresh(ts int64, now time.Time, windowSeconds int) bool {
	if windowSeconds <= 0 {
		windowSeconds = defaultLegacyAPISignWindowSeconds
	}
	delta := now.Unix() - ts
	return delta >= 0 && delta <= int64(windowSeconds)
}
