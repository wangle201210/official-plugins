// ironlocation_code.go defines the iron-cow location seam business error codes.

package ironlocation

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// CodeQueryFailed reports that reading iron-cow positions failed.
var CodeQueryFailed = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_IRON_LOCATION_QUERY_FAILED",
	"Failed to query iron-cow locations",
	gcode.CodeInternalError,
)

// CodeIOTRequestFailed reports that the external IOT positioning platform request failed.
var CodeIOTRequestFailed = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_REQUEST_FAILED",
	"Failed to reach the IOT positioning platform",
	gcode.CodeInternalError,
)

// CodeIOTResponseInvalid reports that the external IOT positioning platform returned an unusable response.
var CodeIOTResponseInvalid = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_IRON_LOCATION_IOT_RESPONSE_INVALID",
	"The IOT positioning platform response is invalid",
	gcode.CodeInternalError,
)
