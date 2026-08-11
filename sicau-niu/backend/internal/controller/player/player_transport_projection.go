package player

import (
	"strconv"

	"lina-core/pkg/apitime"
	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func toIronTransportState(state *irontransportsvc.State) *v1.IronTransportState {
	if state == nil {
		return nil
	}
	out := &v1.IronTransportState{
		Enabled: state.Enabled, Explanation: state.Explanation, IdleTimeoutSec: state.IdleTimeoutSec,
		MinTeamSize: state.MinTeamSize, CampusId: state.CampusID, Teams: make([]*v1.IronTransportTeam, 0, len(state.Teams)),
	}
	if state.MyTeamID > 0 {
		out.MyTeamId = strconv.FormatInt(state.MyTeamID, 10)
	}
	if state.IronCow != nil {
		out.IronCow = &v1.IronTransportCow{Id: strconv.FormatInt(state.IronCow.ID, 10), Name: state.IronCow.Name, Lat: state.IronCow.Lat, Lng: state.IronCow.Lng, Status: state.IronCow.Status, MovedMeters: state.IronCow.MovedMeters}
	}
	for _, team := range state.Teams {
		projected := &v1.IronTransportTeam{Id: strconv.FormatInt(team.ID, 10), Name: team.Name, Code: team.Code, CampusId: team.CampusID, LeaderId: strconv.FormatInt(team.LeaderID, 10), MinMembers: team.MinMembers, MaxMembers: team.MaxMembers, Visible: team.Visible, Members: make([]*v1.IronTransportMember, 0, len(team.Members))}
		for _, member := range team.Members {
			projected.Members = append(projected.Members, &v1.IronTransportMember{Id: strconv.FormatInt(member.ID, 10), Name: member.Name, Avatar: member.Avatar, Role: member.Role, JoinedAt: apitime.Milli(member.JoinedAt)})
		}
		out.Teams = append(out.Teams, projected)
	}
	if state.Session != nil {
		out.Session = &v1.IronTransportSession{Id: strconv.FormatInt(state.Session.ID, 10), TeamId: strconv.FormatInt(state.Session.TeamID, 10), Status: state.Session.Status, StartedAt: apitime.Milli(state.Session.StartedAt), LastActiveAt: apitime.Milli(state.Session.LastActiveAt), MovedMeters: state.Session.MovedMeters, Trace: make([]*v1.IronTransportPoint, 0, len(state.Session.Trace))}
		for _, point := range state.Session.Trace {
			out.Session.Trace = append(out.Session.Trace, &v1.IronTransportPoint{Lat: point.Lat, Lng: point.Lng, At: apitime.Milli(point.At)})
		}
	}
	return out
}
