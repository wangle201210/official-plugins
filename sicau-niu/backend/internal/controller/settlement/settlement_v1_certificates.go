// settlement_v1_certificates.go implements the batch certificate issuance handler.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
)

// IssueCertificates batch-issues a certificate honor to its eligible cohort.
func (c *ControllerV1) IssueCertificates(ctx context.Context, req *v1.IssueCertificatesReq) (res *v1.IssueCertificatesRes, err error) {
	result, err := c.settlementSvc.IssueCertificates(ctx, req.HonorId)
	if err != nil {
		return nil, err
	}
	return &v1.IssueCertificatesRes{
		Eligible: result.Eligible,
		Issued:   result.Issued,
		Skipped:  result.Skipped,
	}, nil
}
