// irontransport_code.go defines stable cloud-moving business errors.
package irontransport

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	CodeInvalidInput  = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_INVALID_INPUT", "Cloud-moving request is invalid", gcode.CodeBusinessValidationFailed)
	CodeTeamNotFound  = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_NOT_FOUND", "Cloud-moving team does not exist", gcode.CodeNotFound)
	CodeAlreadyInTeam = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_ALREADY_IN_TEAM", "Player already belongs to a cloud-moving team", gcode.CodeBusinessValidationFailed)
	CodeTeamNameTaken = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_NAME_TAKEN", "An effective cloud-moving team already uses this name", gcode.CodeBusinessValidationFailed)
	CodeTeamLimit     = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_LIMIT", "The maximum number of effective cloud-moving teams has been reached", gcode.CodeBusinessValidationFailed)
	CodeDailyLimit    = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_DAILY_LIMIT", "The daily position report limit has been reached", gcode.CodeBusinessValidationFailed)
	CodeQueryFailed   = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_QUERY_FAILED", "Failed to query cloud-moving data", gcode.CodeInternalError)
	CodeWriteFailed   = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_WRITE_FAILED", "Failed to update cloud-moving data", gcode.CodeInternalError)
)

// InvalidInputError lets controllers reject structurally missing pointer fields
// without duplicating the package's stable business error contract.
func InvalidInputError() error {
	return bizerr.NewCode(CodeInvalidInput)
}
