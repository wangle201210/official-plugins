package player

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	"lina-core/pkg/closeutil"
	"lina-plugin-sicau-niu/backend/api/player/v1"
	activationphotosvc "lina-plugin-sicau-niu/backend/internal/service/activationphoto"
)

func (c *ControllerV1) UploadPhoto(ctx context.Context, req *v1.UploadPhotoReq) (res *v1.UploadPhotoRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	upload := g.RequestFromCtx(ctx).GetUploadFile("file")
	if upload == nil {
		return nil, bizerr.NewCode(activationphotosvc.CodePhotoRequired)
	}
	reader, err := upload.Open()
	if err != nil {
		return nil, bizerr.NewCode(activationphotosvc.CodePhotoInvalid)
	}
	defer closeutil.Close(ctx, reader, &err, "close uploaded activation photo failed")
	out, err := c.photoSvc.Upload(ctx, playerID, &activationphotosvc.UploadInput{
		RequestID: req.RequestId, Filename: upload.Filename, ContentType: upload.Header.Get("Content-Type"), SizeBytes: upload.Size, Reader: reader,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UploadPhotoRes{PhotoId: out.Token, ContentType: out.ContentType, SizeBytes: out.SizeBytes}, nil
}
