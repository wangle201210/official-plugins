// player_transport_projection.go maps cloud-moving service models to player DTOs.
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
		Explanation: state.Explanation, MaxEffectiveTeams: state.MaxEffectiveTeams,
		DailyReportLimit: state.DailyReportLimit, TodayReportCount: state.TodayReportCount,
		TodayReportRemaining: state.TodayReportRemaining,
		Teams:                make([]*v1.IronTransportTeam, 0, len(state.Teams)),
	}
	for _, team := range state.Teams {
		out.Teams = append(out.Teams, toIronTransportTeam(team))
	}
	out.MyTeam = toIronTransportTeam(state.MyTeam)
	return out
}

func toIronTransportTeam(team *irontransportsvc.Team) *v1.IronTransportTeam {
	if team == nil {
		return nil
	}
	return &v1.IronTransportTeam{
		Id: strconv.FormatInt(team.ID, 10), Name: team.Name,
		CreatorId: strconv.FormatInt(team.CreatorID, 10), CreatorName: team.CreatorName,
		MemberCount: team.MemberCount, TotalContributionMeters: team.TotalContributionMeters,
		MyContributionMeters: team.MyContributionMeters, HasReportBaseline: team.HasReportBaseline,
		LastActiveAt: apitime.Milli(team.LastActiveAt), CreatedAt: apitime.Milli(team.CreatedAt), Mine: team.Mine,
	}
}

func toIronTransportMember(member *irontransportsvc.Member) *v1.IronTransportMember {
	return &v1.IronTransportMember{
		Id: strconv.FormatInt(member.ID, 10), UserId: strconv.FormatInt(member.UserID, 10),
		Name: member.Name, Avatar: member.Avatar, Role: member.Role,
		ContributionMeters: member.ContributionMeters, JoinedAt: apitime.Milli(member.JoinedAt),
	}
}

func toIronTransportReportResult(report *irontransportsvc.ReportResult) *v1.IronTransportReportResult {
	if report == nil {
		return nil
	}
	return &v1.IronTransportReportResult{
		Id: strconv.FormatInt(report.ID, 10), TeamId: strconv.FormatInt(report.TeamID, 10),
		StartLat: report.StartLat, StartLng: report.StartLng, EndLat: report.EndLat, EndLng: report.EndLng,
		ContributionMeters: report.ContributionMeters, UserTotalMeters: report.UserTotalMeters,
		TeamTotalMeters: report.TeamTotalMeters, TodayReportCount: report.TodayReportCount,
		TodayReportRemaining: report.TodayReportRemaining, AcceptedAt: apitime.Milli(report.AcceptedAt),
	}
}

func toIronTransportReport(report *irontransportsvc.Report) *v1.IronTransportReport {
	return &v1.IronTransportReport{
		IronTransportReportResult: toIronTransportReportResult(report.ReportResult),
		ActivityDate:              report.ActivityDate, SampledAt: apitime.Milli(report.SampledAt),
	}
}
