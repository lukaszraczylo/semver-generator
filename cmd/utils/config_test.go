package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyForcedVersioning(t *testing.T) {
	tests := []struct {
		name   string
		force  Force
		semver SemVer
		want   SemVer
	}{
		{
			name: "No forced versioning",
			force: Force{
				Major: 0,
				Minor: 0,
				Patch: 0,
			},
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name: "Force major version",
			force: Force{
				Major: 5,
				Minor: 0,
				Patch: 0,
			},
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: SemVer{
				Major: 5,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name: "Force minor version",
			force: Force{
				Major: 0,
				Minor: 7,
				Patch: 0,
			},
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: SemVer{
				Major: 1,
				Minor: 7,
				Patch: 3,
			},
		},
		{
			name: "Force patch version",
			force: Force{
				Major: 0,
				Minor: 0,
				Patch: 9,
			},
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 9,
			},
		},
		{
			name: "Force all versions",
			force: Force{
				Major: 5,
				Minor: 7,
				Patch: 9,
			},
			semver: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
			want: SemVer{
				Major: 5,
				Minor: 7,
				Patch: 9,
			},
		},
	}

	// Initialize logger for tests
	InitLogger(false)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			semver := tt.semver
			ApplyForcedVersioning(tt.force, &semver)
			assert.Equal(t, tt.want.Major, semver.Major, "Major version mismatch")
			assert.Equal(t, tt.want.Minor, semver.Minor, "Minor version mismatch")
			assert.Equal(t, tt.want.Patch, semver.Patch, "Patch version mismatch")
		})
	}
}

func TestReadConfig(t *testing.T) {
	// Create a temporary config file for testing
	configContent := `
version: 1
force:
  major: 2
  minor: 3
  patch: 4
  commit: abcdef1234567890
  existing: true
  strict: false
blacklist:
  - "Merge branch"
  - "Merge pull request"
wording:
  patch:
    - update
    - fix
  minor:
    - change
    - feature
  major:
    - breaking
  release:
    - release-candidate
`
	tempFile, err := os.CreateTemp("", "semver-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Initialize logger for tests
	InitLogger(false)

	// Test reading the config
	config, err := ReadConfig(tempFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify force settings
	assert.Equal(t, 2, config.Force.Major)
	assert.Equal(t, 3, config.Force.Minor)
	assert.Equal(t, 4, config.Force.Patch)
	assert.Equal(t, "abcdef1234567890", config.Force.Commit)
	assert.True(t, config.Force.Existing)
	assert.False(t, config.Force.Strict)

	// Verify blacklist
	assert.Len(t, config.Blacklist, 2)
	assert.Contains(t, config.Blacklist, "Merge branch")
	assert.Contains(t, config.Blacklist, "Merge pull request")

	// Verify wording
	assert.Len(t, config.Wording.Patch, 2)
	assert.Contains(t, config.Wording.Patch, "update")
	assert.Contains(t, config.Wording.Patch, "fix")

	assert.Len(t, config.Wording.Minor, 2)
	assert.Contains(t, config.Wording.Minor, "change")
	assert.Contains(t, config.Wording.Minor, "feature")

	assert.Len(t, config.Wording.Major, 1)
	assert.Contains(t, config.Wording.Major, "breaking")

	assert.Len(t, config.Wording.Release, 1)
	assert.Contains(t, config.Wording.Release, "release-candidate")

	// Test reading a non-existent config
	_, err = ReadConfig("non-existent-file.yaml")
	assert.Error(t, err)
}
