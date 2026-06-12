// grasssocial_code.go defines the steal, gift and inbox business error codes
// used across the grass-social capability.

package grasssocial

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeTargetRequired reports that a steal/gift target player ID is required.
	CodeTargetRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_TARGET_REQUIRED",
		"Target player ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeSelfActionForbidden reports that a player targeted themselves.
	CodeSelfActionForbidden = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_SELF_FORBIDDEN",
		"Cannot target yourself",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeTargetNotStealable reports that the target is not in the player's daily
	// stealable list.
	CodeTargetNotStealable = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_TARGET_NOT_STEALABLE",
		"Target is not in today's stealable list",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeDuplicateRequest reports that a steal/gift request ID was already
	// processed, so the retry must not execute the action a second time.
	CodeDuplicateRequest = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_DUPLICATE_REQUEST",
		"Duplicate request already processed",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeStealLimitReached reports that the player reached the daily steal limit.
	CodeStealLimitReached = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_STEAL_LIMIT",
		"Daily steal limit reached",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeTargetNoGrass reports that the steal target has no grass to take.
	CodeTargetNoGrass = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_TARGET_NO_GRASS",
		"Target has no grass to steal",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeGiftAmountTooSmall reports that the gift amount is below the per-gift
	// minimum.
	CodeGiftAmountTooSmall = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_GIFT_TOO_SMALL",
		"Gift amount is below the per-gift minimum",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeGiftLimitReached reports that the player reached the daily gift limit.
	CodeGiftLimitReached = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_GIFT_LIMIT",
		"Daily gift limit reached",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeRecipientNotFound reports that the gift recipient does not exist.
	CodeRecipientNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_RECIPIENT_NOT_FOUND",
		"Recipient player does not exist",
		gcode.CodeNotFound,
	)
	// CodeMessageNotFound reports that the message does not exist or does not
	// belong to the player.
	CodeMessageNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_MESSAGE_NOT_FOUND",
		"Message does not exist",
		gcode.CodeNotFound,
	)
	// CodeQueryFailed reports that a grass-social store query failed.
	CodeQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_QUERY_FAILED",
		"Failed to query grass-social data",
		gcode.CodeInternalError,
	)
	// CodeWriteFailed reports that a grass-social store write failed.
	CodeWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SOCIAL_WRITE_FAILED",
		"Failed to write grass-social data",
		gcode.CodeInternalError,
	)
)
