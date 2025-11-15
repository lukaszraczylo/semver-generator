package utils

import (
	"testing"
	"time"

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
			want: SemVer{
				Major:                  0,
				Minor:                  0,
				Patch:                  1,
				Release:                1,
				EnableReleaseCandidate: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateSemver(
				tt.commits,
				tt.tags,
				tt.wording,
				tt.blacklist,
				tt.initialSemver,
				tt.respectExisting,
				tt.strictMode,
			)

			assert.Equal(t, tt.want.Major, got.Major, "Major version mismatch")
			assert.Equal(t, tt.want.Minor, got.Minor, "Minor version mismatch")
			assert.Equal(t, tt.want.Patch, got.Patch, "Patch version mismatch")
			assert.Equal(t, tt.want.Release, got.Release, "Release version mismatch")
			assert.Equal(t, tt.want.EnableReleaseCandidate, got.EnableReleaseCandidate, "EnableReleaseCandidate mismatch")
		})
	}
}
