// record_code.go defines the activity-record query business error code.

package record

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// CodeRecordQueryFailed reports that an activity-record query failed.
var CodeRecordQueryFailed = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_RECORD_QUERY_FAILED",
	"Failed to query activity records",
	gcode.CodeInternalError,
)
