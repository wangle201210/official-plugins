// This file adapts legacy upload requests to the plugin-owned storage service,
// including multipart file extraction from the current GoFrame request.

package uidentity

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// LegacyUpload stores legacy upload payloads in plugin-owned local storage.
func (c *ControllerV1) LegacyUpload(ctx context.Context, req *v1.LegacyUploadReq) (res *v1.LegacyUploadRes, err error) {
	var files []*v1.LegacyUploadFile
	r := g.RequestFromCtx(ctx)
	out, err := c.uidentitySvc.UploadLegacyFiles(ctx, uidentitysvc.LegacyUploadInput{
		Type:        req.Type,
		Source:      req.Source,
		Base64File:  req.File,
		UploadFiles: r.GetUploadFiles("file"),
	})
	if err != nil {
		return nil, err
	}
	for _, file := range out.Files {
		files = append(files, &v1.LegacyUploadFile{
			Size:     file.Size,
			Path:     file.Path,
			FullPath: file.FullPath,
			Name:     file.Name,
			Type:     file.Type,
		})
	}
	return &v1.LegacyUploadRes{Files: files}, nil
}
