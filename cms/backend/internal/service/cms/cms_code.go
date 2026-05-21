// This file defines CMS plugin business error codes and runtime i18n metadata.

package cms

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeSiteNotFound reports that CMS site settings do not exist.
	CodeSiteNotFound = bizerr.MustDefine(
		"CMS_SITE_NOT_FOUND",
		"CMS site settings do not exist",
		gcode.CodeNotFound,
	)
	// CodeCategoryNotFound reports that a CMS category does not exist.
	CodeCategoryNotFound = bizerr.MustDefine(
		"CMS_CATEGORY_NOT_FOUND",
		"CMS category does not exist",
		gcode.CodeNotFound,
	)
	// CodeCategoryHasChildren reports that a CMS category cannot be deleted while it has child categories.
	CodeCategoryHasChildren = bizerr.MustDefine(
		"CMS_CATEGORY_HAS_CHILDREN",
		"CMS category has child categories",
		gcode.CodeInvalidOperation,
	)
	// CodeCategoryHasArticles reports that a CMS category cannot be deleted while it owns articles.
	CodeCategoryHasArticles = bizerr.MustDefine(
		"CMS_CATEGORY_HAS_ARTICLES",
		"CMS category has articles",
		gcode.CodeInvalidOperation,
	)
	// CodeCategoryCodeExists reports that a CMS category code is already used.
	CodeCategoryCodeExists = bizerr.MustDefine(
		"CMS_CATEGORY_CODE_EXISTS",
		"CMS category code already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeCategoryParentInvalid reports that a CMS category parent would create a cycle.
	CodeCategoryParentInvalid = bizerr.MustDefine(
		"CMS_CATEGORY_PARENT_INVALID",
		"CMS category parent cannot be itself or its descendant",
		gcode.CodeInvalidParameter,
	)
	// CodeArticleNotFound reports that a CMS article does not exist or is not publicly visible.
	CodeArticleNotFound = bizerr.MustDefine(
		"CMS_ARTICLE_NOT_FOUND",
		"CMS article does not exist",
		gcode.CodeNotFound,
	)
	// CodeArticleSlugExists reports that a CMS article slug is already used.
	CodeArticleSlugExists = bizerr.MustDefine(
		"CMS_ARTICLE_SLUG_EXISTS",
		"CMS article slug already exists",
		gcode.CodeInvalidParameter,
	)
	// CodeMessageNotFound reports that a CMS visitor message does not exist.
	CodeMessageNotFound = bizerr.MustDefine(
		"CMS_MESSAGE_NOT_FOUND",
		"CMS visitor message does not exist",
		gcode.CodeNotFound,
	)
	// CodeLinkNotFound reports that a CMS friendly link does not exist.
	CodeLinkNotFound = bizerr.MustDefine(
		"CMS_LINK_NOT_FOUND",
		"CMS friendly link does not exist",
		gcode.CodeNotFound,
	)
	// CodeSlideNotFound reports that a CMS slide does not exist.
	CodeSlideNotFound = bizerr.MustDefine(
		"CMS_SLIDE_NOT_FOUND",
		"CMS slide does not exist",
		gcode.CodeNotFound,
	)
	// CodePublicContentNotFound reports that requested public CMS content is unavailable.
	CodePublicContentNotFound = bizerr.MustDefine(
		"CMS_PUBLIC_CONTENT_NOT_FOUND",
		"Public CMS content does not exist",
		gcode.CodeNotFound,
	)
	// CodeSampleDataLoadFailed reports that packaged starter content could not be loaded.
	CodeSampleDataLoadFailed = bizerr.MustDefine(
		"CMS_SAMPLE_DATA_LOAD_FAILED",
		"CMS sample data could not be loaded",
		gcode.CodeInternalError,
	)
)
