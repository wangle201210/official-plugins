// irontransport_state.go assembles the bounded player-visible transport state.
// Teams, members, users and tracks are loaded through fixed-count batch queries;
// reads also persist an overdue active-session timeout before projecting state.

package irontransport

import (
	"context"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

const (
	// stateTeamCap bounds visible team rows per state response.
	stateTeamCap = 20
	// stateTrackCap bounds recent chronological trace points per state response.
	stateTrackCap = 100
)

// State returns the current bounded transport projection for playerID.
func (s *serviceImpl) State(ctx context.Context, playerID int64) (*State, error) {
	if err := s.expireIdleSession(ctx, time.Now()); err != nil {
		return nil, err
	}
	iron, err := currentIron(ctx)
	if err != nil {
		return nil, err
	}
	state := &State{
		Enabled:        iron != nil,
		Explanation:    "铁牛由小队协作搬运，移动距离和活跃时间由服务端轨迹记录。",
		IdleTimeoutSec: int(s.idleTimeout.Seconds()),
		MinTeamSize:    s.minTeamSize,
		CampusID:       "cd",
		Teams:          []*Team{},
		IronCow:        &IronCow{Status: "waiting"},
	}
	if iron != nil {
		state.IronCow.ID = iron.Id
		state.IronCow.Name = iron.Name
		state.IronCow.Lat = iron.LastLat
		state.IronCow.Lng = iron.LastLng
	}
	myMember, err := activeMembership(ctx, playerID)
	if err != nil {
		return nil, err
	}

	teams := make([]*entitymodel.TransportTeam, 0)
	err = dao.TransportTeam.Ctx(ctx).
		WhereIn(dao.TransportTeam.Columns().Status, []string{teamStatusForming, teamStatusActive}).
		Where(do.TransportTeam{Visible: 1}).
		OrderDesc(dao.TransportTeam.Columns().Id).
		Limit(stateTeamCap).
		Scan(&teams)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if myMember != nil {
		ownTeamIncluded := false
		for _, team := range teams {
			if team.Id == myMember.TeamId {
				ownTeamIncluded = true
				break
			}
		}
		if !ownTeamIncluded {
			var ownTeam *entitymodel.TransportTeam
			if err = dao.TransportTeam.Ctx(ctx).
				Where(do.TransportTeam{Id: myMember.TeamId}).
				WhereIn(dao.TransportTeam.Columns().Status, []string{teamStatusForming, teamStatusActive}).
				Limit(1).
				Scan(&ownTeam); err != nil {
				return nil, bizerr.WrapCode(err, CodeQueryFailed)
			}
			if ownTeam == nil {
				myMember = nil
			} else {
				if len(teams) >= stateTeamCap {
					teams = teams[:stateTeamCap-1]
				}
				teams = append([]*entitymodel.TransportTeam{ownTeam}, teams...)
			}
		}
	}
	teamIDs := make([]int64, 0, len(teams))
	for _, team := range teams {
		teamIDs = append(teamIDs, team.Id)
	}
	membersByTeam, err := batchMembers(ctx, teamIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range teams {
		state.Teams = append(state.Teams, &Team{
			ID: row.Id, Name: row.Name, Code: row.Code, CampusID: row.CampusId,
			LeaderID: row.LeaderUserId, MinMembers: row.MinMembers, MaxMembers: row.MaxMembers,
			Members: membersByTeam[row.Id], Visible: row.Visible == 1,
		})
	}
	if myMember != nil {
		state.MyTeamID = myMember.TeamId
		for _, team := range teams {
			if team.Id == myMember.TeamId {
				state.CampusID = team.CampusId
				break
			}
		}
	}

	session, err := stateSession(ctx, state.MyTeamID)
	if err != nil {
		return nil, err
	}
	if session != nil {
		state.Session, err = loadSession(ctx, session)
		if err != nil {
			return nil, err
		}
		state.IronCow.MovedMeters = session.MovedMeters
		state.IronCow.Lat = session.LastLat
		state.IronCow.Lng = session.LastLng
		switch session.Status {
		case sessionStatusActive:
			state.IronCow.Status = "moving"
		case sessionStatusTimeout:
			state.IronCow.Status = "timeout"
		}
	} else if len(teams) > 0 {
		state.IronCow.Status = "forming"
	}
	return state, nil
}

// activeMembership loads the player's single current membership independently
// from the public team cap so an older own team cannot disappear from state.
func activeMembership(ctx context.Context, playerID int64) (*entitymodel.TransportMember, error) {
	if playerID <= 0 {
		return nil, nil
	}
	var member *entitymodel.TransportMember
	if err := activeMemberModel(ctx, playerID).Limit(1).Scan(&member); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return member, nil
}

// currentIron returns the latest valid located iron cow or nil when transport is
// unavailable.
func currentIron(ctx context.Context) (*entitymodel.Iron, error) {
	var iron *entitymodel.Iron
	err := dao.Iron.Ctx(ctx).
		Where(dao.Iron.Columns().LocatedAt+" IS NOT NULL").
		WhereNot(dao.Iron.Columns().LastLat, 0).
		WhereNot(dao.Iron.Columns().LastLng, 0).
		OrderDesc(dao.Iron.Columns().LocatedAt).
		Limit(1).
		Scan(&iron)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return iron, nil
}

// batchMembers loads all active members for bounded teamIDs and batch-assembles
// their public player identity fields without N+1 queries.
func batchMembers(ctx context.Context, teamIDs []int64) (map[int64][]*Member, error) {
	result := make(map[int64][]*Member, len(teamIDs))
	if len(teamIDs) == 0 {
		return result, nil
	}
	rows := make([]*entitymodel.TransportMember, 0)
	err := dao.TransportMember.Ctx(ctx).
		WhereIn(dao.TransportMember.Columns().TeamId, teamIDs).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL").
		OrderAsc(dao.TransportMember.Columns().Id).
		Limit(stateTeamCap * 12).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	userIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
	}
	users := make([]*entitymodel.User, 0)
	if len(userIDs) > 0 {
		if err = dao.User.Ctx(ctx).Fields(dao.User.Columns().Id, dao.User.Columns().Nickname, dao.User.Columns().Avatar).WhereIn(dao.User.Columns().Id, userIDs).Scan(&users); err != nil {
			return nil, bizerr.WrapCode(err, CodeQueryFailed)
		}
	}
	usersByID := make(map[int64]*entitymodel.User, len(users))
	for _, user := range users {
		usersByID[user.Id] = user
	}
	for _, row := range rows {
		member := &Member{ID: row.UserId, Role: row.Role, JoinedAt: row.JoinedAt}
		if user := usersByID[row.UserId]; user != nil {
			member.Name = user.Nickname
			member.Avatar = user.Avatar
		}
		result[row.TeamId] = append(result[row.TeamId], member)
	}
	return result, nil
}

// stateSession returns the player's most recent team session, or the latest
// globally relevant active/timed-out session when the player has no team.
func stateSession(ctx context.Context, myTeamID int64) (*entitymodel.TransportSession, error) {
	model := dao.TransportSession.Ctx(ctx)
	if myTeamID > 0 {
		model = model.Where(dao.TransportSession.Columns().TeamId, myTeamID)
	} else {
		model = model.WhereIn(dao.TransportSession.Columns().Status, []string{sessionStatusActive, sessionStatusTimeout})
	}
	var session *entitymodel.TransportSession
	if err := model.OrderDesc(dao.TransportSession.Columns().Id).Limit(1).Scan(&session); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return session, nil
}

// loadSession maps one session and its bounded latest track rows to a
// chronological player projection.
func loadSession(ctx context.Context, row *entitymodel.TransportSession) (*Session, error) {
	tracks := make([]*entitymodel.TransportTrack, 0)
	err := dao.TransportTrack.Ctx(ctx).
		Where(dao.TransportTrack.Columns().SessionId, row.Id).
		OrderDesc(dao.TransportTrack.Columns().Id).
		Limit(stateTrackCap).
		Scan(&tracks)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	trace := make([]*TrackPoint, 0, len(tracks))
	for i := len(tracks) - 1; i >= 0; i-- {
		trace = append(trace, &TrackPoint{Lat: tracks[i].Lat, Lng: tracks[i].Lng, At: tracks[i].RecordedAt})
	}
	return &Session{ID: row.Id, TeamID: row.TeamId, Status: row.Status, StartedAt: row.StartedAt, LastActiveAt: row.LastActiveAt, MovedMeters: row.MovedMeters, Trace: trace}, nil
}

// expireIdleSession persists timeout state when the unique active session has
// exceeded the configured idle duration.
func (s *serviceImpl) expireIdleSession(ctx context.Context, now time.Time) error {
	var session *entitymodel.TransportSession
	err := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{Status: sessionStatusActive}).Limit(1).Scan(&session)
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if session == nil || session.LastActiveAt == nil || now.Sub(*session.LastActiveAt) <= s.idleTimeout {
		return nil
	}
	return s.transaction(ctx, func(ctx context.Context) error {
		locked, err := activeSessionForTeam(ctx, session.TeamId, true)
		if err != nil || locked == nil {
			return err
		}
		if locked.LastActiveAt == nil || now.Sub(*locked.LastActiveAt) <= s.idleTimeout {
			return nil
		}
		return timeoutSession(ctx, locked, now)
	})
}

// timeoutSession closes one locked session and its team at now.
func timeoutSession(ctx context.Context, session *entitymodel.TransportSession, now time.Time) error {
	if _, err := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{Id: session.Id}).Data(do.TransportSession{Status: sessionStatusTimeout, EndedAt: &now}).Update(); err != nil {
		return bizerr.WrapCode(err, CodeWriteFailed)
	}
	if err := markTeamEnded(ctx, session.TeamId, now); err != nil {
		return bizerr.WrapCode(err, CodeWriteFailed)
	}
	return nil
}
