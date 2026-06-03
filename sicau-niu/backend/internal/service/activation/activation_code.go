// activation_code.go defines the C3 activation, collection and poster business
// error codes used across the activation capability.

package activation

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeNiuIDRequired reports that a target cattle ID is required.
	CodeNiuIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_NIU_ID_REQUIRED",
		"Cattle ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNiuNotFound reports that the target cattle does not exist.
	CodeNiuNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_NIU_NOT_FOUND",
		"Cattle does not exist",
		gcode.CodeNotFound,
	)
	// CodeNiuNotVisible reports that the target cattle is not currently visible and
	// cannot be activated yet.
	CodeNiuNotVisible = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_NIU_NOT_VISIBLE",
		"Cattle is not currently visible",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeOutOfRange reports that the reported location is beyond the LBS threshold.
	CodeOutOfRange = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_OUT_OF_RANGE",
		"Reported location is too far from the cattle anchor",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeDailyLimitReached reports that the player already activated a cattle today.
	CodeDailyLimitReached = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_DAILY_LIMIT",
		"Already activated a cattle today",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeAlreadyActivated reports that the player already activated this cattle.
	CodeAlreadyActivated = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_ALREADY_ACTIVATED",
		"Cattle already activated by this player",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeActivationNotFound reports that the player has not activated the cattle, so
	// no poster can be produced.
	CodeActivationNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_NOT_FOUND",
		"Player has not activated this cattle",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeCategoryInvalid reports that the collection category filter is not an
	// allowed card category.
	CodeCategoryInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_CATEGORY_INVALID",
		"Card category filter is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeQueryFailed reports that an activation-related store query failed.
	CodeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_QUERY_FAILED",
		"Failed to query activation data",
		gcode.CodeInternalError,
	)
	// CodeWriteFailed reports that an activation store write failed.
	CodeWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_ACTIVATION_WRITE_FAILED",
		"Failed to write activation data",
		gcode.CodeInternalError,
	)
)
