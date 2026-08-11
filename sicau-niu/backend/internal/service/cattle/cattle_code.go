// cattle_code.go defines the cattle and iron-cow business error codes.

package cattle

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeNiuCodeRequired reports that a cattle code is required.
	CodeNiuCodeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_CODE_REQUIRED",
		"Cattle code cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuCodeExists reports that a cattle with the same code already exists.
	CodeNiuCodeExists = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_CODE_EXISTS",
		"Cattle code already exists",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeNiuTypeInvalid reports that the cattle type is not an allowed enum value.
	CodeNiuTypeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_TYPE_INVALID",
		"Cattle type is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuSubtypeInvalid reports that the special subtype is invalid for the type.
	CodeNiuSubtypeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_SUBTYPE_INVALID",
		"Cattle special subtype is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuNameRequired reports that a special cattle name is required.
	CodeNiuNameRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_NAME_REQUIRED",
		"Special cattle name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuCollegeRequired reports that a college cattle must link a college.
	CodeNiuCollegeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_COLLEGE_REQUIRED",
		"College cattle must link an existing college",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuCollegeInvalid reports that the linked college does not exist.
	CodeNiuCollegeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_COLLEGE_INVALID",
		"Linked college does not exist",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeNiuIDRequired reports that a cattle ID is required.
	CodeNiuIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_ID_REQUIRED",
		"Cattle ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuNotFound reports that the requested cattle does not exist.
	CodeNiuNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_NOT_FOUND",
		"Cattle does not exist",
		gcode.CodeNotFound,
	)
	// CodeNiuQueryFailed reports that a cattle store query failed.
	CodeNiuQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_QUERY_FAILED",
		"Failed to query cattle data",
		gcode.CodeInternalError,
	)
	// CodeNiuWriteFailed reports that a cattle store write failed.
	CodeNiuWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_WRITE_FAILED",
		"Failed to write cattle data",
		gcode.CodeInternalError,
	)
	CodeNiuImportInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_NIU_IMPORT_INVALID",
		"Cattle import must contain 1 to 200 valid unique items",
		gcode.CodeInvalidParameter,
	)

	// CodeIronCodeRequired reports that an iron-cow code is required.
	CodeIronCodeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_CODE_REQUIRED",
		"Iron-cow code cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeIronNameRequired reports that an iron-cow name is required.
	CodeIronNameRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_NAME_REQUIRED",
		"Iron-cow name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeIronCodeExists reports that an iron-cow with the same code already exists.
	CodeIronCodeExists = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_CODE_EXISTS",
		"Iron-cow code already exists",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeIronIDRequired reports that an iron-cow ID is required.
	CodeIronIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_ID_REQUIRED",
		"Iron-cow ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeIronNotFound reports that the requested iron-cow does not exist.
	CodeIronNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_NOT_FOUND",
		"Iron-cow does not exist",
		gcode.CodeNotFound,
	)
	// CodeIronQueryFailed reports that an iron-cow store query failed.
	CodeIronQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_QUERY_FAILED",
		"Failed to query iron-cow data",
		gcode.CodeInternalError,
	)
	// CodeIronWriteFailed reports that an iron-cow store write failed.
	CodeIronWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IRON_WRITE_FAILED",
		"Failed to write iron-cow data",
		gcode.CodeInternalError,
	)
)
