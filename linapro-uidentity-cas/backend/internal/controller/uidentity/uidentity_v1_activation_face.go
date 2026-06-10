package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationFace verifies and records an activation face proof.
func (c *ControllerV1) ActivationFace(ctx context.Context, req *v1.ActivationFaceReq) (res *v1.ActivationFaceRes, err error) {
	out, err := c.uidentitySvc.RecordActivationFace(ctx, uidentitysvc.ActivationFaceInput{
		ChallengeID: req.ChallengeId,
		FaceURL:     req.FaceUrl,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationFaceStepRes{UUID: out.ChallengeID, Pass: out.Success, Msg: out.Message}, nil
}
