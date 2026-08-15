// honor_certificate.go implements the player electronic-certificate generation. It
// verifies the target honor is a certificate type and that the requesting player
// already holds it (so a player can only obtain their own granted certificate),
// assembles the certificate fields (the player's nickname, the honor name/code and
// the campus badge) and renders the personalized PNG through the replaceable
// certificate renderer seam, returning the base64-encoded image together with the
// structured fields.

package honor

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// certificateHonorType is the persisted honor-type string selecting certificate
// honors, reusing the package's stable enum constant.
const certificateHonorType = string(HonorTypeCertificate)

// Certificate is the player electronic-certificate result: the structured fields
// the mini-program draws on canvas.
type Certificate struct {
	// Nickname is the holder nickname.
	Nickname string
	// HonorName is the certificate honor display name.
	HonorName string
	// HonorCode is the certificate honor unique code.
	HonorCode string
	// CampusBadge is the campus anniversary badge text; empty when unset.
	CampusBadge string
	// UnlockedAt is the grant time as Unix milliseconds; nil when unset.
	UnlockedAt *int64
}

// PlayerCertificate returns the certificate fields for a certificate the player holds.
func (s *serviceImpl) PlayerCertificate(ctx context.Context, playerID, honorID int64) (*Certificate, error) {
	if honorID <= 0 {
		return nil, bizerr.NewCode(CodeHonorIDRequired)
	}
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeCertificateNotOwned)
	}

	var honor *entitymodel.HonorDef
	if err := dao.HonorDef.Ctx(ctx).Where(dao.HonorDef.Columns().Id, honorID).Scan(&honor); err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	if honor == nil {
		return nil, bizerr.NewCode(CodeHonorNotFound)
	}
	if honor.HonorType != certificateHonorType {
		return nil, bizerr.NewCode(CodeHonorNotCertificate)
	}

	var grant *entitymodel.UserHonor
	err := dao.UserHonor.Ctx(ctx).
		Where(dao.UserHonor.Columns().UserId, playerID).
		Where(dao.UserHonor.Columns().HonorId, honorID).
		Scan(&grant)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	if grant == nil {
		return nil, bizerr.NewCode(CodeCertificateNotOwned)
	}

	nickname, err := s.playerNickname(ctx, playerID)
	if err != nil {
		return nil, err
	}
	campusBadge, err := s.certificateCampusBadge(ctx)
	if err != nil {
		return nil, err
	}

	return &Certificate{
		CampusBadge: campusBadge,
		Nickname:    nickname,
		HonorName:   honor.Name,
		HonorCode:   honor.Code,
		UnlockedAt:  apitime.Milli(grant.UnlockedAt),
	}, nil
}

// playerNickname loads the player's nickname in one projected query; an empty
// nickname is returned when the player row is missing.
func (s *serviceImpl) playerNickname(ctx context.Context, playerID int64) (string, error) {
	var player *entitymodel.User
	err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Nickname).
		Where(dao.User.Columns().Id, playerID).
		Scan(&player)
	if err != nil {
		return "", bizerr.WrapCode(err, CodeHonorQueryFailed)
	}
	if player == nil {
		return "", nil
	}
	return player.Nickname, nil
}

// certificateCampusBadge returns the operator-maintained badge text when rules
// are injected, otherwise the constructor fallback.
func (s *serviceImpl) certificateCampusBadge(ctx context.Context) (string, error) {
	if s.rulesSvc == nil {
		return s.campusBadge, nil
	}
	return s.rulesSvc.PosterCampusBadge(ctx)
}
