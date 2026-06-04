// activation_fake_test.go defines the in-test stub identity service used by the
// poster tests. It implements the identity Service contract with a fixed profile
// so the poster composition path can be exercised without the real C1 service
// graph. Tests that need a specific nickname/identity set the profile fields
// before calling.

package activation

import (
	"context"

	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// fakeIdentityService is a stub identity Service returning a configurable profile.
type fakeIdentityService struct {
	// profile is the profile returned by GetProfile; nil yields a zero profile.
	profile *identitysvc.ProfileOutput
}

// Login is unused by the activation tests and panics if called.
func (f *fakeIdentityService) Login(ctx context.Context, in *identitysvc.LoginInput) (*identitysvc.LoginOutput, error) {
	panic("fakeIdentityService.Login is not used by activation tests")
}

// BindPhone is unused by the activation tests and panics if called.
func (f *fakeIdentityService) BindPhone(ctx context.Context, playerID int64, in *identitysvc.BindPhoneInput) error {
	panic("fakeIdentityService.BindPhone is not used by activation tests")
}

// GetProfile returns the configured profile or a zero profile carrying playerID.
func (f *fakeIdentityService) GetProfile(ctx context.Context, playerID int64) (*identitysvc.ProfileOutput, error) {
	if f.profile != nil {
		return f.profile, nil
	}
	return &identitysvc.ProfileOutput{Id: playerID}, nil
}

// UpdateProfile is unused by the activation tests and panics if called.
func (f *fakeIdentityService) UpdateProfile(ctx context.Context, playerID int64, in *identitysvc.UpdateProfileInput) error {
	panic("fakeIdentityService.UpdateProfile is not used by activation tests")
}

// ListPlayers is unused by the activation tests and panics if called.
func (f *fakeIdentityService) ListPlayers(ctx context.Context, in *identitysvc.ListPlayersInput) (*identitysvc.ListPlayersOutput, error) {
	panic("fakeIdentityService.ListPlayers is not used by activation tests")
}
