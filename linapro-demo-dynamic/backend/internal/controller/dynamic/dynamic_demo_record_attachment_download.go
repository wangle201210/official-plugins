// Demo-record attachment download route controller.

package dynamic

import (
	"context"
	"fmt"

	v1 "lina-plugin-linapro-demo-dynamic/backend/api/dynamic/v1"

	bridgeguest "lina-core/pkg/plugin/pluginbridge/guest"
)

// DownloadDemoRecordAttachment streams one plugin-owned attachment file.
func (c *Controller) DownloadDemoRecordAttachment(
	ctx context.Context,
	req *v1.DownloadDemoRecordAttachmentReq,
) (res *v1.DownloadDemoRecordAttachmentRes, err error) {
	payload, err := c.dynamicSvc.BuildDemoRecordAttachmentDownload(req.Id)
	if err != nil {
		return nil, wrapDynamicError(err)
	}
	if err = bridgeguest.SetResponseHeader(
		ctx,
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, payload.OriginalName),
	); err != nil {
		return nil, err
	}
	if err = bridgeguest.WriteResponse(ctx, 200, payload.ContentType, payload.Body); err != nil {
		return nil, err
	}
	return nil, nil
}
