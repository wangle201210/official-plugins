// =================================================================================
// This is the aggregated interface for the sicau-niu activity-record query API. It
// mirrors the GoFrame controller-interface convention so the record controller can
// be bound on the host operator router like the admin controller.
// =================================================================================

package record

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/record/v1"
)

// IRecordV1 is the operator activity-record query contract: the feeding, steal,
// gift, check-in, activation and grass-ledger read-only paged queries.
type IRecordV1 interface {
	Feedings(ctx context.Context, req *v1.FeedingsReq) (res *v1.FeedingsRes, err error)
	Steals(ctx context.Context, req *v1.StealsReq) (res *v1.StealsRes, err error)
	Gifts(ctx context.Context, req *v1.GiftsReq) (res *v1.GiftsRes, err error)
	Checkins(ctx context.Context, req *v1.CheckinsReq) (res *v1.CheckinsRes, err error)
	Activations(ctx context.Context, req *v1.ActivationsReq) (res *v1.ActivationsRes, err error)
	ActivationAttempts(ctx context.Context, req *v1.ActivationAttemptsReq) (res *v1.ActivationAttemptsRes, err error)
	GrassTxns(ctx context.Context, req *v1.GrassTxnsReq) (res *v1.GrassTxnsRes, err error)
}
