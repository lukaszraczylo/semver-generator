package utils

import (
	"testing"
	"time"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/stretchr/testify/assert"
)

func TestCalculateSemver(t *testing.T) {
	// Initialize logger for tests
	InitLogger(false)

	// Mock the fuzzy find function for testing
	originalFuzzyFind := FuzzyFind
	defer func() { FuzzyFind = originalFuzzyFind }()

	FuzzyFind = func(needle string, haystack []string) []string {
		// More sophisticated mock implementation for testing
		for _, h := range haystack {
			// Check for substring match to better simulate fuzzy search
			if h == needle || (len(h) >= 3 && len(needle) >= 3 &&
				(h[:3] == needle[:3] || h[len(h)-3:] == needle[len(needle)-3:])) {
				return []string{h}
			}
		}
		return nil
	}

	// Test data
	now := time.Now()

	// Common wording and blacklist for all tests
	wording := Wording{
		Patch:   []string{"update", "fix", "initial"},
		Minor:   []string{"change", "feature", "improve"},
		Major:   []string{"breaking"},
		Release: []string{"rc", "release-candidate"},
	}

	blacklist := []string{"skip-ci", "no-version"}

	tests := []struct {
		name            string
		commits         []CommitDetails
		tags            []TagDetails
		wording         Wording
		blacklist       []string
		initialSemver   SemVer
		respectExisting bool
		strictMode      bool
		tagPrefixes     []string
		want            SemVer
	}{
		{
			name: "Standard mode with existing tags",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Update documentation",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			tags: []TagDetails{
				{
					Name: "2.0.0",
					Hash: "commit1",
				},
			},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: true,
			strictMode:      false,
			tagPrefixes:     []string{},
			want: SemVer{
				Major:                  2,
				Minor:                  0,
				Patch:                  1, // Initial tag 2.0.0 + one patch increment
				Release:                1,
				EnableReleaseCandidate: true,
			},
		},
		{
			name: "Strict mode with existing tags",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Update documentation",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			tags: []TagDetails{
				{
					Name: "2.0.0",
					Hash: "commit1",
				},
			},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: true,
			strictMode:      true,
			tagPrefixes:     []string{},
			want: SemVer{
				Major:                  2,
				Minor:                  0,
				Patch:                  1, // Initial tag 2.0.0 + patch from "update" keyword
				Release:                1,
				EnableReleaseCandidate: true,
			},
		},
		{
			name: "Standard mode without existing tags",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Update documentation",
					Timestamp: now.Add(-2 * time.Hour),
				},
				{
					Hash:      "commit3",
					Message:   "Change API interface",
					Timestamp: now.Add(-1 * time.Hour),
				},
			},
			tags:            []TagDetails{},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: false,
			strictMode:      false,
			tagPrefixes:     []string{},
			want: SemVer{
				Major: 0,
				Minor: 1,
				Patch: 1, // Minor increment resets patch to 1
			},
		},
		{
			name: "Strict mode without existing tags",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Update documentation",
					Timestamp: now.Add(-2 * time.Hour),
				},
				{
					Hash:      "commit3",
					Message:   "Change API interface",
					Timestamp: now.Add(-1 * time.Hour),
				},
			},
			tags:            []TagDetails{},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{Major: 1},
			respectExisting: false,
			strictMode:      true,
			tagPrefixes:     []string{},
			want: SemVer{
				Major: 1,
				Minor: 1,
				Patch: 1, // Minor increment resets patch to 1
			},
		},
		{
			name: "With blacklisted commits",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Update documentation skip-ci",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			tags:            []TagDetails{},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: false,
			strictMode:      false,
			tagPrefixes:     []string{},
			want: SemVer{
				Major: 0,
				Minor: 0,
				Patch: 3, // Default patch increment + patch from initial
			},
		},
		{
			name: "With release candidate",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "Initial commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "Add release-candidate",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			tags:            []TagDetails{},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: false,
			strictMode:      false,
			tagPrefixes:     []string{},
			want: SemVer{
				Major:                  0,
				Minor:                  0,
				Patch:                  1,
				Release:                1,
				EnableReleaseCandidate: true,
			},
		},
		{
			name: "With prefixed tags should not be RC",
			commits: []CommitDetails{
				{
					Hash:      "commit1",
					Message:   "tagged commit",
					Timestamp: now.Add(-3 * time.Hour),
				},
				{
					Hash:      "commit2",
					Message:   "another commit",
					Timestamp: now.Add(-2 * time.Hour),
				},
			},
			tags: []TagDetails{
				{
					Name: "app-0.0.16",
					Hash: "commit1",
				},
			},
			wording:         wording,
			blacklist:       blacklist,
			initialSemver:   SemVer{},
			respectExisting: true,
			strictMode:      true, // Use strict mode for predictable results
			tagPrefixes:     []string{"app-", "infra-"},
			want: SemVer{
				Major:                  0,
				Minor:                  0,
				Patch:                  16, // From tag, no additional increments in strict mode
				EnableReleaseCandidate: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSemver(
				tt.commits,
				tt.tags,
				tt.wording,
				FuzzyMatcher{},
				tt.blacklist,
				tt.initialSemver,
				tt.respectExisting,
				tt.strictMode,
				tt.tagPrefixes,
			)

			assert.Equal(t, tt.want.Major, got.Major, "Major version mismatch")
			assert.Equal(t, tt.want.Minor, got.Minor, "Minor version mismatch")
			assert.Equal(t, tt.want.Patch, got.Patch, "Patch version mismatch")
			assert.Equal(t, tt.want.Release, got.Release, "Release version mismatch")
			assert.Equal(t, tt.want.EnableReleaseCandidate, got.EnableReleaseCandidate, "EnableReleaseCandidate mismatch")
		})
	}
}

// TestCalculateSemverRegexModeEndToEnd exercises the whole config->matcher->
// CalculateSemver path in regex mode, with anchored conventional-commit
// patterns and a "Semver-Major:" trailer.
func TestCalculateSemverRegexModeEndToEnd(t *testing.T) {
	InitLogger(false)

	wording := Wording{
		Patch:   []string{`^fix:`},
		Minor:   []string{`(?m)^feat(\([^)]*\))?!?:`},
		Major:   []string{`(?mi)^semver-major:`},
		Release: []string{},
	}

	matcher, err := NewKeywordMatcher(MatchingRegex, wording)
	assert.NoError(t, err)

	// c4 ("fix(delegation): ...") is placed AFTER the major commit (c5) on
	// purpose: the anchored minor pattern `(?m)^feat(\([^)]*\))?!?:` must not
	// match a message that starts with "fix", so c4 contributes no version
	// bump at all under correct regex matching. If a future regression made
	// RegexMatcher match per-word instead of against the full raw message
	// (the false positive fuzzy mode has, see matcher_test.go's
	// "full-message anchoring" cases), c4 would wrongly bump Minor. With c4
	// last, nothing runs after it to reset that spurious bump back to 0, so
	// the final assertions below would catch it. (With c4 before c5, as it
	// used to be, the major commit's Minor=0 reset would mask the same bug.)
	now := time.Now()
	commits := []CommitDetails{
		{Hash: "c1", Message: "chore: init repo", Timestamp: now.Add(-5 * time.Hour)},
		{Hash: "c2", Message: "fix: correct off-by-one", Timestamp: now.Add(-4 * time.Hour)},
		{Hash: "c3", Message: "feat: add widget\n\nsome body text", Timestamp: now.Add(-3 * time.Hour)},
		{Hash: "c5", Message: "docs: update README\n\nSemver-Major: breaking storage format", Timestamp: now.Add(-2 * time.Hour)},
		{Hash: "c4", Message: "fix(delegation): route around dead worker", Timestamp: now.Add(-1 * time.Hour)},
	}

	got := CalculateSemver(
		commits,
		nil,
		wording,
		matcher,
		nil,
		SemVer{},
		false,
		true, // strict mode: only wording matches move the version
		nil,
	)

	assert.Equal(t, 1, got.Major, "Major version mismatch")
	assert.Equal(t, 0, got.Minor, "Minor version mismatch")
	assert.Equal(t, 1, got.Patch, "Patch version mismatch")
	assert.Equal(t, 0, got.Release, "Release version mismatch")
	assert.False(t, got.EnableReleaseCandidate)
}

// TestCalculateSemverRegexModeReleaseCandidateEndToEnd exercises
// wording.Release in regex mode end to end: a commit matching the release
// pattern must bump the release-candidate counter, exactly like fuzzy mode
// does for wording.Release matches.
func TestCalculateSemverRegexModeReleaseCandidateEndToEnd(t *testing.T) {
	InitLogger(false)

	wording := Wording{
		Patch:   []string{`^fix:`},
		Minor:   []string{`(?m)^feat(\([^)]*\))?!?:`},
		Major:   []string{`(?mi)^semver-major:`},
		Release: []string{`(?m)^release-candidate:`},
	}

	matcher, err := NewKeywordMatcher(MatchingRegex, wording)
	assert.NoError(t, err)

	now := time.Now()
	commits := []CommitDetails{
		{Hash: "c1", Message: "fix: correct off-by-one", Timestamp: now.Add(-2 * time.Hour)},
		{Hash: "c2", Message: "release-candidate: cut rc for QA", Timestamp: now.Add(-1 * time.Hour)},
	}

	got := CalculateSemver(
		commits,
		nil,
		wording,
		matcher,
		nil,
		SemVer{},
		false,
		true, // strict mode: only wording matches move the version
		nil,
	)

	assert.Equal(t, 0, got.Major, "Major version mismatch")
	assert.Equal(t, 0, got.Minor, "Minor version mismatch")
	assert.Equal(t, 1, got.Patch, "Patch version mismatch")
	assert.Equal(t, 1, got.Release, "Release version mismatch")
	assert.True(t, got.EnableReleaseCandidate)
}

// TestCalculateSemverFuzzyModeUnchanged runs CalculateSemver with the real
// fuzzy.FindNormalizedFold implementation (the same one main.go wires up) to
// confirm fuzzy mode's results are unchanged by the matcher refactor.
func TestCalculateSemverFuzzyModeUnchanged(t *testing.T) {
	InitLogger(false)

	originalFuzzyFind := FuzzyFind
	defer func() { FuzzyFind = originalFuzzyFind }()
	FuzzyFind = fuzzy.FindNormalizedFold

	wording := Wording{
		Patch: []string{"fix"},
		Minor: []string{"feat"},
		Major: []string{"breaking"},
	}
	blacklist := []string{"skip-ci"}

	now := time.Now()
	commits := []CommitDetails{
		{Hash: "c1", Message: "chore: init", Timestamp: now.Add(-6 * time.Hour)},
		{Hash: "c2", Message: "fix: correct bug", Timestamp: now.Add(-5 * time.Hour)},
		{Hash: "c3", Message: "feat: add widget", Timestamp: now.Add(-4 * time.Hour)},
		{Hash: "c4", Message: "docs: update readme", Timestamp: now.Add(-3 * time.Hour)},
		{Hash: "c5", Message: "breaking: change api contract", Timestamp: now.Add(-2 * time.Hour)},
		{Hash: "c6", Message: "fix: cleanup skip-ci", Timestamp: now.Add(-1 * time.Hour)},
	}

	matcher, err := NewKeywordMatcher(MatchingFuzzy, wording)
	assert.NoError(t, err)

	got := CalculateSemver(
		commits,
		nil,
		wording,
		matcher,
		blacklist,
		SemVer{},
		false,
		false, // non-strict: default patch increment applies too
		nil,
	)

	assert.Equal(t, 1, got.Major, "Major version mismatch")
	assert.Equal(t, 0, got.Minor, "Minor version mismatch")
	assert.Equal(t, 2, got.Patch, "Patch version mismatch")
	assert.Equal(t, 0, got.Release, "Release version mismatch")
	assert.False(t, got.EnableReleaseCandidate)
}
