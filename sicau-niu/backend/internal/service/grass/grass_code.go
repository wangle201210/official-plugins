// grass_code.go defines the grass account, ledger and check-in business error
// codes used across the grass capability.

package grass

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeAlreadyCheckedIn reports that the player already checked in today.
	CodeAlreadyCheckedIn = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_ALREADY_CHECKED_IN",
		"Already checked in today",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeInsufficientBalance reports that a debit exceeds the available balance.
	CodeInsufficientBalance = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_INSUFFICIENT_BALANCE",
		"Insufficient grass balance",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeInvalidDelta reports that an accounting delta of zero was requested.
	CodeInvalidDelta = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_INVALID_DELTA",
		"Grass change amount cannot be zero",
		gcode.CodeInvalidParameter,
	)
	// CodeRequestIDRequired reports that the mandatory idempotency key is missing
	// or exceeds the service contract's length limit.
	CodeRequestIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_REQUEST_ID_REQUIRED",
		"Request ID is required and must not exceed 64 characters",
		gcode.CodeInvalidParameter,
	)
	// CodeQueryFailed reports that a grass store query failed.
	CodeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_QUERY_FAILED",
		"Failed to query grass data",
		gcode.CodeInternalError,
	)
	// CodeWriteFailed reports that a grass store write failed.
	CodeWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_GRASS_WRITE_FAILED",
		"Failed to write grass data",
		gcode.CodeInternalError,
	)
)
