// =================================================================================
// This is the aggregated interface for the sicau-niu public memorial-wall API.
// It mirrors the GoFrame controller-interface convention so the wall controller
// can be bound on the host router like the player and admin controllers.
// =================================================================================

package wall

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/wall/v1"
)

// IWallV1 is the public memorial-wall controller contract: the first-activator
// wall, the campus-history highlights and the public activity stats.
type IWallV1 interface {
	FirstActivators(ctx context.Context, req *v1.FirstActivatorsReq) (res *v1.FirstActivatorsRes, err error)
	Highlights(ctx context.Context, req *v1.HighlightsReq) (res *v1.HighlightsRes, err error)
	Stats(ctx context.Context, req *v1.StatsReq) (res *v1.StatsRes, err error)
	Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error)
}
