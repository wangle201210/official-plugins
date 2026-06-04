// college_code.go defines the college dictionary business error codes.

package college

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCollegeNameRequired reports that a college name is required.
	CodeCollegeNameRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_NAME_REQUIRED",
		"College name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeCollegeNameExists reports that a college with the same name already exists.
	CodeCollegeNameExists = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_NAME_EXISTS",
		"College name already exists",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeCollegeIDRequired reports that a college ID is required.
	CodeCollegeIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_ID_REQUIRED",
		"College ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeCollegeNotFound reports that the requested college does not exist.
	CodeCollegeNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_NOT_FOUND",
		"College does not exist",
		gcode.CodeNotFound,
	)
	// CodeCollegeReferenced reports that the college is still referenced by players.
	CodeCollegeReferenced = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_REFERENCED",
		"College is selected by players and cannot be deleted",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeCollegeQueryFailed reports that a college store query failed.
	CodeCollegeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_QUERY_FAILED",
		"Failed to query college data",
		gcode.CodeInternalError,
	)
	// CodeCollegeWriteFailed reports that a college store write failed.
	CodeCollegeWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_COLLEGE_WRITE_FAILED",
		"Failed to write college data",
		gcode.CodeInternalError,
	)
)
