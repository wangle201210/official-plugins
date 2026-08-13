// feeding_code.go defines the feeding business error codes used across the
// feeding capability.

package feeding

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeNiuIDRequired reports that a target cattle ID is required.
	CodeNiuIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_NIU_ID_REQUIRED",
		"Cattle ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeAmountInvalid reports that the feeding base amount is not positive.
	CodeAmountInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_AMOUNT_INVALID",
		"Feeding amount must be positive",
		gcode.CodeInvalidParameter,
	)
	// CodeRequestIDRequired reports that the mandatory idempotency key is missing
	// or exceeds the public 64-character contract.
	CodeRequestIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_REQUEST_ID_REQUIRED",
		"Request ID is required and must not exceed 64 characters",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuNotFound reports that the target cattle does not exist.
	CodeNiuNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_NIU_NOT_FOUND",
		"Cattle does not exist",
		gcode.CodeNotFound,
	)
	// CodeNiuNotActive reports that the target cattle is not activated and cannot
	// be fed.
	CodeNiuNotActive = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_NIU_NOT_ACTIVE",
		"Cattle is not activated yet",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeQueryFailed reports that a feeding-related store query failed.
	CodeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_QUERY_FAILED",
		"Failed to query feeding data",
		gcode.CodeInternalError,
	)
	// CodeWriteFailed reports that a feeding store write failed.
	CodeWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_FEEDING_WRITE_FAILED",
		"Failed to write feeding data",
		gcode.CodeInternalError,
	)
)
