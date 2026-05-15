package watermark

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func IsWatermarkServer(ctx context.Context) bool {
	return g.Cfg().MustGet(ctx, "tieta.watermark").Bool()
}
