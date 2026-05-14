// This file implements the CMS site settings update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// SiteUpdate updates CMS site settings.
func (c *ControllerV1) SiteUpdate(ctx context.Context, req *v1.SiteUpdateReq) (res *v1.SiteUpdateRes, err error) {
	err = c.cmsSvc.UpdateSite(ctx, cmssvc.SiteUpdateInput{
		Name:        req.Name,
		Logo:        req.Logo,
		Weixin:      req.Weixin,
		Domain:      req.Domain,
		Slogan:      req.Slogan,
		Keywords:    req.Keywords,
		Description: req.Description,
		Icp:         req.Icp,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Status:      req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SiteUpdateRes{}, nil
}
