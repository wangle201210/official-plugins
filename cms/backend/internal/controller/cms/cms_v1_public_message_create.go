// This file implements the public CMS visitor message submission controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"

	"github.com/gogf/gf/v2/net/ghttp"
)

// PublicMessageCreate creates a public CMS visitor message.
func (c *ControllerV1) PublicMessageCreate(ctx context.Context, req *v1.PublicMessageCreateReq) (res *v1.PublicMessageCreateRes, err error) {
	request := ghttp.RequestFromCtx(ctx)
	userIP := ""
	userAgent := ""
	if request != nil {
		userIP = request.GetClientIp()
		userAgent = request.Header.Get("User-Agent")
	}
	id, err := c.cmsSvc.CreatePublicMessage(ctx, cmssvc.PublicMessageCreateInput{
		Name:      req.Name,
		Mobile:    req.Mobile,
		Email:     req.Email,
		Content:   req.Content,
		UserIp:    userIP,
		UserAgent: userAgent,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PublicMessageCreateRes{Id: id}, nil
}
