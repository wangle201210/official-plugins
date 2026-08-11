package player

import (
	"context"
	"encoding/base64"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) PhotoContent(ctx context.Context, req *v1.PhotoContentReq) (res *v1.PhotoContentRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.photoSvc.Content(ctx, playerID, req.PhotoId)
	if err != nil {
		return nil, err
	}
	return &v1.PhotoContentRes{ContentType: out.ContentType, SizeBytes: out.SizeBytes, ImageBase64: base64.StdEncoding.EncodeToString(out.Bytes)}, nil
}
