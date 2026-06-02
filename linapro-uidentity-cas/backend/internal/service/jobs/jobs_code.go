// This file defines UIdentity managed scheduled-job business error codes.

package jobs

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeJobInvalid reports malformed managed job input or impossible local state.
	CodeJobInvalid = bizerr.MustDefine("UIDENTITY_JOB_INVALID", "UIdentity scheduled job input is invalid", gcode.CodeInvalidParameter)
	// CodeJobExecutorUnsupported reports a required external Oracle or LDAP configuration is missing.
	CodeJobExecutorUnsupported = bizerr.MustDefine("UIDENTITY_JOB_EXECUTOR_UNSUPPORTED", "UIdentity scheduled job executor is not configured", gcode.CodeInvalidParameter)
)
