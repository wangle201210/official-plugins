// irontransport_actions.go implements transactional team and session commands.
// Player row locks serialize commands per player, team/session row locks protect
// shared state, and persisted request IDs provide stable replay semantics.

package irontransport

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// CreateTeam creates a visible forming team led by playerID and returns the
// resulting bounded state projection.
func (s *serviceImpl) CreateTeam(ctx context.Context, playerID int64, in *CreateTeamInput) (*State, error) {
	if playerID <= 0 || in == nil || strings.TrimSpace(in.Name) == "" || !validCampus(in.CampusID) {
		return nil, bizerr.NewCode(CodeInvalidState)
	}
	minMembers := in.MinMembers
	if minMembers <= 0 {
		minMembers = s.minTeamSize
	}
	if minMembers < 2 || minMembers > s.maxTeamSize {
		return nil, bizerr.NewCode(CodeInvalidState)
	}

	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		requestID := normalizeRequestID(in.RequestID)
		if requestID != "" {
			count, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{LeaderUserId: playerID, CreateRequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		count, err := activeMemberModel(ctx, playerID).Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if count > 0 {
			return bizerr.NewCode(CodeAlreadyInTeam)
		}
		iron, err := currentIron(ctx)
		if err != nil {
			return err
		}
		if iron == nil {
			return bizerr.NewCode(CodeUnavailable)
		}
		teamID, err := dao.TransportTeam.Ctx(ctx).Data(do.TransportTeam{
			Code:            strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:6]),
			Name:            strings.TrimSpace(in.Name),
			CampusId:        strings.TrimSpace(in.CampusID),
			LeaderUserId:    playerID,
			CreateRequestId: requestID,
			IronId:          iron.Id,
			Status:          teamStatusForming,
			MinMembers:      minMembers,
			MaxMembers:      s.maxTeamSize,
			Visible:         1,
		}).InsertAndGetId()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		_, err = dao.TransportMember.Ctx(ctx).Data(do.TransportMember{TeamId: teamID, UserId: playerID, JoinRequestId: requestID, Role: roleLeader}).Insert()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// JoinTeam adds playerID to one visible forming team when capacity permits.
func (s *serviceImpl) JoinTeam(ctx context.Context, playerID, teamID int64, requestIDs ...string) (*State, error) {
	requestID := normalizeRequestID(requestIDs...)
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if requestID != "" {
			count, err := dao.TransportMember.Ctx(ctx).Where(do.TransportMember{UserId: playerID, JoinRequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		var team *entitymodel.TransportTeam
		if err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil || team.Status != teamStatusForming || team.Visible != 1 {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		var membership *entitymodel.TransportMember
		if err := activeMemberModel(ctx, playerID).Scan(&membership); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if membership != nil {
			if membership.TeamId == teamID {
				return nil
			}
			return bizerr.NewCode(CodeAlreadyInTeam)
		}
		count, err := dao.TransportMember.Ctx(ctx).
			Where(dao.TransportMember.Columns().TeamId, teamID).
			Where(dao.TransportMember.Columns().LeftAt + " IS NULL").
			Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if count >= team.MaxMembers {
			return bizerr.NewCode(CodeTeamFull)
		}
		_, err = dao.TransportMember.Ctx(ctx).Data(do.TransportMember{TeamId: teamID, UserId: playerID, JoinRequestId: requestID, Role: roleMember}).Insert()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// LeaveTeam removes a member from a forming team or closes the whole forming team
// when its leader leaves.
func (s *serviceImpl) LeaveTeam(ctx context.Context, playerID, teamID int64, requestIDs ...string) (*State, error) {
	requestID := normalizeRequestID(requestIDs...)
	now := time.Now()
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if requestID != "" {
			count, err := dao.TransportMember.Ctx(ctx).Where(do.TransportMember{UserId: playerID, LeaveRequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		var member *entitymodel.TransportMember
		if err := activeMemberModel(ctx, playerID).Where(dao.TransportMember.Columns().TeamId, teamID).LockUpdate().Scan(&member); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if member == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		active, err := activeSessionForTeam(ctx, teamID, true)
		if err != nil {
			return err
		}
		if active != nil {
			return bizerr.NewCode(CodeInvalidState)
		}
		if member.Role == roleLeader {
			if _, err = dao.TransportMember.Ctx(ctx).Where(do.TransportMember{Id: member.Id}).Data(do.TransportMember{LeaveRequestId: requestID}).Update(); err != nil {
				return bizerr.WrapCode(err, CodeWriteFailed)
			}
			if err = markTeamEnded(ctx, teamID, now); err != nil {
				return bizerr.WrapCode(err, CodeWriteFailed)
			}
			return nil
		}
		_, err = dao.TransportMember.Ctx(ctx).Where(do.TransportMember{Id: member.Id}).Data(do.TransportMember{LeftAt: &now, LeaveRequestId: requestID}).Update()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// Start creates the unique active transport session for a ready team led by
// playerID and records the initial iron-cow position.
func (s *serviceImpl) Start(ctx context.Context, playerID, teamID int64, requestIDs ...string) (*State, error) {
	requestID := normalizeRequestID(requestIDs...)
	now := time.Now()
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if requestID != "" {
			count, err := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{StartedByUserId: playerID, StartRequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		var team *entitymodel.TransportTeam
		if err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		if team.LeaderUserId != playerID {
			return bizerr.NewCode(CodeForbidden)
		}
		if team.Status != teamStatusForming {
			return bizerr.NewCode(CodeInvalidState)
		}
		count, err := dao.TransportMember.Ctx(ctx).
			Where(dao.TransportMember.Columns().TeamId, teamID).
			Where(dao.TransportMember.Columns().LeftAt + " IS NULL").Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if count < team.MinMembers {
			return bizerr.NewCode(CodeTeamNotReady)
		}
		activeCount, err := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{Status: sessionStatusActive}).Count()
		if err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if activeCount > 0 {
			return bizerr.NewCode(CodeInvalidState)
		}
		var iron *entitymodel.Iron
		if err = dao.Iron.Ctx(ctx).Where(do.Iron{Id: team.IronId}).Scan(&iron); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if iron == nil || !validCoordinate(iron.LastLat, iron.LastLng) {
			return bizerr.NewCode(CodeUnavailable)
		}
		sessionID, err := dao.TransportSession.Ctx(ctx).Data(do.TransportSession{
			TeamId: teamID, IronId: team.IronId, Status: sessionStatusActive,
			StartedByUserId: playerID, StartRequestId: requestID,
			StartedAt: &now, LastActiveAt: &now, LastLat: iron.LastLat, LastLng: iron.LastLng,
		}).InsertAndGetId()
		if err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportTrack.Ctx(ctx).Data(do.TransportTrack{SessionId: sessionID, UserId: playerID, Lat: iron.LastLat, Lng: iron.LastLng, RecordedAt: &now}).Insert(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{Status: teamStatusActive}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// Heartbeat validates membership and plausible movement, then appends one trace
// point and updates cumulative distance. An overdue session is timed out first.
func (s *serviceImpl) Heartbeat(ctx context.Context, playerID int64, in *HeartbeatInput) (*State, error) {
	if in == nil || !validCoordinate(in.Lat, in.Lng) {
		return nil, bizerr.NewCode(CodeMovementInvalid)
	}
	now := time.Now()
	timedOut := false
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		requestID := normalizeRequestID(in.RequestID)
		if requestID != "" {
			count, err := dao.TransportTrack.Ctx(ctx).Where(do.TransportTrack{UserId: playerID, RequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		var member *entitymodel.TransportMember
		if err := activeMemberModel(ctx, playerID).Where(dao.TransportMember.Columns().TeamId, in.TeamID).LockUpdate().Scan(&member); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if member == nil {
			return bizerr.NewCode(CodeForbidden)
		}
		session, err := activeSessionForTeam(ctx, in.TeamID, true)
		if err != nil {
			return err
		}
		if session == nil {
			return bizerr.NewCode(CodeNotActive)
		}
		if session.LastActiveAt != nil && now.Sub(*session.LastActiveAt) > s.idleTimeout {
			if err = timeoutSession(ctx, session, now); err != nil {
				return err
			}
			timedOut = true
			return nil
		}
		step := haversineMeters(session.LastLat, session.LastLng, in.Lat, in.Lng)
		if step > 250 {
			return bizerr.NewCode(CodeMovementInvalid)
		}
		moved := session.MovedMeters + step
		if _, err = dao.TransportSession.Ctx(ctx).Where(do.TransportSession{Id: session.Id}).Data(do.TransportSession{
			LastActiveAt: &now, LastLat: in.Lat, LastLng: in.Lng, MovedMeters: moved,
		}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportMember.Ctx(ctx).Where(do.TransportMember{Id: member.Id}).Data(do.TransportMember{LastHeartbeatAt: &now}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if _, err = dao.TransportTrack.Ctx(ctx).Data(do.TransportTrack{SessionId: session.Id, UserId: playerID, RequestId: requestID, Lat: in.Lat, Lng: in.Lng, DistanceMeters: step, RecordedAt: &now}).Insert(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if timedOut {
		return nil, bizerr.NewCode(CodeNotActive)
	}
	return s.State(ctx, playerID)
}

// End closes the leader's active session and team and returns the post-action
// state projection.
func (s *serviceImpl) End(ctx context.Context, playerID, teamID int64, requestIDs ...string) (*State, error) {
	requestID := normalizeRequestID(requestIDs...)
	now := time.Now()
	err := s.transaction(ctx, func(ctx context.Context) error {
		if err := lockPlayer(ctx, playerID); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if requestID != "" {
			count, err := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{EndedByUserId: playerID, EndRequestId: requestID}).Count()
			if err != nil {
				return bizerr.WrapCode(err, CodeQueryFailed)
			}
			if count > 0 {
				return nil
			}
		}
		var team *entitymodel.TransportTeam
		if err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).LockUpdate().Scan(&team); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if team == nil {
			return bizerr.NewCode(CodeTeamNotFound)
		}
		if team.LeaderUserId != playerID {
			return bizerr.NewCode(CodeForbidden)
		}
		session, err := activeSessionForTeam(ctx, teamID, true)
		if err != nil {
			return err
		}
		if session == nil {
			return bizerr.NewCode(CodeNotActive)
		}
		if _, err = dao.TransportSession.Ctx(ctx).Where(do.TransportSession{Id: session.Id}).Data(do.TransportSession{Status: sessionStatusEnded, EndedAt: &now, EndedByUserId: playerID, EndRequestId: requestID}).Update(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if err = markTeamEnded(ctx, teamID, now); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, playerID)
}

// activeSessionForTeam loads one active team session and optionally locks it for
// update inside the caller's transaction.
func activeSessionForTeam(ctx context.Context, teamID int64, lock bool) (*entitymodel.TransportSession, error) {
	model := dao.TransportSession.Ctx(ctx).Where(do.TransportSession{TeamId: teamID, Status: sessionStatusActive})
	if lock {
		model = model.LockUpdate()
	}
	var session *entitymodel.TransportSession
	if err := model.Scan(&session); err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return session, nil
}
