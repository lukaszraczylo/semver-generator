package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatSemver(t *testing.T) {
	tests := []struct {
		name   string
		semver SemVer
		want   string
	}{
		{
			name: "Basic version",
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: "1.2.3",
		},
		{
			name: "With release candidate",
			semver: SemVer{
				Major:                  2,
				Minor:                  0,
				Patch:                  1,
				Release:                5,
				EnableReleaseCandidate: true,
			},
			want: "2.0.1-rc.5",
		},
		{
			name: "With release candidate disabled",
			semver: SemVer{
				Major:                  3,
				Minor:                  1,
				Patch:                  0,
				Release:                2,
				EnableReleaseCandidate: false,
			},
			want: "3.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatSemver(tt.semver)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseExistingSemver(t *testing.T) {
	// Initialize logger for tests
	InitLogger(false)

	tests := []struct {
		name         string
		tagName      string
		currentSemver SemVer
		want         SemVer
	}{
		{
			name:    "Standard semver",
			tagName: "1.2.3",
			currentSemver: SemVer{},
			want: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name:    "With v prefix",
			tagName: "v2.3.4",
			currentSemver: SemVer{},
			want: SemVer{
				Major: 2,
				Minor: 3,
				Patch: 4,
			},
		},
		{
			name:    "With release candidate",
			tagName: "3.4.5-rc.2",
			currentSemver: SemVer{},
			want: SemVer{
				Major:                  3,
				Minor:                  4,
				Patch:                  5,
				Release:                2,
				EnableReleaseCandidate: true,
			},
		},
		{
			name:    "Invalid format",
			tagName: "not-a-semver",
			currentSemver: SemVer{
				Major: 1,
				Minor: 1,
				Patch: 1,
			},
			want: SemVer{
				Major: 1,
				Minor: 1,
				Patch: 1,
			},
		},
		{
			name:    "Incomplete format",
			tagName: "1.2",
			currentSemver: SemVer{
				Major: 5,
				Minor: 5,
				Patch: 5,
			},
			want: SemVer{
				Major: 5,
				Minor: 5,
				Patch: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseExistingSemver(tt.tagName, tt.currentSemver)
			assert.Equal(t, tt.want.Major, got.Major, "Major version mismatch")
			assert.Equal(t, tt.want.Minor, got.Minor, "Minor version mismatch")
			assert.Equal(t, tt.want.Patch, got.Patch, "Patch version mismatch")
			assert.Equal(t, tt.want.Release, got.Release, "Release version mismatch")
			assert.Equal(t, tt.want.EnableReleaseCandidate, got.EnableReleaseCandidate, "EnableReleaseCandidate mismatch")
		})
	}
}

func TestCheckMatches(t *testing.T) {
	// Initialize logger for tests
	InitLogger(false)

	// Mock the fuzzy find function for testing
	originalFuzzyFind := FuzzyFind
	defer func() { FuzzyFind = originalFuzzyFind }()

	FuzzyFind = func(needle string, haystack []string) []string {
		// Simple mock implementation for testing
		for _, h := range haystack {
			if h == needle {
				return []string{h}
			}
		}
		return nil
	}

	tests := []struct {
		name      string
		content   []string
		targets   []string
		blacklist []string
		want      bool
	}{
		{
			name:    "Simple match",
			content: []string{"update", "dependencies"},
			targets: []string{"update", "fix"},
			want:    true,
		},
		{
			name:    "No match",
			content: []string{"chore", "dependencies"},
			targets: []string{"update", "fix"},
			want:    false,
		},
		{
			name:      "Match but blacklisted",
			content:   []string{"update", "dependencies", "skip-ci"},
			targets:   []string{"update", "fix"},
			blacklist: []string{"skip-ci"},
			want:      false,
		},
		{
			name:      "Match with empty blacklist",
			content:   []string{"update", "dependencies"},
			targets:   []string{"update", "fix"},
			blacklist: []string{},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckMatches(tt.content, tt.targets, tt.blacklist)
			assert.Equal(t, tt.want, got)
		})
	}
}