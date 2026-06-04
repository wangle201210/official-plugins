// certrender_code.go defines the certificate-renderer business error code.

package certrender

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

// CodeCertRenderFailed reports that electronic-certificate PNG encoding failed.
var CodeCertRenderFailed = bizerr.MustDefine(
	"PLUGIN_SICAU_NIU_CERT_RENDER_FAILED",
	"Failed to render electronic certificate image",
	gcode.CodeInternalError,
)
