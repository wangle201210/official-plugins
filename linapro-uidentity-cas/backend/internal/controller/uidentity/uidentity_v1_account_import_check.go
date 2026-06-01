package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// AccountImportCheck validates one account import workbook.
func (c *ControllerV1) AccountImportCheck(ctx context.Context, req *v1.AccountImportCheckReq) (res *v1.AccountImportCheckRes, err error) {
	out, err := c.uidentitySvc.CheckAccountImport(ctx, uidentitysvc.AccountImportInput{
		Filepath: req.Filepath,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &v1.AccountImportCheckRes{Rows: out.Rows}, nil
}
