// This file implements the HotGo-compatible token and device strategy endpoint.

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// UserDeviceStrategyByToken resolves one device strategy through the HotGo-compatible token endpoint.
func (c *ControllerV1) UserDeviceStrategyByToken(ctx context.Context, req *v1.UserDeviceStrategyByTokenReq) (res *v1.UserDeviceStrategyByTokenRes, err error) {
	out, err := c.mediaSvc.UserDeviceStrategyByToken(ctx, mediasvc.UserDeviceStrategyByTokenInput{
		Token:    req.Token,
		DeviceId: req.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UserDeviceStrategyByTokenRes{
		UserInfo:   buildCompatTietaUserInfo(out.UserInfo),
		HasAccess:  out.HasAccess,
		StrategyId: out.StrategyId,
		Strategy:   buildCompatStrategyInfo(out.Strategy),
	}, nil
}

// buildCompatTietaUserInfo maps the service Tieta identity into the compatibility DTO.
func buildCompatTietaUserInfo(user *mediasvc.TietaUser) *v1.TietaUserInfo {
	if user == nil {
		return nil
	}
	return &v1.TietaUserInfo{
		Id:           user.Id,
		DeptId:       user.DeptId,
		Username:     user.Username,
		RealName:     user.RealName,
		Mobile:       user.Mobile,
		UserType:     user.UserType,
		CustomerCode: user.CustomerCode,
		TenantID:     user.TenantId,
		DeptName:     user.DeptName,
		RegionCode:   user.RegionCode,
		OrgId:        user.OrgId,
		Enable:       user.Enable,
	}
}

// buildCompatStrategyInfo maps the service strategy projection into the compatibility DTO.
func buildCompatStrategyInfo(strategy *mediasvc.StrategyInfoOutput) *v1.StrategyInfo {
	if strategy == nil {
		return nil
	}
	return &v1.StrategyInfo{
		Id:              strategy.Id,
		Name:            strategy.Name,
		StrategyContent: strategy.StrategyContent,
	}
}
