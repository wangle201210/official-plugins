// player_v1_certificate.go implements the player electronic-certificate handler. It
// resolves the authenticated player from the request context and delegates to the
// honor service, which enforces the certificate-type and ownership checks.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// Certificate returns the personalized electronic certificate the player holds.
func (c *ControllerV1) Certificate(ctx context.Context, req *v1.CertificateReq) (res *v1.CertificateRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	certificate, err := c.honorSvc.PlayerCertificate(ctx, playerID, req.HonorId)
	if err != nil {
		return nil, err
	}
	return &v1.CertificateRes{
		Nickname:    certificate.Nickname,
		HonorName:   certificate.HonorName,
		HonorCode:   certificate.HonorCode,
		CampusBadge: certificate.CampusBadge,
		UnlockedAt:  certificate.UnlockedAt,
	}, nil
}
