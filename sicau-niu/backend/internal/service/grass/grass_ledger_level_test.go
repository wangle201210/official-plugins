// grass_ledger_level_test.go covers the arithmetic level curve that maps cumulative
// feeding experience onto a player level. These are pure-logic tests with no
// database dependency.

package grass

import "testing"

// TestLevelProgressBoundaries verifies the exact level boundaries of the curve:
// level L costs levelStep*(L-1), so reaching level L needs levelStep*(L-1)*L/2.
func TestLevelProgressBoundaries(t *testing.T) {
	cases := []struct {
		exp          int64
		level        int
		intoLevel    int64
		forNextLevel int64
	}{
		{exp: 0, level: 1, intoLevel: 0, forNextLevel: 100},
		{exp: 99, level: 1, intoLevel: 99, forNextLevel: 100},
		{exp: 100, level: 2, intoLevel: 0, forNextLevel: 200},
		{exp: 299, level: 2, intoLevel: 199, forNextLevel: 200},
		{exp: 300, level: 3, intoLevel: 0, forNextLevel: 300},
		{exp: 599, level: 3, intoLevel: 299, forNextLevel: 300},
		{exp: 600, level: 4, intoLevel: 0, forNextLevel: 400},
		{exp: 1000, level: 5, intoLevel: 0, forNextLevel: 500},
		{exp: 1500, level: 6, intoLevel: 0, forNextLevel: 600},
		{exp: 2100, level: 7, intoLevel: 0, forNextLevel: 700},
	}
	for _, c := range cases {
		level, intoLevel, forNextLevel := levelProgress(c.exp)
		if level != c.level || intoLevel != c.intoLevel || forNextLevel != c.forNextLevel {
			t.Fatalf("exp=%d: got level=%d into=%d next=%d, want level=%d into=%d next=%d",
				c.exp, level, intoLevel, forNextLevel, c.level, c.intoLevel, c.forNextLevel)
		}
	}
}

// TestLevelProgressCostsGrowByOneStep verifies each level costs exactly one more
// step than the previous one, which is what distinguishes this curve from the
// previous flat 100-per-level rule.
func TestLevelProgressCostsGrowByOneStep(t *testing.T) {
	var cumulative int64
	for level := 1; level <= 12; level++ {
		got, intoLevel, forNextLevel := levelProgress(cumulative)
		if got != level || intoLevel != 0 {
			t.Fatalf("cumulative=%d: expected the exact start of level %d, got level=%d into=%d", cumulative, level, got, intoLevel)
		}
		if want := int64(level) * levelStep; forNextLevel != want {
			t.Fatalf("level %d advances at %d experience, want %d", level, forNextLevel, want)
		}
		cumulative += forNextLevel
	}
}

// TestLevelProgressIsMonotonic verifies experience never lowers the level and the
// in-level progress always stays below the advance threshold.
func TestLevelProgressIsMonotonic(t *testing.T) {
	previous := 1
	for exp := int64(0); exp <= 5000; exp += 7 {
		level, intoLevel, forNextLevel := levelProgress(exp)
		if level < previous {
			t.Fatalf("exp=%d lowered the level from %d to %d", exp, previous, level)
		}
		if intoLevel < 0 || intoLevel >= forNextLevel {
			t.Fatalf("exp=%d: in-level progress %d is outside [0, %d)", exp, intoLevel, forNextLevel)
		}
		previous = level
	}
}

// TestLevelProgressClampsNegativeExperience verifies a negative aggregate, which a
// database sum could in principle produce, is treated as a fresh level 1 player.
func TestLevelProgressClampsNegativeExperience(t *testing.T) {
	level, intoLevel, forNextLevel := levelProgress(-500)
	if level != 1 || intoLevel != 0 || forNextLevel != levelStep {
		t.Fatalf("negative experience produced level=%d into=%d next=%d", level, intoLevel, forNextLevel)
	}
}
