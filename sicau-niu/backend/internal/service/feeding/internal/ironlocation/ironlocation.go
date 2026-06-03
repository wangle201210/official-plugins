// Package ironlocation is the feeding capability's iron-cow real-time location
// seam. It returns the current GPS positions of registered iron cows used to
// decide the feeding proximity bonus. The seam isolates the (currently mocked)
// external positioning integration behind a narrow contract so the real external
// API can replace the implementation without touching the feeding bonus logic.
// The mock implementation returns the iron cows' stored last_lat/last_lng from
// the plugin iron table; iron cows without a recorded position are skipped by the
// caller. All store access uses the generated DAO so GoFrame manages soft-delete
// and timestamp columns.
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

// Interface compliance assertion for the mock location gateway.
var _ Gateway = (*mockGateway)(nil)

// mockGateway is the development location gateway. It returns the iron cows'
// stored last_lat/last_lng instead of calling an external positioning API.
type mockGateway struct{}

// NewMock creates the mock iron-cow location gateway backed by the stored
// last_lat/last_lng columns. The real external-API implementation replaces this
// behind the Gateway seam without changing the feeding bonus logic.
func NewMock() Gateway {
	return &mockGateway{}
}

// Positions returns the stored positions of iron cows that have a non-zero
// recorded location. The whole iron set is small and read in one query, so the
// proximity check never issues a per-iron query.
func (g *mockGateway) Positions(ctx context.Context) ([]*IronPosition, error) {
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
