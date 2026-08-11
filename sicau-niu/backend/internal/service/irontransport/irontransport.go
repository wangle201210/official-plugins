// Package irontransport implements persistent team-based iron-cow transport for
// the mini program. All writes are transactional and all state reads are bounded
// and batch-assembled.
package irontransport

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

const (
	// teamStatusForming marks a visible team that can still accept members.
	teamStatusForming = "forming"
	// teamStatusActive marks a team with an active transport session.
	teamStatusActive = "active"
	// teamStatusEnded marks a closed team whose memberships are no longer active.
	teamStatusEnded = "ended"

	// sessionStatusActive marks a session that accepts movement heartbeats.
	sessionStatusActive = "active"
	// sessionStatusTimeout marks a session closed by the idle timeout.
	sessionStatusTimeout = "idle_timeout"
	// sessionStatusEnded marks a session explicitly ended by its leader.
	sessionStatusEnded = "ended"

	// roleLeader grants team lifecycle authority to one member.
	roleLeader = "leader"
	// roleMember identifies a regular transport team participant.
	roleMember = "member"
)

// Config carries bounded transport timing and team-size defaults.
type Config struct {
	// IdleTimeout closes a session after this duration without valid movement.
	IdleTimeout time.Duration
	// MinTeamSize is the default minimum number of active members needed to start.
	MinTeamSize int
	// MaxTeamSize is the hard capacity of every transport team.
	MaxTeamSize int
}

// Service defines the persistent player-facing iron transport lifecycle.
type Service interface {
	// State returns the bounded transport projection visible to playerID. It
	// persists an overdue idle timeout before returning and reports query/write
	// bizerrs when state cannot be assembled.
	State(ctx context.Context, playerID int64) (*State, error)
	// CreateTeam creates a forming team led by playerID and returns the resulting
	// state. Invalid input, unavailable iron, duplicate membership and store
	// failures return transport bizerrs; a repeated RequestID is replayed.
	CreateTeam(ctx context.Context, playerID int64, in *CreateTeamInput) (*State, error)
	// JoinTeam adds playerID to a visible forming team and returns current state.
	// The optional request ID is player-scoped and idempotent.
	JoinTeam(ctx context.Context, playerID, teamID int64, requestID ...string) (*State, error)
	// LeaveTeam closes the player's active membership when no session is running.
	// A leader closes the whole team; the optional request ID is idempotent.
	LeaveTeam(ctx context.Context, playerID, teamID int64, requestID ...string) (*State, error)
	// Start creates the unique active session when playerID leads a ready team.
	// The optional request ID is idempotent and failures return transport bizerrs.
	Start(ctx context.Context, playerID, teamID int64, requestID ...string) (*State, error)
	// Heartbeat appends one bounded plausible movement point for a current team
	// member. It commits an overdue timeout before returning CodeNotActive.
	Heartbeat(ctx context.Context, playerID int64, in *HeartbeatInput) (*State, error)
	// End closes the leader's active session and team and returns the post-action
	// state. The optional request ID is idempotent.
	End(ctx context.Context, playerID, teamID int64, requestID ...string) (*State, error)
}

// serviceImpl implements Service against plugin-owned transport tables.
type serviceImpl struct {
	// idleTimeout is the accepted maximum interval between valid heartbeats.
	idleTimeout time.Duration
	// minTeamSize is the default start threshold.
	minTeamSize int
	// maxTeamSize is the hard membership capacity.
	maxTeamSize int
}

// Compile-time service contract assertion.
var _ Service = (*serviceImpl)(nil)

// CreateTeamInput carries one create-team command and its idempotency key.
type CreateTeamInput struct {
	// RequestID is the player-scoped idempotency key; empty disables replay.
	RequestID string
	// Name is the trimmed team display name.
	Name string
	// CampusID is one of cd, djy or ya.
	CampusID string
	// MinMembers overrides the default threshold when positive.
	MinMembers int
}

// HeartbeatInput carries one transport movement sample.
type HeartbeatInput struct {
	// RequestID is the player-scoped idempotency key; empty disables replay.
	RequestID string
	// TeamID identifies the player's active transport team.
	TeamID int64
	// Lat is the current GCJ-02 latitude.
	Lat float64
	// Lng is the current GCJ-02 longitude.
	Lng float64
}

// normalizeRequestID trims the optional first request ID and rejects values over
// 64 characters by returning an empty value.
func normalizeRequestID(values ...string) string {
	if len(values) == 0 {
		return ""
	}
	value := strings.TrimSpace(values[0])
	if len(value) > 64 {
		return ""
	}
	return value
}

// lockPlayer serializes transport commands for one existing player row.
func lockPlayer(ctx context.Context, playerID int64) error {
	var row struct {
		ID int64 `json:"id"`
	}
	if err := dao.User.Ctx(ctx).Fields(dao.User.Columns().Id).Where(dao.User.Columns().Id, playerID).LockUpdate().Scan(&row); err != nil {
		return err
	}
	if row.ID <= 0 {
		return errors.New("transport player not found")
	}
	return nil
}

// State is the complete bounded transport projection returned to one player.
type State struct {
	// Enabled reports whether a located iron cow is currently available.
	Enabled bool
	// Explanation is the stable user-facing transport rule summary.
	Explanation string
	// IdleTimeoutSec is the idle timeout in whole seconds.
	IdleTimeoutSec int
	// MinTeamSize is the default member threshold for new teams.
	MinTeamSize int
	// CampusID is the current player's team campus or the default campus.
	CampusID string
	// IronCow is the current location and movement projection.
	IronCow *IronCow
	// Teams contains at most the bounded visible teams.
	Teams []*Team
	// MyTeamID is zero when the current player has no active membership.
	MyTeamID int64
	// Session is nil when no relevant active or timed-out session exists.
	Session *Session
}

// IronCow is the player-safe current iron-cow movement projection.
type IronCow struct {
	// ID is the backing iron-cow ID; zero when transport is unavailable.
	ID int64
	// Name is the iron-cow display name.
	Name string
	// Lat is the latest GCJ-02 latitude.
	Lat float64
	// Lng is the latest GCJ-02 longitude.
	Lng float64
	// Status is waiting, forming, moving or timeout.
	Status string
	// MovedMeters is the current session cumulative accepted distance.
	MovedMeters float64
}

// Team is one visible transport team with batch-assembled active members.
type Team struct {
	// ID is the team primary key.
	ID int64
	// Name is the team display name.
	Name string
	// Code is the short public join code.
	Code string
	// CampusID is one of cd, djy or ya.
	CampusID string
	// LeaderID is the player allowed to start and end the session.
	LeaderID int64
	// MinMembers is the threshold required to start.
	MinMembers int
	// MaxMembers is the hard team capacity.
	MaxMembers int
	// Members contains current active memberships.
	Members []*Member
	// Visible reports whether the forming team appears in the public list.
	Visible bool
}

// Member is one active transport team participant.
type Member struct {
	// ID is the player ID.
	ID int64
	// Name is the player's display nickname.
	Name string
	// Avatar is the player's avatar URL or value.
	Avatar string
	// Role is leader or member.
	Role string
	// JoinedAt is the absolute membership creation time.
	JoinedAt *time.Time
}

// Session is one relevant transport session with a bounded chronological trace.
type Session struct {
	// ID is the session primary key.
	ID int64
	// TeamID is the owning team ID.
	TeamID int64
	// Status is active, idle_timeout or ended.
	Status string
	// StartedAt is the absolute session start time.
	StartedAt *time.Time
	// LastActiveAt is the last accepted heartbeat time.
	LastActiveAt *time.Time
	// MovedMeters is the cumulative accepted movement distance.
	MovedMeters float64
	// Trace contains at most the latest bounded movement points.
	Trace []*TrackPoint
}

// TrackPoint is one accepted GCJ-02 movement sample.
type TrackPoint struct {
	// Lat is the GCJ-02 latitude.
	Lat float64
	// Lng is the GCJ-02 longitude.
	Lng float64
	// At is the absolute recorded time.
	At *time.Time
}

// New creates a transport service and normalizes zero or inconsistent plain
// configuration to safe bounded defaults.
func New(config Config) Service {
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.MinTeamSize < 2 {
		config.MinTeamSize = 3
	}
	if config.MaxTeamSize < config.MinTeamSize {
		config.MaxTeamSize = 6
	}
	return &serviceImpl{idleTimeout: config.IdleTimeout, minTeamSize: config.MinTeamSize, maxTeamSize: config.MaxTeamSize}
}

// validCampus reports whether value is one supported stable campus ID.
func validCampus(value string) bool {
	switch strings.TrimSpace(value) {
	case "cd", "djy", "ya":
		return true
	default:
		return false
	}
}

// validCoordinate reports whether a non-zero latitude/longitude pair is valid.
func validCoordinate(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !(lat == 0 && lng == 0)
}

// haversineMeters returns the approximate surface distance between two GCJ-02
// points in meters; both points share the same coordinate system.
func haversineMeters(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadius = 6371000.0
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	deltaPhi := (lat2 - lat1) * math.Pi / 180
	deltaLambda := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(deltaPhi/2)*math.Sin(deltaPhi/2) + math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2)*math.Sin(deltaLambda/2)
	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// transaction executes one transport mutation in a GoFrame transaction.
func (s *serviceImpl) transaction(ctx context.Context, fn func(context.Context) error) error {
	return dao.TransportTeam.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error { return fn(ctx) })
}

// activeMemberModel builds the reusable active-membership query for playerID.
func activeMemberModel(ctx context.Context, playerID int64) *gdb.Model {
	return dao.TransportMember.Ctx(ctx).
		Where(dao.TransportMember.Columns().UserId, playerID).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL")
}

// markTeamEnded closes one team and all of its still-active memberships.
func markTeamEnded(ctx context.Context, teamID int64, endedAt time.Time) error {
	if _, err := dao.TransportTeam.Ctx(ctx).Where(do.TransportTeam{Id: teamID}).Data(do.TransportTeam{Status: teamStatusEnded}).Update(); err != nil {
		return err
	}
	_, err := dao.TransportMember.Ctx(ctx).
		Where(dao.TransportMember.Columns().TeamId, teamID).
		Where(dao.TransportMember.Columns().LeftAt + " IS NULL").
		Data(do.TransportMember{LeftAt: &endedAt}).
		Update()
	return err
}
