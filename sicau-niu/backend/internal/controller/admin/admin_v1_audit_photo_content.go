// admin_v1_audit_photo_content.go implements protected evidence content reads.
package admin

import (
	"context"
	"encoding/base64"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

func (c *ControllerV1) AuditPhotoContent(ctx context.Context, req *v1.AuditPhotoContentReq) (*v1.AuditPhotoContentRes, error) {
	out, err := c.photoSvc.ContentForAudit(ctx, req.PhotoId)
	if err != nil {
		return nil, err
	}
	return &v1.AuditPhotoContentRes{ContentType: out.ContentType, SizeBytes: out.SizeBytes, ImageBase64: base64.StdEncoding.EncodeToString(out.Bytes)}, nil
}
