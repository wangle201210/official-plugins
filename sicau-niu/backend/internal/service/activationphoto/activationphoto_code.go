// activationphoto_code.go defines stable activation-photo business and storage
// errors returned by player and operator endpoints.

package activationphoto

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodePhotoRequired reports a missing player, request ID or image body.
	CodePhotoRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_REQUIRED",
		"Activation photo is required",
		gcode.CodeInvalidParameter,
	)
	// CodePhotoInvalid reports an unsupported, oversized or undecodable image.
	CodePhotoInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_INVALID",
		"Activation photo must be a valid JPEG, PNG or HEIC image no larger than 5 MiB",
		gcode.CodeInvalidParameter,
	)
	// CodePhotoTooLarge reports a source image whose header declares more pixels
	// than the transcoder accepts, rejected before any decode allocates memory.
	CodePhotoTooLarge = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_TOO_LARGE",
		"Activation photo resolution is too large",
		gcode.CodeInvalidParameter,
	)
	// CodePhotoBusy reports that the bounded transcoding slots are all taken.
	CodePhotoBusy = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_BUSY",
		"Activation photo processing is busy, please retry shortly",
		gcode.CodeBusinessValidationFailed,
	)
	// CodePhotoDailyLimit reports an exhausted per-player daily upload quota.
	CodePhotoDailyLimit = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_DAILY_LIMIT",
		"Daily activation photo upload limit reached",
		gcode.CodeBusinessValidationFailed,
	)
	// CodePhotoNotFound reports an unknown or inaccessible private photo token.
	CodePhotoNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_NOT_FOUND",
		"Activation photo does not exist",
		gcode.CodeNotFound,
	)
	// CodePhotoAlreadyUsed reports a token consumed by an earlier activation.
	CodePhotoAlreadyUsed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_ALREADY_USED",
		"Activation photo has already been used",
		gcode.CodeBusinessValidationFailed,
	)
	// CodePhotoQueryFailed wraps photo metadata or object read failures.
	CodePhotoQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_QUERY_FAILED",
		"Failed to query activation photo",
		gcode.CodeInternalError,
	)
	// CodePhotoWriteFailed wraps photo metadata or object write failures.
	CodePhotoWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_PHOTO_WRITE_FAILED",
		"Failed to store activation photo",
		gcode.CodeInternalError,
	)
)
