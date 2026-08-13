// Package irontransport implements persistent cloud-moving teams and immutable
// player contribution reports for the mini program.
package irontransport

import (
	"context"
	"time"
)

type teamStatus string
type memberRole string
type invalidReason string

const (
	teamStatusEffective teamStatus    = "effective"
	teamStatusInvalid   teamStatus    = "invalid"
	roleCreator         memberRole    = "creator"
	roleMember          memberRole    = "member"
	invalidReasonIdle   invalidReason = "inactive_72h"
	defaultActivityKey                = "default"
	maxEffectiveTeams                 = 120
	dailyReportLimit                  = 12
	defaultPageNum                    = 1
	defaultPageSize                   = 20
	maxPageSize                       = 100
	transportLockKey                  = int64(736_943_201)
)

var defaultInactiveAfter = 72 * time.Hour

// Config carries the inactivity policy. Zero values use the product rule.
type Config struct {
	InactiveAfter time.Duration
}

// Service defines all player and operator cloud-moving capabilities.
type Service interface {
	// State returns all effective teams plus the caller's membership and daily quota.
	State(ctx context.Context, playerID int64) (*State, error)
	// CreateTeam creates one effectively unique named team and binds its creator.
	CreateTeam(ctx context.Context, playerID int64, in *CreateTeamInput) (*State, error)
	// JoinTeam binds the caller to one effective team without distance constraints.
	JoinTeam(ctx context.Context, playerID, teamID int64, requestID string) (*State, error)
	// GetTeam returns one player-visible effective team without exposing report coordinates.
	GetTeam(ctx context.Context, playerID, teamID int64) (*Team, error)
	// ListMembers returns a bounded page only when the caller currently belongs to the team.
	ListMembers(ctx context.Context, playerID, teamID int64, in *PageInput) (*MemberList, error)
	// Report persists one successful immutable location fact and updates contribution snapshots.
	Report(ctx context.Context, playerID int64, in *ReportInput) (*ReportResult, error)
	// ListMyReports returns only the caller's own raw location facts with bounded pagination.
	ListMyReports(ctx context.Context, playerID int64, in *PageInput) (*ReportList, error)
	// ListAdminTeams returns an operator-authorized bounded team history page.
	ListAdminTeams(ctx context.Context, in *AdminTeamListInput) (*AdminTeamList, error)
	// RenameTeam updates the operator-managed name, preserving effective-name uniqueness without refreshing activity.
	RenameTeam(ctx context.Context, teamID int64, name string) error
	// AdminStats aggregates team and contribution facts for an optional Beijing day.
	AdminStats(ctx context.Context, activityDate string) (*Stats, error)
	// ListAdminReports returns an operator-authorized bounded raw report audit page.
	ListAdminReports(ctx context.Context, in *AdminReportListInput) (*AdminReportList, error)
	// ExpireInactive closes effective teams and memberships idle for the configured duration.
	ExpireInactive(ctx context.Context, now time.Time) (int, error)
}

type serviceImpl struct {
	inactiveAfter time.Duration
	now           func() time.Time
}

var _ Service = (*serviceImpl)(nil)

type CreateTeamInput struct {
	RequestID string
	Name      string
}

type ReportInput struct {
	RequestID string
	TeamID    int64
	Lat       float64
	Lng       float64
	SampledAt time.Time
}

type PageInput struct {
	PageNum  int
	PageSize int
}

type State struct {
	Explanation          string  `json:"explanation"`
	MaxEffectiveTeams    int     `json:"maxEffectiveTeams"`
	DailyReportLimit     int     `json:"dailyReportLimit"`
	TodayReportCount     int     `json:"todayReportCount"`
	TodayReportRemaining int     `json:"todayReportRemaining"`
	Teams                []*Team `json:"teams"`
	MyTeam               *Team   `json:"myTeam"`
}

type Team struct {
	ID                      int64      `json:"id"`
	Name                    string     `json:"name"`
	CreatorID               int64      `json:"creatorId"`
	CreatorName             string     `json:"creatorName"`
	MemberCount             int        `json:"memberCount"`
	TotalContributionMeters int64      `json:"totalContributionMeters"`
	MyContributionMeters    int64      `json:"myContributionMeters"`
	HasReportBaseline       bool       `json:"hasReportBaseline"`
	LastActiveAt            *time.Time `json:"lastActiveAt"`
	CreatedAt               *time.Time `json:"createdAt"`
	Mine                    bool       `json:"mine"`
}

type Member struct {
	ID                 int64
	UserID             int64
	Name               string
	Avatar             string
	Role               string
	ContributionMeters int64
	JoinedAt           *time.Time
}

type MemberList struct {
	List  []*Member
	Total int
}

type ReportResult struct {
	ID                   int64
	TeamID               int64
	StartLat             *float64
	StartLng             *float64
	EndLat               float64
	EndLng               float64
	ContributionMeters   int64
	UserTotalMeters      int64
	TeamTotalMeters      int64
	TodayReportCount     int
	TodayReportRemaining int
	AcceptedAt           *time.Time
}

type Report struct {
	*ReportResult
	ActivityDate string
	SampledAt    *time.Time
}

type ReportList struct {
	List  []*Report
	Total int
}

type AdminTeamListInput struct {
	PageNum  int
	PageSize int
	Name     string
	Status   string
}

type AdminTeam struct {
	*Team
	Status        string
	InvalidatedAt *time.Time
	InvalidReason string
}

type AdminTeamList struct {
	List  []*AdminTeam
	Total int
}

type Stats struct {
	EffectiveTeamCount int
	InvalidTeamCount   int
	ActiveMemberCount  int
	ReportCount        int
	ContributionMeters int64
}

type AdminReportListInput struct {
	PageNum      int
	PageSize     int
	TeamID       int64
	UserID       int64
	ActivityDate string
}

type AdminReport struct {
	*Report
	TeamName string
	MemberID int64
	UserID   int64
	UserName string
}

type AdminReportList struct {
	List  []*AdminReport
	Total int
}

// New creates a stateless cloud-moving service with bounded defaults.
func New(config Config) Service {
	if config.InactiveAfter <= 0 {
		config.InactiveAfter = defaultInactiveAfter
	}
	return &serviceImpl{inactiveAfter: config.InactiveAfter, now: time.Now}
}

func (s *serviceImpl) nowTime() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return time.Now()
}
