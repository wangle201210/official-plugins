// irontransport_code.go defines stable transport business and storage errors.

package irontransport

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeUnavailable reports that no valid located iron cow can be transported.
	CodeUnavailable = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_UNAVAILABLE", "No located iron-cow is available", gcode.CodeBusinessValidationFailed)
	// CodeTeamNotFound reports an unknown or unavailable transport team.
	CodeTeamNotFound = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_NOT_FOUND", "Transport team does not exist", gcode.CodeNotFound)
	// CodeAlreadyInTeam reports that the player already has active membership.
	CodeAlreadyInTeam = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_ALREADY_IN_TEAM", "Leave the current transport team first", gcode.CodeBusinessValidationFailed)
	// CodeTeamFull reports that a forming team reached its capacity.
	CodeTeamFull = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_FULL", "Transport team is full", gcode.CodeBusinessValidationFailed)
	// CodeTeamNotReady reports that a team has too few members to start.
	CodeTeamNotReady = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_TEAM_NOT_READY", "Transport team does not have enough members", gcode.CodeBusinessValidationFailed)
	// CodeForbidden reports a command attempted by an unauthorized player.
	CodeForbidden = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_FORBIDDEN", "Transport operation is not allowed", gcode.CodeNotAuthorized)
	// CodeInvalidState reports a command incompatible with current team state.
	CodeInvalidState = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_INVALID_STATE", "Transport state does not allow this operation", gcode.CodeBusinessValidationFailed)
	// CodeNotActive reports that no active session accepts the command.
	CodeNotActive = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_NOT_ACTIVE", "No active transport session exists", gcode.CodeBusinessValidationFailed)
	// CodeMovementInvalid reports an invalid or implausible movement sample.
	CodeMovementInvalid = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_MOVEMENT_INVALID", "Transport heartbeat movement is invalid", gcode.CodeBusinessValidationFailed)
	// CodeQueryFailed wraps transport state query failures.
	CodeQueryFailed = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_QUERY_FAILED", "Failed to query transport state", gcode.CodeInternalError)
	// CodeWriteFailed wraps transport mutation failures.
	CodeWriteFailed = bizerr.MustDefine("PLUGIN_SICAU_NIU_TRANSPORT_WRITE_FAILED", "Failed to update transport state", gcode.CodeInternalError)
)
