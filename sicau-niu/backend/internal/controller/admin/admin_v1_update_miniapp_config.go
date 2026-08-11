// admin_v1_update_miniapp_config.go implements operator runtime config writes.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
)

func (c *ControllerV1) UpdateMiniappConfig(ctx context.Context, req *v1.UpdateMiniappConfigReq) (*v1.UpdateMiniappConfigRes, error) {
	campuses := make([]*miniappconfigsvc.Campus, 0, len(req.Campuses))
	for _, item := range req.Campuses {
		if item == nil {
			campuses = append(campuses, nil)
			continue
		}
		campuses = append(campuses, &miniappconfigsvc.Campus{Id: item.Id, Name: item.Name, MapImage: item.MapImage, Calibration: miniappconfigsvc.Calibration{
			Version:   item.Calibration.Version,
			Origin:    miniappconfigsvc.Origin{Lat0: item.Calibration.Origin.Lat0, Lng0: item.Calibration.Origin.Lng0},
			ImageSize: miniappconfigsvc.ImageSize{W: item.Calibration.ImageSize.W, H: item.Calibration.ImageSize.H},
			Affine:    miniappconfigsvc.Affine{A: item.Calibration.Affine.A, B: item.Calibration.Affine.B, C: item.Calibration.Affine.C, D: item.Calibration.Affine.D, Tx: item.Calibration.Affine.Tx, Ty: item.Calibration.Affine.Ty},
		}})
	}
	out, err := c.miniappConfigSvc.Update(ctx, &miniappconfigsvc.Snapshot{
		Campuses: campuses, DefaultCampus: req.DefaultCampus, Anniversary: req.Anniversary,
		AnniversaryAt: req.AnniversaryAt, Debug: req.Debug, AssetsVersion: req.AssetsVersion,
		StaticAssetBaseURL: req.StaticAssetBaseUrl, ActivityPhase: req.ActivityPhase,
	})
	if err != nil {
		return nil, err
	}
	return &v1.UpdateMiniappConfigRes{MiniappConfig: toMiniappConfig(out)}, nil
}
