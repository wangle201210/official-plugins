// honor_code.go defines the honor business error codes and the operator-listing
// pagination defaults.

package honor

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// Listing paging defaults bound the operator honor-definition page size.
const (
	defaultPageNum  = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

var (
	// CodeHonorCodeRequired reports that a honor code is required.
	CodeHonorCodeRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_CODE_REQUIRED",
		"Honor code cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorNameRequired reports that a honor name is required.
	CodeHonorNameRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_NAME_REQUIRED",
		"Honor name cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorCodeExists reports that a honor with the same code already exists.
	CodeHonorCodeExists = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_CODE_EXISTS",
		"Honor code already exists",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeHonorTypeInvalid reports that the honor type is not an allowed enum value.
	CodeHonorTypeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_TYPE_INVALID",
		"Honor type is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorUnlockTypeInvalid reports that the unlock rule is not an allowed enum value.
	CodeHonorUnlockTypeInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_UNLOCK_TYPE_INVALID",
		"Honor unlock rule is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorThresholdInvalid reports that a count-based unlock rule lacks a positive threshold.
	CodeHonorThresholdInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_THRESHOLD_INVALID",
		"Count-based unlock rule requires a positive threshold",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorCategoryInvalid reports that the category-complete rule has an invalid or missing card category.
	CodeHonorCategoryInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_CATEGORY_INVALID",
		"Category-complete unlock rule requires a valid card category",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorIDRequired reports that a honor ID is required.
	CodeHonorIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_ID_REQUIRED",
		"Honor ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeHonorNotFound reports that the requested honor does not exist.
	CodeHonorNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_NOT_FOUND",
		"Honor does not exist",
		gcode.CodeNotFound,
	)
	// CodeHonorQueryFailed reports that a honor store query failed.
	CodeHonorQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_QUERY_FAILED",
		"Failed to query honor data",
		gcode.CodeInternalError,
	)
	// CodeHonorWriteFailed reports that a honor store write failed.
	CodeHonorWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_WRITE_FAILED",
		"Failed to write honor data",
		gcode.CodeInternalError,
	)
	// CodeHonorNotCertificate reports that the target honor is not a certificate
	// and therefore has no certificate to render.
	CodeHonorNotCertificate = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_HONOR_NOT_CERTIFICATE",
		"Target honor is not a certificate",
		gcode.CodeInvalidParameter,
	)
	// CodeCertificateNotOwned reports that the requesting player has not been
	// granted the certificate.
	CodeCertificateNotOwned = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CERTIFICATE_NOT_OWNED",
		"Player has not been granted this certificate",
		gcode.CodeInvalidParameter,
	)
)

// normalizePagination applies the paging defaults and the max page-size cap to a
// requested page number and size.
func normalizePagination(pageNum, pageSize int) (int, int) {
	if pageNum <= 0 {
		pageNum = defaultPageNum
	}
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return pageNum, pageSize
}
