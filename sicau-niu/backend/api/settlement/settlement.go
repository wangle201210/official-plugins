// =================================================================================
// This is the aggregated interface for the sicau-niu operator settlement API. It
// mirrors the GoFrame controller-interface convention so the settlement controller
// can be bound on the host operator router like the admin controller.
// =================================================================================

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
)

// ISettlementV1 is the operator settlement controller contract: the dashboard, the
// roster export, the batch certificate issuance, the shared-device risk view and
// the settlement archive (create and list).
type ISettlementV1 interface {
	Dashboard(ctx context.Context, req *v1.DashboardReq) (res *v1.DashboardRes, err error)
	ExportPlayers(ctx context.Context, req *v1.ExportPlayersReq) (res *v1.ExportPlayersRes, err error)
	IssueCertificates(ctx context.Context, req *v1.IssueCertificatesReq) (res *v1.IssueCertificatesRes, err error)
	RiskDeviceClusters(ctx context.Context, req *v1.RiskDeviceClustersReq) (res *v1.RiskDeviceClustersRes, err error)
	CreateArchive(ctx context.Context, req *v1.CreateArchiveReq) (res *v1.CreateArchiveRes, err error)
	ListArchives(ctx context.Context, req *v1.ListArchivesReq) (res *v1.ListArchivesRes, err error)
}
