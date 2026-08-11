// miniappconfig_code.go defines stable business errors for runtime config.

package miniappconfig

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeConfigInvalid reports an invalid complete runtime projection.
	CodeConfigInvalid = bizerr.MustDefine("PLUGIN_SICAU_NIU_MINIAPP_CONFIG_INVALID", "Mini-program config is invalid", gcode.CodeInvalidParameter)
	// CodeConfigQueryFailed wraps runtime configuration read failures.
	CodeConfigQueryFailed = bizerr.MustDefine("PLUGIN_SICAU_NIU_MINIAPP_CONFIG_QUERY_FAILED", "Failed to query mini-program config", gcode.CodeInternalError)
	// CodeConfigWriteFailed wraps runtime configuration persistence failures.
	CodeConfigWriteFailed = bizerr.MustDefine("PLUGIN_SICAU_NIU_MINIAPP_CONFIG_WRITE_FAILED", "Failed to write mini-program config", gcode.CodeInternalError)
)
