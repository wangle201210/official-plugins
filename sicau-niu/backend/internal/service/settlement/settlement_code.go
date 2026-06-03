// settlement_code.go defines the operator settlement business error codes.

package settlement

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSettlementQueryFailed reports that a settlement aggregation or read query
	// failed.
	CodeSettlementQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_QUERY_FAILED",
		"Failed to query settlement data",
		gcode.CodeInternalError,
	)
	// CodeSettlementIssueFailed reports that the certificate grant write failed.
	CodeSettlementIssueFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_ISSUE_FAILED",
		"Failed to issue certificates",
		gcode.CodeInternalError,
	)
	// CodeSettlementArchiveFailed reports that the archive snapshot write failed.
	CodeSettlementArchiveFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_ARCHIVE_FAILED",
		"Failed to create settlement archive",
		gcode.CodeInternalError,
	)
	// CodeSettlementHonorNotFound reports that the target honor does not exist.
	CodeSettlementHonorNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_HONOR_NOT_FOUND",
		"Target honor not found",
		gcode.CodeInvalidParameter,
	)
	// CodeSettlementNotCertificate reports that the target honor is not a
	// certificate honor and cannot be batch-issued.
	CodeSettlementNotCertificate = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_NOT_CERTIFICATE",
		"Target honor is not a certificate and cannot be batch-issued",
		gcode.CodeInvalidParameter,
	)
	// CodeSettlementUnlockUnsupported reports that the certificate's unlock rule is
	// a collection rule that is unlocked individually in the mini-program and is not
	// batch-settled.
	CodeSettlementUnlockUnsupported = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_SETTLEMENT_UNLOCK_UNSUPPORTED",
		"Collection-based unlock rules are not batch-settled; players unlock them individually",
		gcode.CodeInvalidParameter,
	)
)
