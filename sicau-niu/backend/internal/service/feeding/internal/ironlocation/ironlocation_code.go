// ironlocation_code.go defines the iron-cow location seam business error codes.

package ironlocation

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeQueryFailed reports that reading iron-cow positions failed.
	CodeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_LOCATION_QUERY_FAILED",
		"Failed to query iron-cow locations",
		gcode.CodeInternalError,
	)
	// CodeIOTRequestFailed reports that the external IOT positioning platform request failed.
	CodeIOTRequestFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_REQUEST_FAILED",
		"Failed to reach the IOT positioning platform",
		gcode.CodeInternalError,
	)
	// CodeIOTResponseInvalid reports that the external IOT positioning platform returned an unusable response.
	CodeIOTResponseInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_RESPONSE_INVALID",
		"The IOT positioning platform response is invalid",
		gcode.CodeInternalError,
	)
	// CodeIOTNotConfigured reports that locator platform credentials are absent.
	CodeIOTNotConfigured = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_NOT_CONFIGURED",
		"The IOT positioning platform is not configured",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeIOTCommandInvalid reports an invalid device code or reporting cycle.
	CodeIOTCommandInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_COMMAND_INVALID",
		"The IOT locator command is invalid",
		gcode.CodeInvalidParameter,
	)
)
