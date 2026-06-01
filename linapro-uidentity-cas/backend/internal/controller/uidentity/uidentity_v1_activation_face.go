package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationFace records an activation face proof marker.
func (c *ControllerV1) ActivationFace(ctx context.Context, req *v1.ActivationFaceReq) (res *v1.ActivationFaceRes, err error) {
	out, err := c.uidentitySvc.RecordActivationFace(ctx, uidentitysvc.ActivationFaceInput{
		ChallengeID: req.ChallengeId,
		FaceURL:     req.FaceUrl,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStepRes{ChallengeId: out.ChallengeID, Success: out.Success}, nil
}
