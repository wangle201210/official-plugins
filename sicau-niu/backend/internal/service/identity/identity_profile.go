// identity_profile.go implements reading and updating the current player's own
// identity profile, including identity-type enum validation, college existence
// validation and grade/graduation-year numeric validation.

package identity

import (
	"context"
	"strings"
	"time"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// graduationYearLowerBound and graduationYearUpperOffset bound an accepted
// graduation year so obvious garbage is rejected while leaving room for past
// alumni and near-future graduates.
const (
	graduationYearLowerBound  = 1900
	graduationYearUpperOffset = 10
)

// ProfileOutput defines the current player's identity profile.
type ProfileOutput struct {
	// Id is the player ID.
	Id int64
	// Openid is the WeChat openid.
	Openid string
	// Phone is the bound phone number, empty when not yet bound.
	Phone string
	// Nickname is the player nickname.
	Nickname string
	// Avatar is the player avatar URL.
	Avatar string
	// IdentityType is the identity tag string, empty when not yet set.
	IdentityType string
	// CollegeId is the selected college ID, 0 when none.
	CollegeId int64
	// Grade is the grade number, 0 when unset.
	Grade int
	// GraduationYear is the graduation year, 0 when unset.
	GraduationYear int
	// CreatedAt is the creation time as a Unix timestamp in milliseconds.
	CreatedAt *int64
	// UpdatedAt is the update time as a Unix timestamp in milliseconds.
	UpdatedAt *int64
}

// UpdateProfileInput defines the editable identity profile fields.
type UpdateProfileInput struct {
	// Nickname is the player nickname.
	Nickname string
	// Avatar is the player avatar URL.
	Avatar string
	// IdentityType is the identity tag: student / alumni / friend.
	IdentityType string
	// CollegeId is the selected college ID; required for students.
	CollegeId int64
	// Grade is the grade number; required positive for students.
	Grade int
	// GraduationYear is the graduation year; optional, validated when non-zero.
	GraduationYear int
}

// GetProfile returns the current player's own identity profile.
func (s *serviceImpl) GetProfile(ctx context.Context, playerID int64) (*ProfileOutput, error) {
	player, err := s.loadPlayer(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &ProfileOutput{
		Id:             player.Id,
		Openid:         player.Openid,
		Phone:          player.Phone,
		Nickname:       player.Nickname,
		Avatar:         player.Avatar,
		IdentityType:   player.IdentityType,
		CollegeId:      player.CollegeId,
		Grade:          player.Grade,
		GraduationYear: player.GraduationYear,
		CreatedAt:      apitime.Milli(player.CreatedAt),
		UpdatedAt:      apitime.Milli(player.UpdatedAt),
	}, nil
}

// UpdateProfile validates and updates the current player's own identity profile.
func (s *serviceImpl) UpdateProfile(ctx context.Context, playerID int64, in *UpdateProfileInput) error {
	if in == nil {
		return bizerr.NewCode(CodeIdentityTypeInvalid)
	}

	identityType := IdentityType(strings.TrimSpace(in.IdentityType))
	if !identityType.valid() {
		return bizerr.NewCode(CodeIdentityTypeInvalid)
	}

	collegeID, grade, err := s.validateCollegeAndGrade(ctx, identityType, in)
	if err != nil {
		return err
	}
	if err = validateGraduationYear(in.GraduationYear); err != nil {
		return err
	}

	if _, err = s.loadPlayer(ctx, playerID); err != nil {
		return err
	}

	_, err = dao.User.Ctx(ctx).
		Where(do.User{Id: playerID}).
		Data(do.User{
			Nickname:       strings.TrimSpace(in.Nickname),
			Avatar:         strings.TrimSpace(in.Avatar),
			IdentityType:   identityType.String(),
			CollegeId:      collegeID,
			Grade:          grade,
			GraduationYear: in.GraduationYear,
		}).
		Update()
	if err != nil {
		return bizerr.WrapCode(err, CodePlayerWriteFailed)
	}
	return nil
}

// validateCollegeAndGrade validates the college and grade fields against the
// identity type. Students must select an existing college and a positive grade;
// alumni and friends may omit both, in which case they are cleared.
func (s *serviceImpl) validateCollegeAndGrade(
	ctx context.Context,
	identityType IdentityType,
	in *UpdateProfileInput,
) (int64, int, error) {
	if !identityType.requiresCollege() {
		if in.CollegeId <= 0 {
			return 0, 0, nil
		}
		if err := s.ensureCollegeExists(ctx, in.CollegeId); err != nil {
			return 0, 0, err
		}
		return in.CollegeId, maxInt(in.Grade, 0), nil
	}

	if in.CollegeId <= 0 {
		return 0, 0, bizerr.NewCode(CodeCollegeRequired)
	}
	if err := s.ensureCollegeExists(ctx, in.CollegeId); err != nil {
		return 0, 0, err
	}
	if in.Grade <= 0 {
		return 0, 0, bizerr.NewCode(CodeGradeInvalid)
	}
	return in.CollegeId, in.Grade, nil
}

// ensureCollegeExists returns CodeCollegeInvalid when the college does not exist.
func (s *serviceImpl) ensureCollegeExists(ctx context.Context, collegeID int64) error {
	exists, err := s.collegeSvc.Exists(ctx, collegeID)
	if err != nil {
		return err
	}
	if !exists {
		return bizerr.NewCode(CodeCollegeInvalid)
	}
	return nil
}

// loadPlayer loads one player row by primary key or returns CodePlayerNotFound.
func (s *serviceImpl) loadPlayer(ctx context.Context, playerID int64) (*entitymodel.User, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodePlayerNotFound)
	}
	var player *entitymodel.User
	err := dao.User.Ctx(ctx).Where(do.User{Id: playerID}).Scan(&player)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodePlayerQueryFailed)
	}
	if player == nil {
		return nil, bizerr.NewCode(CodePlayerNotFound)
	}
	return player, nil
}

// validateGraduationYear validates an optional graduation year. Zero means unset
// and is accepted; non-zero values must fall within a sane range.
func validateGraduationYear(year int) error {
	if year == 0 {
		return nil
	}
	upperBound := time.Now().Year() + graduationYearUpperOffset
	if year < graduationYearLowerBound || year > upperBound {
		return bizerr.NewCode(CodeGraduationYearInvalid)
	}
	return nil
}

// maxInt returns the larger of a and b.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
