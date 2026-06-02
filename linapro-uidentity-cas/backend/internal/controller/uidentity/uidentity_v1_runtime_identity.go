// This file resolves the old runtime account identity for generated V1
// controller methods. The old uidentity/admin API middleware injected the
// account number from request headers, so runtime endpoints must not require
// frontends to send number in JSON bodies.

package uidentity

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

func (c *ControllerV1) runtimeNumber(ctx context.Context) (string, error) {
	r := g.RequestFromCtx(ctx)
	if r != nil {
		for _, header := range []string{"X-Uidentity-Number", "X-Account-Number", "X-User-Number", "number"} {
			if value := strings.TrimSpace(r.GetHeader(header)); value != "" {
				return value, nil
			}
		}
		for _, key := range []string{"number", "userNumber", "runtimeNumber"} {
			if value := strings.TrimSpace(r.GetCtxVar(key).String()); value != "" {
				return value, nil
			}
		}
		if accessToken := legacyAccessToken(r); accessToken != "" {
			out, err := c.uidentitySvc.GetUserInfoByRuntimeToken(ctx, accessToken)
			if err != nil {
				return "", err
			}
			if out != nil && out.User != nil && out.User.Number != "" {
				return out.User.Number, nil
			}
		}
	}
	return "", bizerr.NewCode(uidentitysvc.CodeInvalidCredentials)
}
