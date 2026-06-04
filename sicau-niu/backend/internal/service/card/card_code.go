// card_code.go defines the card and quote business error codes.

package card

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeCardCategoryInvalid reports that the card category is not an allowed enum value.
	CodeCardCategoryInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_CATEGORY_INVALID",
		"Card category is invalid",
		gcode.CodeInvalidParameter,
	)
	// CodeCardTitleRequired reports that a card title is required.
	CodeCardTitleRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_TITLE_REQUIRED",
		"Card title cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeCardNiuRequired reports that a card must reference an owning cattle.
	CodeCardNiuRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_NIU_REQUIRED",
		"Card owning cattle is required",
		gcode.CodeInvalidParameter,
	)
	// CodeCardNiuInvalid reports that the owning cattle does not exist.
	CodeCardNiuInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_NIU_INVALID",
		"Owning cattle does not exist",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeCardNiuTaken reports that the owning cattle already has a card.
	CodeCardNiuTaken = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_NIU_TAKEN",
		"Owning cattle already has a card",
		gcode.CodeBusinessValidationFailed,
	)
	// CodeCardIDRequired reports that a card ID is required.
	CodeCardIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_ID_REQUIRED",
		"Card ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeCardNotFound reports that the requested card does not exist.
	CodeCardNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_NOT_FOUND",
		"Card does not exist",
		gcode.CodeNotFound,
	)
	// CodeCardQueryFailed reports that a card store query failed.
	CodeCardQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_QUERY_FAILED",
		"Failed to query card data",
		gcode.CodeInternalError,
	)
	// CodeCardWriteFailed reports that a card store write failed.
	CodeCardWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_CARD_WRITE_FAILED",
		"Failed to write card data",
		gcode.CodeInternalError,
	)

	// CodeQuoteContentRequired reports that a quote content is required.
	CodeQuoteContentRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_QUOTE_CONTENT_REQUIRED",
		"Quote content cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeQuoteIDRequired reports that a quote ID is required.
	CodeQuoteIDRequired = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_QUOTE_ID_REQUIRED",
		"Quote ID cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeQuoteNotFound reports that the requested quote does not exist.
	CodeQuoteNotFound = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_QUOTE_NOT_FOUND",
		"Quote does not exist",
		gcode.CodeNotFound,
	)
	// CodeQuoteQueryFailed reports that a quote store query failed.
	CodeQuoteQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_QUOTE_QUERY_FAILED",
		"Failed to query quote data",
		gcode.CodeInternalError,
	)
	// CodeQuoteWriteFailed reports that a quote store write failed.
	CodeQuoteWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_QUOTE_WRITE_FAILED",
		"Failed to write quote data",
		gcode.CodeInternalError,
	)
)
