// admin_v1_get_miniapp_config.go implements operator runtime config reads.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
)

func (c *ControllerV1) GetMiniappConfig(ctx context.Context, req *v1.GetMiniappConfigReq) (*v1.GetMiniappConfigRes, error) {
	out, err := c.miniappConfigSvc.Snapshot(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetMiniappConfigRes{MiniappConfig: toMiniappConfig(out)}, nil
}

func toMiniappConfig(in *miniappconfigsvc.Snapshot) *v1.MiniappConfig {
	campuses := make([]*v1.MiniappCampus, 0, len(in.Campuses))
	for _, item := range in.Campuses {
		cal := item.Calibration
		campuses = append(campuses, &v1.MiniappCampus{Id: item.Id, Name: item.Name, MapImage: item.MapImage, Calibration: v1.MiniappCalibration{
			Version: cal.Version, Origin: v1.MiniappOrigin{Lat0: cal.Origin.Lat0, Lng0: cal.Origin.Lng0},
			ImageSize: v1.MiniappImageSize{W: cal.ImageSize.W, H: cal.ImageSize.H},
			Affine:    v1.MiniappAffineTransform{A: cal.Affine.A, B: cal.Affine.B, C: cal.Affine.C, D: cal.Affine.D, Tx: cal.Affine.Tx, Ty: cal.Affine.Ty},
		}})
	}
	return &v1.MiniappConfig{
		Campuses: campuses, DefaultCampus: in.DefaultCampus, ActivateRadiusM: in.ActivateRadiusM,
		StealDailyLimit: in.StealDailyLimit, GiftDailyLimit: in.GiftDailyLimit, GiftMinAmount: in.GiftMinAmount,
		CountdownDays: in.CountdownDays, Anniversary: in.Anniversary, AnniversaryAt: in.AnniversaryAt,
		Debug: in.Debug, AssetsVersion: in.AssetsVersion, StaticAssetBaseUrl: in.StaticAssetBaseURL, ActivityPhase: in.ActivityPhase,
	}
}
