package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// ActivationStart starts an account activation challenge.
func (c *ControllerV1) ActivationStart(ctx context.Context, req *v1.ActivationStartReq) (res *v1.ActivationStartRes, err error) {
	// The old activation entry verified a captcha first, accepting the
	// rotating common pass in place of the captcha code.
	if err := c.uidentitySvc.VerifyLoginCaptcha(ctx, req.Code, req.UUID); err != nil {
		return nil, err
	}
	out, err := c.uidentitySvc.StartActivation(ctx, uidentitysvc.ActivationStartInput{
		Number: req.Number,
		Name:   req.Name,
		Idcard: req.Idcard,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ActivationStartRes{UUID: out.ChallengeID, Face: out.NeedFace}, nil
}
