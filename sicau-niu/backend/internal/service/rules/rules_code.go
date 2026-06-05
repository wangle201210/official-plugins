// rules_code.go defines the runtime-rule configuration business error codes.

package rules

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeRuleInvalid reports invalid operator-maintained runtime rule values.
	CodeRuleInvalid = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_RULE_INVALID",
		"Invalid sicau-niu runtime rule configuration",
		gcode.CodeInvalidParameter,
	)
	// CodeRuleQueryFailed reports a rule-config read failure.
	CodeRuleQueryFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_RULE_QUERY_FAILED",
		"Failed to query sicau-niu runtime rule configuration",
		gcode.CodeInternalError,
	)
	// CodeRuleWriteFailed reports a rule-config write failure.
	CodeRuleWriteFailed = bizerr.MustDefine(
		"PLUGIN_SICAU_NIU_RULE_WRITE_FAILED",
		"Failed to write sicau-niu runtime rule configuration",
		gcode.CodeInternalError,
	)
)
