package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

// AccountImport imports account rows from one workbook.
func (c *ControllerV1) AccountImport(ctx context.Context, req *v1.AccountImportReq) (res *v1.AccountImportRes, err error) {
	out, err := c.uidentitySvc.ImportAccounts(ctx, uidentitysvc.AccountImportInput{
		Filepath: req.Filepath,
		Limit:    req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return &v1.AccountImportRes{Success: out.Success, FailedNumber: out.FailedNumber}, nil
}
