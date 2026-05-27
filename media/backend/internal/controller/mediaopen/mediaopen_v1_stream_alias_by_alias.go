// This file implements the public mediaopen stream-alias lookup endpoint.

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// GetStreamAliasByAlias returns one stream alias config by alias value.
func (c *ControllerV1) GetStreamAliasByAlias(ctx context.Context, req *v1.GetStreamAliasByAliasReq) (res *v1.GetStreamAliasByAliasRes, err error) {
	out, err := c.mediaSvc.GetAliasByAlias(ctx, req.Alias)
	if err != nil {
		return nil, err
	}
	return aliasOutputToPublicConfig(out), nil
}

// aliasOutputToPublicConfig maps the service alias projection into the public DTO.
func aliasOutputToPublicConfig(out *mediasvc.AliasOutput) *v1.GetStreamAliasByAliasRes {
	if out == nil {
		return nil
	}
	return &v1.GetStreamAliasByAliasRes{
		Id:         out.Id,
		Alias:      out.Alias,
		AutoRemove: out.AutoRemove,
		StreamPath: out.StreamPath,
		DeviceId:   out.DeviceId,
		ChannelId:  out.ChannelId,
		CreateTime: out.CreateTime,
	}
}
