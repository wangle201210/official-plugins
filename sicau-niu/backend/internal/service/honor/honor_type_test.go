// honor_type_test.go covers the pure honor enum validation and the in-memory
// unlock-rule evaluation. These tests use no database and always run.

package honor

import "testing"

// TestHonorTypeValid verifies the honor-type enum accepts only the allowed values.
func TestHonorTypeValid(t *testing.T) {
	valid := []HonorType{HonorTypeBadge, HonorTypeAvatarFrame, HonorTypeCertificate}
	for _, v := range valid {
		if !v.valid() {
			t.Fatalf("expected honor type %q to be valid", v)
		}
	}
	for _, v := range []HonorType{"", "medal", "frame"} {
		if v.valid() {
			t.Fatalf("expected honor type %q to be invalid", v)
		}
	}
}

// TestUnlockTypeValidAndShape verifies the unlock-rule enum validity and the
// threshold/category shape predicates.
func TestUnlockTypeValidAndShape(t *testing.T) {
	valid := []UnlockType{
		UnlockTypeParticipation,
		UnlockTypeFeedCount,
		UnlockTypeActivationCount,
		UnlockTypeCategoryComplete,
		UnlockTypeFullComplete,
	}
	for _, v := range valid {
		if !v.valid() {
			t.Fatalf("expected unlock type %q to be valid", v)
		}
	}
	for _, v := range []UnlockType{"", "feed", "complete"} {
		if v.valid() {
			t.Fatalf("expected unlock type %q to be invalid", v)
		}
	}

	if !UnlockTypeFeedCount.requiresThreshold() || !UnlockTypeActivationCount.requiresThreshold() {
		t.Fatalf("expected count rules to require a threshold")
	}
	if UnlockTypeParticipation.requiresThreshold() || UnlockTypeCategoryComplete.requiresThreshold() {
		t.Fatalf("expected non-count rules to not require a threshold")
	}
	if !UnlockTypeCategoryComplete.requiresCategory() {
		t.Fatalf("expected category_complete to require a category")
	}
	if UnlockTypeFeedCount.requiresCategory() || UnlockTypeFullComplete.requiresCategory() {
		t.Fatalf("expected non-category rules to not require a category")
	}
}

// TestPlayerProgressUnlocked verifies the in-memory unlock evaluation for every
// unlock rule against pre-aggregated counts.
func TestPlayerProgressUnlocked(t *testing.T) {
	progress := &playerProgress{
		feedCount:           10,
		activationCount:     3,
		collectedByCategory: map[string]int{"person": 2, "event": 1},
		collectedTotal:      3,
		activeByCategory:    map[string]int{"person": 2, "event": 3},
		activeTotal:         5,
	}

	cases := []struct {
		name       string
		unlockType UnlockType
		threshold  int
		category   string
		want       bool
	}{
		{name: "participation always unlocked", unlockType: UnlockTypeParticipation, want: true},
		{name: "feed_count met", unlockType: UnlockTypeFeedCount, threshold: 10, want: true},
		{name: "feed_count not met", unlockType: UnlockTypeFeedCount, threshold: 11, want: false},
		{name: "feed_count zero threshold never unlocks", unlockType: UnlockTypeFeedCount, threshold: 0, want: false},
		{name: "activation_count met", unlockType: UnlockTypeActivationCount, threshold: 3, want: true},
		{name: "activation_count not met", unlockType: UnlockTypeActivationCount, threshold: 4, want: false},
		{name: "category complete person (2/2)", unlockType: UnlockTypeCategoryComplete, category: "person", want: true},
		{name: "category incomplete event (1/3)", unlockType: UnlockTypeCategoryComplete, category: "event", want: false},
		{name: "category with no active cards never unlocks", unlockType: UnlockTypeCategoryComplete, category: "spirit", want: false},
		{name: "full complete not met (3/5)", unlockType: UnlockTypeFullComplete, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := progress.unlocked(tc.unlockType, tc.threshold, tc.category)
			if got != tc.want {
				t.Fatalf("unlocked(%q, %d, %q) = %v, want %v", tc.unlockType, tc.threshold, tc.category, got, tc.want)
			}
		})
	}
}

// TestPlayerProgressFullComplete verifies full_complete unlocks once every active
// card is collected and stays locked when no active card exists.
func TestPlayerProgressFullComplete(t *testing.T) {
	complete := &playerProgress{collectedTotal: 5, activeTotal: 5}
	if !complete.unlocked(UnlockTypeFullComplete, 0, "") {
		t.Fatalf("expected full_complete unlocked when 5/5 collected")
	}
	none := &playerProgress{collectedTotal: 0, activeTotal: 0}
	if none.unlocked(UnlockTypeFullComplete, 0, "") {
		t.Fatalf("expected full_complete locked when no active card exists")
	}
}
