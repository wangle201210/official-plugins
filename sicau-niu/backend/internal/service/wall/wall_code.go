// wall_code.go defines the public memorial-wall business error codes.

package wall

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeWallQueryFailed reports that a public memorial-wall query failed.
	CodeWallQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_WALL_QUERY_FAILED",
		"Failed to query memorial wall data",
		gcode.CodeInternalError,
	)
)
