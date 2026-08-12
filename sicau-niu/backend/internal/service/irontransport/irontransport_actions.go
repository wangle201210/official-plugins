// irontransport_actions.go implements transactional cloud-moving writes.
package irontransport

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/activityday"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// CreateTeam creates one effective team and its creator membership atomically.
func (s *serviceImpl) CreateTeam(ctx context.Context, playerID int64, in *CreateTeamInput) (*State, error) {
	if in == nil || playerID <= 0 {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	name, nameOK := normalizeName(in.Name)
	requestID, requestOK := normalizeRequestID(in.RequestID)
	if !nameOK || !requestOK {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		var replay *entitymodel.TransportTeam
		if err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{LeaderUserId: playerID, CreateRequestId: requestID}).Limit(1).Scan(&replay); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if replay != nil {
			return nil
		}
		member, err := activeMembership(ctx, playerID, true)
		if err != nil {
			return err
		}
		if member != nil {
			return bizerr.NewCode(CodeAlreadyInTeam)
		}
		if err = lockTeamCapacity(ctx); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		nameTaken, err := effectiveTeamNameTaken(ctx, name, 0)
		if err != nil {
			return err
		}
		if nameTaken {
			return bizerr.NewCode(CodeTeamNameTaken)
		}
		count, err := dao.TransportTeam.Ctx(ctx).
			Where(do.TransportTeam{Status: teamStatusEffective, Visible: 1}).Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if count >= maxEffectiveTeams {
			return bizerr.NewCode(CodeTeamLimit)
		}
		now := time.Now()
		teamID, err := dao.TransportTeam.Ctx(ctx).Data(do.TransportTeam{
			Name: name, LeaderUserId: playerID, CreateRequestId: requestID,
			Status: teamStatusEffective, Visible: 1, MemberCount: 1,
			TotalContributionMeters: 0, LastActiveAt: &now,
		}).InsertAndGetId()
		if err != nil {
			return transportWriteError(err)
		}
		if _, err = dao.TransportMember.Ctx(ctx).Data(do.TransportMember{
			TeamId: teamID, UserId: playerID, JoinRequestId: requestID,
			Role: roleCreator, JoinedAt: &now, TotalContributionMeters: 0,
		}).Insert(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// JoinTeam binds the player to one effective team without a distance or size rule.
func (s *serviceImpl) JoinTeam(ctx context.Context, playerID, teamID int64, requestValue string) (*State, error) {
	requestID, requestOK := normalizeRequestID(requestValue)
	if playerID <= 0 || teamID <= 0 || !requestOK {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	teamExpired := false
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		var replay *entitymodel.TransportMember
		if err := dao.TransportMember.Ctx(ctx).Where(do.TransportMember{UserId: playerID, JoinRequestId: requestID}).Limit(1).Scan(&replay); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if replay != nil {
			return nil
		}
		member, err := activeMembership(ctx, playerID, true)
		if err != nil {
			return err
		}
		if member != nil {
			return bizerr.NewCode(CodeAlreadyInTeam)
		}
		var team *entitymodel.TransportTeam
		if err = dao.TransportTeam.Ctx(ctx).
			Where(do.TransportTeam{Id: teamID, Status: teamStatusEffective, Visible: 1}).
			LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		acceptedAt := time.Now()
		teamExpired, err = s.invalidateLockedTeamIfExpired(ctx, team, acceptedAt)
		if err != nil || teamExpired {
			return err
		}
		if _, err = dao.TransportMember.Ctx(ctx).Data(do.TransportMember{
			TeamId: teamID, UserId: playerID, JoinRequestId: requestID,
			Role: roleMember, JoinedAt: &acceptedAt, TotalContributionMeters: 0,
		}).Insert(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{
			MemberCount:  gdb.Raw(dao.TransportTeam.Columns().MemberCount + " + 1"),
			LastActiveAt: &acceptedAt,
		}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if teamExpired {
		return nil, bizerr.NewCode(CodeTeamNotFound)
	}
	return s.State(ctx, playerID)
}

// Report persists one immutable successful position fact and updates totals.
func (s *serviceImpl) Report(ctx context.Context, playerID int64, in *ReportInput) (*ReportResult, error) {
	if in == nil || playerID <= 0 || in.TeamID <= 0 || !validCoordinate(in.Lat, in.Lng) || in.SampledAt.IsZero() {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	requestID, requestOK := normalizeRequestID(in.RequestID)
	if !requestOK {
		return nil, bizerr.NewCode(CodeInvalidInput)
	}
	if _, err := s.ExpireInactive(ctx, time.Now()); err != nil {
		return nil, err
	}
	var (
		result      *ReportResult
		teamExpired bool
	)
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		replay, err := findReportByRequest(ctx, playerID, requestID)
		if err != nil {
			return err
		}
		if replay != nil {
			result = reportResultFromRow(replay)
			return nil
		}
		var team *entitymodel.TransportTeam
		if err = dao.TransportTeam.Ctx(ctx).
			Where(do.TransportTeam{Id: in.TeamID, Status: teamStatusEffective, Visible: 1}).
			LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		acceptedAt := time.Now()
		teamExpired, err = s.invalidateLockedTeamIfExpired(ctx, team, acceptedAt)
		if err != nil || teamExpired {
			return err
		}
		var member *entitymodel.TransportMember
		if err = activeMemberModel(ctx, playerID).
			Where(dao.TransportMember.Columns().TeamId, in.TeamID).
			LockUpdate().Scan(&member); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if member == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		day := activityday.Date(acceptedAt)
		count, err := dao.TransportReport.Ctx(ctx).Where(do.TransportReport{UserId: playerID, ActivityDate: day}).Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if count >= dailyReportLimit {
			return bizerr.NewCode(CodeDailyLimit)
		}
		var startLat, startLng *float64
		contribution := int64(0)
		if member.LastReportAt != nil {
			startLatValue := member.LastReportLat
			startLngValue := member.LastReportLng
			startLat, startLng = &startLatValue, &startLngValue
			contribution = haversineMeters(member.LastReportLat, member.LastReportLng, in.Lat, in.Lng)
		}
		userTotal := member.TotalContributionMeters + contribution
		teamTotal := team.TotalContributionMeters + contribution
		dailyCount := count + 1
		reportID, err := dao.TransportReport.Ctx(ctx).Data(do.TransportReport{
			ActivityKey: defaultActivityKey, TeamId: in.TeamID, MemberId: member.Id,
			UserId: playerID, RequestId: requestID, ActivityDate: day,
			StartLat: startLat, StartLng: startLng, EndLat: in.Lat, EndLng: in.Lng,
			SampledAt: &in.SampledAt, AcceptedAt: &acceptedAt, ContributionMeters: contribution,
			UserTotalMeters: userTotal, TeamTotalMeters: teamTotal, DailyReportCount: dailyCount,
		}).InsertAndGetId()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportMember.Ctx(ctx).Where(do.TransportMember{Id: member.Id}).Data(do.TransportMember{
			TotalContributionMeters: userTotal, LastReportLat: in.Lat,
			LastReportLng: in.Lng, LastReportAt: &acceptedAt,
		}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: team.Id}).Data(do.TransportTeam{
			TotalContributionMeters: teamTotal, LastActiveAt: &acceptedAt,
		}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		result = &ReportResult{
			ID: reportID, TeamID: team.Id, StartLat: startLat, StartLng: startLng,
			EndLat: in.Lat, EndLng: in.Lng, ContributionMeters: contribution,
			UserTotalMeters: userTotal, TeamTotalMeters: teamTotal,
			TodayReportCount: dailyCount, TodayReportRemaining: dailyReportLimit - dailyCount,
			AcceptedAt: &acceptedAt,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if teamExpired {
		return nil, bizerr.NewCode(CodeTeamNotFound)
	}
	return result, nil
}

// RenameTeam changes only the display name; it does not refresh team activity.
func (s *serviceImpl) RenameTeam(ctx context.Context, teamID int64, value string) error {
	name, ok := normalizeName(value)
	if teamID <= 0 || !ok {
		return bizerr.NewCode(CodeInvalidInput)
	}
	return s.transaction(ctx, func(ctx context.Context) error {
		if err := lockTeamCapacity(ctx); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		var team *entitymodel.TransportTeam
		if err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		if team.Status == string(teamStatusEffective) {
			nameTaken, err := effectiveTeamNameTaken(ctx, name, teamID)
			if err != nil {
				return err
			}
			if nameTaken {
				return bizerr.NewCode(CodeTeamNameTaken)
			}
		}
		if _, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{Name: name}).Update(); err != nil {
			return transportWriteError(err)
		}
		return nil
	})
}

// ExpireInactive invalidates every effective team idle for at least the policy
// duration and releases all active memberships in the same transaction.
func (s *serviceImpl) ExpireInactive(ctx context.Context, now time.Time) (int, error) {
	if now.IsZero() {
		now = time.Now()
	}
	cutoff := now.Add(-s.inactiveAfter)
	invalidated := 0
	err := s.transaction(ctx, func(ctx context.Context) error {
		teams := make([]*entitymodel.TransportTeam, 0, maxEffectiveTeams)
		if err := dao.TransportTeam.Ctx(ctx).
			Where(do.TransportTeam{Status: teamStatusEffective, Visible: 1}).
			Where(dao.TransportTeam.Columns().LastActiveAt+" <= ?", cutoff).
			OrderAsc(dao.TransportTeam.Columns().Id).Limit(maxEffectiveTeams).LockUpdate().Scan(&teams); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if len(teams) == 0 {
			return nil
		}
		teamIDs := make([]int64, 0, len(teams))
		for _, team := range teams {
			teamIDs = append(teamIDs, team.Id)
		}
		if _, err := dao.TransportTeam.Ctx(ctx).WhereIn(dao.TransportTeam.Columns().Id, teamIDs).Data(do.TransportTeam{
			Status: string(teamStatusInvalid), Visible: 0, MemberCount: 0,
			InvalidatedAt: &now, InvalidReason: string(invalidReasonIdle),
		}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err := dao.TransportMember.Ctx(ctx).
			WhereIn(dao.TransportMember.Columns().TeamId, teamIDs).
			Where(dao.TransportMember.Columns().LeftAt + " IS NULL").
			Data(do.TransportMember{LeftAt: &now}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		invalidated = len(teamIDs)
		return nil
	})
	return invalidated, err
}
