// posterrender_code.go defines the poster-renderer business error code.

package posterrender

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// CodePosterRenderFailed reports that activation-poster PNG encoding failed.
var CodePosterRenderFailed = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_POSTER_RENDER_FAILED",
	"Failed to render activation poster image",
	gcode.CodeInternalError,
)
