// Package ironlocation is the feeding capability's iron-cow real-time location
// seam. It returns the current GPS positions of registered iron cows used to
// decide the feeding proximity bonus. The player request path reads only the
// latest stored last_lat/last_lng from the plugin iron table; the external IOT
// platform is refreshed by a separate cron job so feeding does not block on
// network calls. Iron cows without a recorded position are skipped by the caller.
// All store access uses the generated DAO so GoFrame manages soft-delete and
// timestamp columns.
package ironlocation

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// IronPosition is one iron cow's current position projection.
type IronPosition struct {
	// IronID is the iron-cow registration ID.
	IronID int64
	// Lat is the iron cow's current GPS latitude.
	Lat float64
	// Lng is the iron cow's current GPS longitude.
	Lng float64
}

// Gateway is the iron-cow real-time location contract consumed by the feeding
// bonus check. Implementations return every iron cow that currently has a usable
// position; iron cows without a position are omitted.
type Gateway interface {
	// Positions returns the current positions of all located iron cows in one
	// bounded query. It returns a query bizerr on store failure.
	Positions(ctx context.Context) (out []*IronPosition, err error)
}

// Interface compliance assertion for the stored location gateway.
var _ Gateway = (*storedGateway)(nil)

// storedGateway returns the iron cows' stored last_lat/last_lng. It is used by
// the feeding request path so external IOT refresh frequency stays decoupled
// from player traffic.
type storedGateway struct{}

// NewStored creates the request-path iron-cow location gateway backed by the
// stored last_lat/last_lng columns.
func NewStored() Gateway {
	return &storedGateway{}
}

// Positions returns the stored positions of iron cows that have a non-zero
// recorded location. The whole iron set is small and read in one query, so the
// proximity check never issues a per-iron query.
func (g *storedGateway) Positions(ctx context.Context) ([]*IronPosition, error) {
	rows := make([]*entitymodel.Iron, 0)
	err := dao.Iron.Ctx(ctx).
		Fields(
			dao.Iron.Columns().Id,
			dao.Iron.Columns().LastLat,
			dao.Iron.Columns().LastLng,
		).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	positions := make([]*IronPosition, 0, len(rows))
	for _, row := range rows {
		if row.LastLat == 0 && row.LastLng == 0 {
			continue
		}
		positions = append(positions, &IronPosition{
			IronID: row.Id,
			Lat:    row.LastLat,
			Lng:    row.LastLng,
		})
	}
	return positions, nil
}
