// This file defines UIdentity legacy cron business error codes.

package cron

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeJobInvalid reports a malformed or unsupported legacy job definition.
	CodeJobInvalid = bizerr.MustDefine("UIDENTITY_LEGACY_JOB_INVALID", "Legacy job definition is invalid", gcode.CodeInvalidParameter)
	// CodeJobScheduleFailed reports a GoFrame cron registration failure.
	CodeJobScheduleFailed = bizerr.MustDefine("UIDENTITY_LEGACY_JOB_SCHEDULE_FAILED", "Legacy job schedule failed", gcode.CodeInvalidParameter)
	// CodeJobExecutorUnsupported reports an exec target with no plugin-local executor.
	CodeJobExecutorUnsupported = bizerr.MustDefine("UIDENTITY_LEGACY_JOB_EXECUTOR_UNSUPPORTED", "Legacy job executor is not configured", gcode.CodeInvalidParameter)
)
