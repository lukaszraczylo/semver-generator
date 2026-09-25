package utils

import (
	"errors"
	"os"
	"testing"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	tempFile, err := os.CreateTemp("", "semver-config-matching-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(tempFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("failed to remove temp config %s: %v", tempFile.Name(), err)
		}
	})
	return tempFile.Name()
}

func TestReadConfigMatching(t *testing.T) {
	InitLogger(false)

	t.Run("absent matching key defaults to fuzzy", func(t *testing.T) {
		path := writeTempConfig(t, `
wording:
  patch:
    - feat
`)
		config, err := ReadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, MatchingFuzzy, config.Matching)
		assert.IsType(t, FuzzyMatcher{}, config.Matcher)

		// Behavioural check, not just the type: absent/fuzzy mode means
		// "fix(delegation): x" fuzzy-matches keyword "feat", using the real
		// fuzzy library (the same one main.go wires up), not a stand-in.
		originalFuzzyFind := FuzzyFind
		t.Cleanup(func() { FuzzyFind = originalFuzzyFind })
		FuzzyFind = fuzzy.FindNormalizedFold
		assert.True(t, config.Matcher.Matches("fix(delegation): x", config.Wording.Patch, nil))
	})

	t.Run("explicit fuzzy", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: fuzzy
wording:
  patch:
    - update
`)
		config, err := ReadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, MatchingFuzzy, config.Matching)
		assert.IsType(t, FuzzyMatcher{}, config.Matcher)

		// Behavioural check, not just the type: same as the absent-key
		// subtest above, using the real fuzzy library.
		originalFuzzyFind := FuzzyFind
		t.Cleanup(func() { FuzzyFind = originalFuzzyFind })
		FuzzyFind = fuzzy.FindNormalizedFold
		assert.True(t, config.Matcher.Matches("fix(delegation): x", []string{"feat"}, nil))
	})

	t.Run("explicit regex compiles keywords", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: regex
wording:
  patch:
    - "^fix:"
  minor:
    - "(?m)^feat(\\([^)]*\\))?!?:"
  major:
    - "(?mi)^semver-major:"
`)
		config, err := ReadConfig(path)
		assert.NoError(t, err)
		assert.Equal(t, MatchingRegex, config.Matching)
		assert.IsType(t, RegexMatcher{}, config.Matcher)
		// Behavioural check: "fix(delegation): x" matches fuzzy keyword "feat"
		// (see the absent/fuzzy subtest below) but must NOT match this
		// anchored regex pattern.
		assert.False(t, config.Matcher.Matches("fix(delegation): x", config.Wording.Minor, nil))
	})

	t.Run("unknown matching value errors naming value and allowed values", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: substring
wording:
  patch:
    - update
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "substring")
		assert.Contains(t, err.Error(), MatchingFuzzy)
		assert.Contains(t, err.Error(), MatchingRegex)
	})

	// Fix 1: an invalid "matching" value must be rejected immediately, before
	// wording/force/blacklist/tag_prefixes are ever parsed. Previously a
	// mistyped value like "substring" was accepted as-is by the case string
	// branch, so a malformed wording section below it returned an unwrapped
	// parse error instead of ErrInvalidMatchingConfig, and main() silently
	// fell back to fuzzy mode instead of failing loudly.
	t.Run("invalid matching value combined with malformed wording section still errors as ErrInvalidMatchingConfig", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: substring
wording: "not-a-map"
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "substring")
	})

	t.Run("invalid regex errors naming wording list and keyword", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: regex
wording:
  major:
    - "[unterminated("
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "major")
		assert.Contains(t, err.Error(), "[unterminated(")
	})

	// Fix 2: no silent fuzzy fallback in regex mode. A malformed "wording"
	// section (a scalar instead of a map) fails viper.UnmarshalKey, not
	// NewKeywordMatcher. In regex mode that must still be a fatal,
	// ErrInvalidMatchingConfig-wrapped error — main() exits non-zero on it
	// (see cmd/main.go's errors.Is(err, utils.ErrInvalidMatchingConfig)
	// check) instead of silently logging and falling back to fuzzy matching.
	t.Run("regex mode + malformed wording section is a fatal error, not a fuzzy fallback", func(t *testing.T) {
		path := writeTempConfig(t, `
matching: regex
wording: "not-a-map"
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig, "main() only treats ErrInvalidMatchingConfig as fatal; anything else falls back to fuzzy and continues")
		assert.Contains(t, err.Error(), "wording")
	})

	// Contrast case: the exact same malformed "wording" section in fuzzy mode
	// (the default) must behave exactly as it did before regex mode existed —
	// an error is returned, but it is NOT wrapped in ErrInvalidMatchingConfig,
	// so main() logs it and continues with defaults rather than exiting.
	t.Run("fuzzy mode + malformed wording section stays non-fatal, same as today", func(t *testing.T) {
		path := writeTempConfig(t, `
wording: "not-a-map"
`)
		_, err := ReadConfig(path)
		assert.Error(t, err)
		assert.False(t, errors.Is(err, ErrInvalidMatchingConfig), "fuzzy mode must not fail loudly for non-matching config errors")
	})

	// Fix 4: a non-scalar "matching" value (a YAML list or map) must be an
	// ErrInvalidMatchingConfig error naming the value, never a silent fuzzy
	// fallback.
	t.Run("matching as a YAML list is an error naming the value", func(t *testing.T) {
		path := writeTempConfig(t, `
matching:
  - fuzzy
  - regex
wording:
  patch:
    - update
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "fuzzy")
		assert.Contains(t, err.Error(), "regex")
	})

	t.Run("matching as a YAML map is an error naming the value", func(t *testing.T) {
		path := writeTempConfig(t, `
matching:
  mode: regex
wording:
  patch:
    - update
`)
		_, err := ReadConfig(path)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "mode")
	})
}
