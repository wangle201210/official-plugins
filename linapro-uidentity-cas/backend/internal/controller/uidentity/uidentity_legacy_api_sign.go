// This file guards legacy third-party user self-service routes with the old
// uidentity/admin APIMiddleWare signature contract.

package uidentity

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"

	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacyAPISignGuard verifies the old appid/ts/sign request headers before
// dispatching legacy third-party user routes. Failures keep the old contract:
// HTTP 200 with a {code: 401, msg} JSON body.
func (c *LegacyController) LegacyAPISignGuard(r *ghttp.Request) {
	err := c.uidentitySvc.VerifyLegacyAPISignature(r.Context(), uidentitysvc.LegacyAPISignatureInput{
		AppID:     r.GetHeader("appid"),
		Timestamp: r.GetHeader("ts"),
		Sign:      r.GetHeader("sign"),
		Body:      r.GetBody(),
	})
	if err != nil {
		legacyAPIUnauthorized(r, err)
		return
	}
	r.Middleware.Next()
}

// legacyAPIUnauthorized writes the old APIUnauthorized envelope and stops the
// request: HTTP 200 with {code: 401, msg}.
func legacyAPIUnauthorized(r *ghttp.Request, err error) {
	msg := "签名有误"
	if err != nil && err.Error() != "" {
		msg = err.Error()
	}
	r.Response.WriteJson(map[string]any{
		"code": http.StatusUnauthorized,
		"msg":  msg,
	})
	r.ExitAll()
}
