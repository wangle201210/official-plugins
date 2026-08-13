// player_v1_config.go maps the bounded public runtime snapshot to the player
// API without exposing credentials or operator-only settings.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) Config(ctx context.Context, req *v1.ConfigReq) (res *v1.ConfigRes, err error) {
	snapshot, err := c.miniappConfigSvc.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	campuses := make([]*v1.CampusConfig, 0, len(snapshot.Campuses))
	for _, campus := range snapshot.Campuses {
		cal := campus.Calibration
		campuses = append(campuses, &v1.CampusConfig{
			Id: campus.Id, Name: campus.Name, MapImage: campus.MapImage,
			Calibration: v1.CampusCalibration{
				Version:   cal.Version,
				Origin:    v1.CalibrationOrigin{Lat0: cal.Origin.Lat0, Lng0: cal.Origin.Lng0},
				ImageSize: v1.CalibrationSize{W: cal.ImageSize.W, H: cal.ImageSize.H},
				Affine:    v1.CalibrationAffine{A: cal.Affine.A, B: cal.Affine.B, C: cal.Affine.C, D: cal.Affine.D, Tx: cal.Affine.Tx, Ty: cal.Affine.Ty},
			},
		})
	}
	return &v1.ConfigRes{
		Campuses: campuses, DefaultCampus: snapshot.DefaultCampus,
		ActivateRadiusM: snapshot.ActivateRadiusM, CountdownDays: snapshot.CountdownDays,
		StealDailyLimit: snapshot.StealDailyLimit, GiftDailyLimit: snapshot.GiftDailyLimit, GiftMinAmount: snapshot.GiftMinAmount,
		Anniversary: snapshot.Anniversary, Debug: snapshot.Debug,
		AssetsVersion: snapshot.AssetsVersion, StaticAssetBaseUrl: snapshot.StaticAssetBaseURL,
		ActivityPhase: snapshot.ActivityPhase,
	}, nil
}
