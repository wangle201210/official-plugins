// identity_phone.go implements phone authorization binding: decode the phone via
// the WeChat gateway, enforce the one-phone-one-account uniqueness constraint,
// and record the device fingerprint on the current player's own row.

package identity

import (
	"context"
	"strings"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	"lina-plugin-sicau-niu/backend/internal/service/wechat"
)

// BindPhoneInput defines the phone authorization input. It carries the WeChat
// phone-authorization payload plus the captured device fingerprint.
type BindPhoneInput struct {
	// Code is the new-style getPhoneNumber authorization code.
	Code string
	// EncryptedData is the legacy encrypted phone payload.
	EncryptedData string
	// IV is the legacy decryption initialization vector.
	IV string
	// PhoneOverride is the explicit phone consumed by the mock gateway flow.
	PhoneOverride string
	// DeviceFingerprint is the lightweight device fingerprint recorded for risk
	// control. It may be empty.
	DeviceFingerprint string
}

// BindPhone binds a WeChat-authorized phone number to the current player.
func (s *serviceImpl) BindPhone(ctx context.Context, playerID int64, in *BindPhoneInput) error {
	if playerID <= 0 {
		return bizerr.NewCode(CodePlayerNotFound)
	}
	if in == nil {
		return bizerr.NewCode(CodePhoneRequired)
	}

	phone, err := s.gateway.DecodePhone(ctx, wechat.DecodePhoneInput{
		Code:          strings.TrimSpace(in.Code),
		EncryptedData: strings.TrimSpace(in.EncryptedData),
		IV:            strings.TrimSpace(in.IV),
		PhoneOverride: strings.TrimSpace(in.PhoneOverride),
	})
	if err != nil {
		return err
	}
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return bizerr.NewCode(CodePhoneRequired)
	}

	taken, err := s.phoneTaken(ctx, phone, playerID)
	if err != nil {
		return err
	}
	if taken {
		return bizerr.NewCode(CodePhoneTaken)
	}

	_, err = dao.User.Ctx(ctx).
		Where(do.User{Id: playerID}).
		Data(do.User{
			Phone:             phone,
			DeviceFingerprint: strings.TrimSpace(in.DeviceFingerprint),
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodePlayerWriteFailed)
	}
	return nil
}

// phoneTaken reports whether phone is bound to a player other than playerID.
func (s *serviceImpl) phoneTaken(ctx context.Context, phone string, playerID int64) (bool, error) {
	count, err := dao.User.Ctx(ctx).
		Where(do.User{Phone: phone}).
		WhereNot(dao.User.Columns().Id, playerID).
		Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodePlayerQueryFailed)
	}
	return count > 0, nil
}
