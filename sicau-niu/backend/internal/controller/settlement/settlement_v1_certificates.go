// settlement_v1_certificates.go implements the batch certificate issuance handler.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
)

// CertificateOptions returns the batch-issuable certificate honor selector list.
func (c *ControllerV1) CertificateOptions(ctx context.Context, _ *v1.CertificateOptionsReq) (res *v1.CertificateOptionsRes, err error) {
	options, err := c.settlementSvc.CertificateOptions(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.CertificateOption, 0, len(options))
	for _, option := range options {
		list = append(list, &v1.CertificateOption{
			Id:         option.Id,
			Code:       option.Code,
			Name:       option.Name,
			UnlockType: option.UnlockType,
			Threshold:  option.Threshold,
		})
	}
	return &v1.CertificateOptionsRes{List: list}, nil
}

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
