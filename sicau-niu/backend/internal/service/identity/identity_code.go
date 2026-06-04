// identity_code.go defines the player identity business error codes.

package identity

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeLoginCodeRequired reports that a WeChat login code is required.
	CodeLoginCodeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_LOGIN_CODE_REQUIRED",
		"WeChat login code cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeLoginFailed reports that the player login flow failed.
	CodeLoginFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_LOGIN_FAILED",
		"Player login failed",
		gcode.CodeNotAuthorized,
	)
	// CodePlayerWriteFailed reports that a player store write failed.
	CodePlayerWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PLAYER_WRITE_FAILED",
		"Failed to write player data",
		gcode.CodeInternalError,
	)
	// CodePlayerQueryFailed reports that a player store query failed.
	CodePlayerQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PLAYER_QUERY_FAILED",
		"Failed to query player data",
		gcode.CodeInternalError,
	)
	// CodePlayerNotFound reports that the requested player does not exist.
	CodePlayerNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PLAYER_NOT_FOUND",
		"Player does not exist",
		gcode.CodeNotFound,
	)
	// CodePhoneRequired reports that a phone authorization payload is required.
	CodePhoneRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHONE_REQUIRED",
		"Phone authorization payload cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodePhoneTaken reports that the phone number already belongs to another player.
	CodePhoneTaken = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHONE_TAKEN",
		"Phone number is already bound to another account",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeIdentityTypeInvalid reports that the submitted identity type is not allowed.
	CodeIdentityTypeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IDENTITY_TYPE_INVALID",
		"Identity type is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCollegeRequired reports that a college is required for the identity type.
	CodeCollegeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IDENTITY_COLLEGE_REQUIRED",
		"College is required for the selected identity type",
		gcode.CodeInvalidParameter,
	)
	// CodeCollegeInvalid reports that the selected college does not exist.
	CodeCollegeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IDENTITY_COLLEGE_INVALID",
		"Selected college does not exist",
		gcode.CodeInvalidParameter,
	)
	// CodeGradeInvalid reports that the grade value is not a positive integer.
	CodeGradeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IDENTITY_GRADE_INVALID",
		"Grade must be a positive integer",
		gcode.CodeInvalidParameter,
	)
	// CodeGraduationYearInvalid reports that the graduation year is out of range.
	CodeGraduationYearInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_IDENTITY_GRADUATION_YEAR_INVALID",
		"Graduation year is invalid",
		gcode.CodeInvalidParameter,
	)
)
