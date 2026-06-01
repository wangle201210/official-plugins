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
	// CodeSMSRateLimited reports that one phone/type exceeded the local send cap.
	CodeSMSRateLimited = bizerr.MustDefine("UIDENTITY_SMS_RATE_LIMITED", "SMS verification code is sent too frequently", gcode.CodeInvalidParameter)
	// CodeSMSTypeInvalid reports an unsupported SMS scenario type.
	CodeSMSTypeInvalid = bizerr.MustDefine("UIDENTITY_SMS_TYPE_INVALID", "SMS verification type is invalid", gcode.CodeInvalidParameter)
	// CodeImportInvalid reports an invalid legacy account import workbook.
	CodeImportInvalid = bizerr.MustDefine("UIDENTITY_IMPORT_INVALID", "Account import workbook is invalid", gcode.CodeInvalidParameter)
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
	// CodeInvalidCredentials reports invalid runtime account credentials.
	CodeInvalidCredentials = bizerr.MustDefine("UIDENTITY_INVALID_CREDENTIALS", "Account credentials are invalid", gcode.CodeNotAuthorized)
	// CodePasswordUnlockNumbersRequired reports an empty password unlock number list.
	CodePasswordUnlockNumbersRequired = bizerr.MustDefine("UIDENTITY_PASSWORD_UNLOCK_NUMBERS_REQUIRED", "Select at least one account number to unlock", gcode.CodeInvalidParameter)
	// CodePasswordUnlockNumbersTooMany reports unlock requests above the supported cap.
	CodePasswordUnlockNumbersTooMany = bizerr.MustDefine("UIDENTITY_PASSWORD_UNLOCK_NUMBERS_TOO_MANY", "Password unlock supports at most {limit} account numbers per request", gcode.CodeInvalidParameter)
	// CodePasswordFailuresLocked reports too many password failures for a runtime account.
	CodePasswordFailuresLocked = bizerr.MustDefine("UIDENTITY_PASSWORD_FAILURES_LOCKED", "Password failures exceeded the limit, unlock the account or retry later", gcode.CodeNotAuthorized)
	// CodeApplicationSecretInvalid reports a client secret mismatch.
	CodeApplicationSecretInvalid = bizerr.MustDefine("UIDENTITY_APPLICATION_SECRET_INVALID", "Application secret is invalid", gcode.CodeNotAuthorized)
	// CodeOAuthGrantInvalid reports unsupported or malformed OAuth authorization-code input.
	CodeOAuthGrantInvalid = bizerr.MustDefine("UIDENTITY_OAUTH_GRANT_INVALID", "OAuth authorization grant is invalid", gcode.CodeInvalidParameter)
	// CodeOAuthRedirectInvalid reports a redirect URI that does not match the application callback.
	CodeOAuthRedirectInvalid = bizerr.MustDefine("UIDENTITY_OAUTH_REDIRECT_INVALID", "OAuth redirect URI is invalid", gcode.CodeInvalidParameter)
	// CodeTicketInvalid reports a missing, expired, consumed, or malformed runtime ticket.
	CodeTicketInvalid = bizerr.MustDefine("UIDENTITY_TICKET_INVALID", "Runtime ticket is invalid or expired", gcode.CodeNotAuthorized)
	// CodeActivationInvalid reports a missing, expired, or malformed activation challenge.
	CodeActivationInvalid = bizerr.MustDefine("UIDENTITY_ACTIVATION_INVALID", "Activation challenge is invalid or expired", gcode.CodeInvalidParameter)
	// CodeUnionIDChallengeInvalid reports a missing or expired union ID bind challenge.
	CodeUnionIDChallengeInvalid = bizerr.MustDefine("UIDENTITY_UNION_ID_CHALLENGE_INVALID", "Union ID bind challenge is invalid or expired", gcode.CodeInvalidParameter)
	// CodeWechatLoginInvalid reports a missing or expired Wechat QR login state.
	CodeWechatLoginInvalid = bizerr.MustDefine("UIDENTITY_WECHAT_LOGIN_INVALID", "Wechat QR login state is invalid or expired", gcode.CodeInvalidParameter)
	// CodeContactConflict reports duplicate phone or Wechat binding.
	CodeContactConflict = bizerr.MustDefine("UIDENTITY_CONTACT_CONFLICT", "Contact information is already bound to another account", gcode.CodeInvalidParameter)
	// CodeUnsupportedExternalFlow reports an external dependency that is not configured for this plugin.
	CodeUnsupportedExternalFlow = bizerr.MustDefine("UIDENTITY_EXTERNAL_FLOW_UNSUPPORTED", "External identity flow is not configured", gcode.CodeInvalidParameter)
	// CodeLegacyUploadRequired reports a missing legacy upload file payload.
	CodeLegacyUploadRequired = bizerr.MustDefine("UIDENTITY_LEGACY_UPLOAD_REQUIRED", "Legacy upload file is required", gcode.CodeInvalidParameter)
	// CodeLegacyUploadInvalid reports malformed legacy upload data.
	CodeLegacyUploadInvalid = bizerr.MustDefine("UIDENTITY_LEGACY_UPLOAD_INVALID", "Legacy upload data is invalid", gcode.CodeInvalidParameter)
	// CodeLegacyUploadFailed reports failure while writing plugin-owned upload storage.
	CodeLegacyUploadFailed = bizerr.MustDefine("UIDENTITY_LEGACY_UPLOAD_FAILED", "Legacy upload storage failed", gcode.CodeInternalError)
	// CodeLegacyLogInvalid reports an invalid bounded log snapshot request.
	CodeLegacyLogInvalid = bizerr.MustDefine("UIDENTITY_LEGACY_LOG_INVALID", "Legacy log snapshot request is invalid", gcode.CodeInvalidParameter)
	// CodeLegacyJobDisabled reports a start request for a disabled legacy job.
	CodeLegacyJobDisabled = bizerr.MustDefine("UIDENTITY_LEGACY_JOB_DISABLED", "Legacy job is disabled", gcode.CodeInvalidParameter)
)
