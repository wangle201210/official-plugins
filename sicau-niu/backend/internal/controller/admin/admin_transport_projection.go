// admin_transport_projection.go maps cloud-moving service models to operator DTOs.
package admin

import (
	"strconv"

	"lina-core/pkg/apitime"
	"lina-plugin-sicau-niu/backend/api/admin/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func toAdminTransportTeam(team *irontransportsvc.AdminTeam) *v1.AdminTransportTeam {
	return &v1.AdminTransportTeam{
		Id: strconv.FormatInt(team.ID, 10), Name: team.Name, Status: team.Status,
		CreatorId: strconv.FormatInt(team.CreatorID, 10), CreatorName: team.CreatorName,
		MemberCount: team.MemberCount, TotalContributionMeters: team.TotalContributionMeters,
		LastActiveAt: apitime.Milli(team.LastActiveAt), InvalidatedAt: apitime.Milli(team.InvalidatedAt),
		InvalidReason: team.InvalidReason, CreatedAt: apitime.Milli(team.CreatedAt),
	}
}

func toAdminTransportReport(report *irontransportsvc.AdminReport) *v1.AdminTransportReport {
	return &v1.AdminTransportReport{
		Id: strconv.FormatInt(report.ID, 10), TeamId: strconv.FormatInt(report.TeamID, 10),
		TeamName: report.TeamName, MemberId: strconv.FormatInt(report.MemberID, 10),
		UserId: strconv.FormatInt(report.UserID, 10), UserName: report.UserName,
		ActivityDate: report.ActivityDate, StartLat: report.StartLat, StartLng: report.StartLng,
		EndLat: report.EndLat, EndLng: report.EndLng, SampledAt: apitime.Milli(report.SampledAt),
		AcceptedAt: apitime.Milli(report.AcceptedAt), ContributionMeters: report.ContributionMeters,
		UserTotalMeters: report.UserTotalMeters, TeamTotalMeters: report.TeamTotalMeters,
	}
}
