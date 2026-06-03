// identity_profile_test.go verifies graduation-year validation (pure logic) and
// the identity profile update flow against a gated PostgreSQL database. Profile
// update validation covers the student college/grade requirement, non-existent
// college rejection and a successful update.

package identity

import (
	"context"
	"testing"
	"time"

	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// TestValidateGraduationYear verifies zero is accepted as unset, in-range years
// pass and out-of-range years are rejected.
func TestValidateGraduationYear(t *testing.T) {
	if err := validateGraduationYear(0); err != nil {
		t.Fatalf("expected zero graduation year to be accepted, got %v", err)
	}

	currentYear := time.Now().Year()
	valid := []int{graduationYearLowerBound, currentYear, currentYear + graduationYearUpperOffset}
	for _, year := range valid {
		if err := validateGraduationYear(year); err != nil {
			t.Fatalf("expected graduation year %d to be valid, got %v", year, err)
		}
	}

	invalid := []int{graduationYearLowerBound - 1, currentYear + graduationYearUpperOffset + 1, -5}
	for _, year := range invalid {
		err := validateGraduationYear(year)
		assertBizCode(t, err, CodeGraduationYearInvalid.RuntimeCode())
	}
}

// TestUpdateProfileRejectsInvalidIdentityType verifies an unknown identity type
// is rejected before any store access. Pure-logic assertion that runs even when
// the DB harness skips.
func TestUpdateProfileRejectsInvalidIdentityType(t *testing.T) {
	ctx := context.Background()
	svc := newIdentityServiceForTest(t, "openid-bad-identity")

	err := svc.UpdateProfile(ctx, 1, &UpdateProfileInput{IdentityType: "teacher"})
	assertBizCode(t, err, CodeIdentityTypeInvalid.RuntimeCode())
}

// TestUpdateProfileStudentValidation verifies student profile validation:
// missing college is rejected, a non-existent college is rejected, and a valid
// student update with an existing college and grade succeeds and persists.
func TestUpdateProfileStudentValidation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLIdentityDB(t, ctx)

	svc, player := provisionPlayer(t, ctx, "openid-profile-student")

	// Student without a college is rejected.
	err := svc.UpdateProfile(ctx, player, &UpdateProfileInput{
		IdentityType: IdentityTypeStudent.String(),
		Grade:        2,
	})
	assertBizCode(t, err, CodeCollegeRequired.RuntimeCode())

	// Student selecting a non-existent college is rejected.
	err = svc.UpdateProfile(ctx, player, &UpdateProfileInput{
		IdentityType: IdentityTypeStudent.String(),
		CollegeId:    999999,
		Grade:        2,
	})
	assertBizCode(t, err, CodeCollegeInvalid.RuntimeCode())

	// Create a real college, then a valid student update must succeed.
	collegeService := collegesvc.New()
	collegeID, err := collegeService.Create(ctx, &collegesvc.MutateInput{Name: "Agronomy", Sort: 1})
	if err != nil {
		t.Fatalf("create college failed: %v", err)
	}

	// Student with an existing college but missing grade is rejected.
	err = svc.UpdateProfile(ctx, player, &UpdateProfileInput{
		IdentityType: IdentityTypeStudent.String(),
		CollegeId:    collegeID,
		Grade:        0,
	})
	assertBizCode(t, err, CodeGradeInvalid.RuntimeCode())

	// Valid student update succeeds.
	if err = svc.UpdateProfile(ctx, player, &UpdateProfileInput{
		Nickname:     "stu",
		IdentityType: IdentityTypeStudent.String(),
		CollegeId:    collegeID,
		Grade:        3,
	}); err != nil {
		t.Fatalf("valid student profile update failed: %v", err)
	}

	profile, err := svc.GetProfile(ctx, player)
	if err != nil {
		t.Fatalf("read profile failed: %v", err)
	}
	if profile.IdentityType != IdentityTypeStudent.String() {
		t.Fatalf("expected identity type student, got %q", profile.IdentityType)
	}
	if profile.CollegeId != collegeID {
		t.Fatalf("expected college %d, got %d", collegeID, profile.CollegeId)
	}
	if profile.Grade != 3 {
		t.Fatalf("expected grade 3, got %d", profile.Grade)
	}
}
