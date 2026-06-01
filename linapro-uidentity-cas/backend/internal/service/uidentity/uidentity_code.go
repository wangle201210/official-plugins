// This file defines UIdentity CAS business error codes and runtime i18n metadata.

package uidentity

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeResourceNotSupported reports an unknown UIdentity resource name.
	CodeResourceNotSupported = bizerr.MustDefine("UIDENTITY_RESOURCE_NOT_SUPPORTED", "UIdentity resource is not supported", gcode.CodeInvalidParameter)
	// CodeResourceNotFound reports a missing tenant-visible resource record.
	CodeResourceNotFound = bizerr.MustDefine("UIDENTITY_RESOURCE_NOT_FOUND", "UIdentity resource record does not exist", gcode.CodeNotFound)
	// CodeDeleteIDsRequired reports an empty delete ID list.
	CodeDeleteIDsRequired = bizerr.MustDefine("UIDENTITY_DELETE_IDS_REQUIRED", "Select at least one record to delete", gcode.CodeInvalidParameter)
	// CodeDeleteIDsTooMany reports delete ID requests above the supported cap.
	CodeDeleteIDsTooMany = bizerr.MustDefine("UIDENTITY_DELETE_IDS_TOO_MANY", "Delete supports at most {limit} IDs per request", gcode.CodeInvalidParameter)
	// CodePasswordWeak reports a password that does not satisfy active policy.
	CodePasswordWeak = bizerr.MustDefine("UIDENTITY_PASSWORD_WEAK", "Password does not satisfy active policy", gcode.CodeInvalidParameter)
	// CodePasswordChallengeInvalid reports a missing, expired, or invalid password challenge.
	CodePasswordChallengeInvalid = bizerr.MustDefine("UIDENTITY_PASSWORD_CHALLENGE_INVALID", "Password reset challenge is invalid or expired", gcode.CodeInvalidParameter)
	// CodeSMSCodeInvalid reports that phone verification did not match a plugin SMS record.
	CodeSMSCodeInvalid = bizerr.MustDefine("UIDENTITY_SMS_CODE_INVALID", "SMS verification code is invalid", gcode.CodeInvalidParameter)
	// CodeAccountLocked reports that runtime access hit a locked account.
	CodeAccountLocked = bizerr.MustDefine("UIDENTITY_ACCOUNT_LOCKED", "Account is locked", gcode.CodeNotAuthorized)
	// CodeAccountInactive reports that runtime access hit a non-normal account.
	CodeAccountInactive = bizerr.MustDefine("UIDENTITY_ACCOUNT_INACTIVE", "Account is not active", gcode.CodeNotAuthorized)
	// CodeApplicationDisabled reports runtime access to a disabled application.
	CodeApplicationDisabled = bizerr.MustDefine("UIDENTITY_APPLICATION_DISABLED", "Application is disabled", gcode.CodeNotAuthorized)
	// CodeAccessDenied reports blacklist or application access rejection.
	CodeAccessDenied = bizerr.MustDefine("UIDENTITY_ACCESS_DENIED", "Application access is denied", gcode.CodeNotAuthorized)
	// CodeCASValidateURLMissing reports a missing plugin CAS validation URL.
	CodeCASValidateURLMissing = bizerr.MustDefine("UIDENTITY_CAS_VALIDATE_URL_MISSING", "CAS validation URL is not configured", gcode.CodeInvalidParameter)
	// CodeCASValidationFailed reports a failed CAS validation response.
	CodeCASValidationFailed = bizerr.MustDefine("UIDENTITY_CAS_VALIDATION_FAILED", "CAS ticket validation failed", gcode.CodeNotAuthorized)
)
