package utils

import (
	"testing"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeywordMatcher(t *testing.T) {
	InitLogger(false)

	wording := Wording{
		Patch: []string{"^fix:"},
		Minor: []string{`(?m)^feat(\([^)]*\))?!?:`},
		Major: []string{"(?mi)^semver-major:"},
	}

	t.Run("empty matching defaults to fuzzy", func(t *testing.T) {
		m, err := NewKeywordMatcher("", wording)
		assert.NoError(t, err)
		assert.IsType(t, FuzzyMatcher{}, m)
		// Behavioural check, not just the type: fuzzy matching means
		// "fix(delegation): x" still matches keyword "feat" (f-e-a-t appear
		// in order inside "fix(delegation):"), using the real fuzzy library
		// (the same one main.go wires up), not a stand-in.
		originalFuzzyFind := FuzzyFind
		t.Cleanup(func() { FuzzyFind = originalFuzzyFind })
		FuzzyFind = fuzzy.FindNormalizedFold
		assert.True(t, m.Matches("fix(delegation): x", []string{"feat"}, nil))
	})

	t.Run("fuzzy", func(t *testing.T) {
		m, err := NewKeywordMatcher(MatchingFuzzy, wording)
		assert.NoError(t, err)
		assert.IsType(t, FuzzyMatcher{}, m)

		// Behavioural check, not just the type: same as the empty-matching
		// subtest above, using the real fuzzy library.
		originalFuzzyFind := FuzzyFind
		t.Cleanup(func() { FuzzyFind = originalFuzzyFind })
		FuzzyFind = fuzzy.FindNormalizedFold
		assert.True(t, m.Matches("fix(delegation): x", []string{"feat"}, nil))
	})

	t.Run("regex compiles every keyword once", func(t *testing.T) {
		m, err := NewKeywordMatcher(MatchingRegex, wording)
		assert.NoError(t, err)
		assert.IsType(t, RegexMatcher{}, m)
		// Behavioural check: the same input that fuzzy-matches "feat" does not
		// match the anchored regex pattern in wording.Minor.
		assert.False(t, m.Matches("fix(delegation): x", wording.Minor, nil))
	})

	t.Run("unknown value errors naming value and allowed values", func(t *testing.T) {
		_, err := NewKeywordMatcher("weird", wording)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "weird")
		assert.Contains(t, err.Error(), MatchingFuzzy)
		assert.Contains(t, err.Error(), MatchingRegex)
	})

	t.Run("invalid regex errors naming wording list and keyword", func(t *testing.T) {
		badWording := Wording{Major: []string{"[unterminated("}}
		_, err := NewKeywordMatcher(MatchingRegex, badWording)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "major")
		assert.Contains(t, err.Error(), "[unterminated(")
	})

	t.Run("invalid release regex errors naming wording list and keyword", func(t *testing.T) {
		badWording := Wording{Release: []string{"[unterminated("}}
		_, err := NewKeywordMatcher(MatchingRegex, badWording)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidMatchingConfig)
		assert.Contains(t, err.Error(), "release")
		assert.Contains(t, err.Error(), "[unterminated(")
	})
}

func TestRegexMatcherMatches(t *testing.T) {
	InitLogger(false)

	t.Run("anchored feat subject matches", func(t *testing.T) {
		m, err := NewKeywordMatcher(MatchingRegex, Wording{
			Minor: []string{`(?m)^feat(\([^)]*\))?!?:`},
		})
		assert.NoError(t, err)
		got := m.Matches("feat: add widget\n\nsome body text", []string{`(?m)^feat(\([^)]*\))?!?:`}, nil)
		assert.True(t, got)
	})

	t.Run("fix(delegation) subject does not match anchored feat pattern", func(t *testing.T) {
		m, err := NewKeywordMatcher(MatchingRegex, Wording{
			Minor: []string{`(?m)^feat(\([^)]*\))?!?:`},
		})
		assert.NoError(t, err)
		got := m.Matches("fix(delegation): route around dead worker", []string{`(?m)^feat(\([^)]*\))?!?:`}, nil)
		assert.False(t, got)
	})

	t.Run("body trailer Semver-Major matches multiline case-insensitive pattern", func(t *testing.T) {
		pattern := "(?mi)^semver-major:"
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Major: []string{pattern}})
		assert.NoError(t, err)
		message := "fix: patch a bug\n\nDetails here.\n\nSemver-Major: breaking storage format\n"
		got := m.Matches(message, []string{pattern}, nil)
		assert.True(t, got)
	})

	t.Run("feat! subject does not match a trailer-only major pattern", func(t *testing.T) {
		pattern := "(?mi)^semver-major:"
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Major: []string{pattern}})
		assert.NoError(t, err)
		got := m.Matches("feat!: rework the API\n\nno trailer here", []string{pattern}, nil)
		assert.False(t, got)
	})

	t.Run("case sensitivity: pattern without (?i) does not match different case", func(t *testing.T) {
		pattern := "^FEAT:"
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{pattern}})
		assert.NoError(t, err)
		assert.False(t, m.Matches("feat: lowercase subject", []string{pattern}, nil))
		assert.True(t, m.Matches("FEAT: uppercase subject", []string{pattern}, nil))
	})

	t.Run("blacklist still suppresses a regex match", func(t *testing.T) {
		pattern := `(?m)^feat(\([^)]*\))?!?:`
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{pattern}})
		assert.NoError(t, err)
		message := "feat: add widget [skip-ci]"
		assert.True(t, m.Matches(message, []string{pattern}, nil))
		assert.False(t, m.Matches(message, []string{pattern}, []string{"skip-ci"}))
	})

	t.Run("blacklist term split by a newline still suppresses, same as fuzzy mode", func(t *testing.T) {
		pattern := `(?m)^feat(\([^)]*\))?!?:`
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{pattern}})
		assert.NoError(t, err)
		// Blacklist term is "skip ci", but the raw commit message has it split
		// across a newline: "skip\nci". CheckMatches (fuzzy mode) would still
		// catch this because it joins strings.Fields(msg) with a single space
		// before the substring check; the regex-mode blacklist check must do
		// the same.
		message := "feat: add widget\n\nskip\nci"
		assert.True(t, m.Matches(message, []string{pattern}, nil), "sanity: matches without blacklist")
		assert.False(t, m.Matches(message, []string{pattern}, []string{"skip ci"}))
	})

	t.Run("blacklist term split by a tab still suppresses, same as fuzzy mode", func(t *testing.T) {
		pattern := `(?m)^feat(\([^)]*\))?!?:`
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{pattern}})
		assert.NoError(t, err)
		message := "feat: add widget\n\nskip\tci"
		assert.True(t, m.Matches(message, []string{pattern}, nil), "sanity: matches without blacklist")
		assert.False(t, m.Matches(message, []string{pattern}, []string{"skip ci"}))
	})

	t.Run("full-message anchoring: mid-message keyword-looking text does not match anchored pattern", func(t *testing.T) {
		// Discriminating case: if RegexMatcher ran MatchString per
		// strings.Fields word instead of against the full raw message, the
		// word "feat:" here would match "(?m)^feat...:" on its own (it starts
		// a "word"), even though the actual commit subject starts with
		// "docs:". Matching against the full message must not match.
		pattern := `(?m)^feat(\([^)]*\))?!?:`
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{pattern}})
		assert.NoError(t, err)
		got := m.Matches("docs: explain feat: usage", []string{pattern}, nil)
		assert.False(t, got)
	})

	t.Run("full-message anchoring: multiline trailer matches (?m) anchor on its own line", func(t *testing.T) {
		// Discriminating case: strings.Fields on this message splits "BREAKING"
		// and "CHANGE:" into separate words, neither of which starts with
		// "BREAKING CHANGE:", so a per-word implementation would miss this.
		// Matching against the full raw message (newlines kept) must match,
		// because (?m) anchors "^" to the start of the trailer's own line.
		pattern := `(?m)^BREAKING CHANGE:`
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Major: []string{pattern}})
		assert.NoError(t, err)
		got := m.Matches("refactor: x\n\nBREAKING CHANGE: drop v1", []string{pattern}, nil)
		assert.True(t, got)
	})

	t.Run("keyword missing from compiled map is a no-match, not a panic", func(t *testing.T) {
		// The matcher was built from wording.Minor only; calling Matches with
		// an unrelated keyword exercises the "not compiled" branch. It must
		// log loudly (via the package's Error logger) rather than silently
		// return, but it must not miss the check or panic on a nil regexp.
		m, err := NewKeywordMatcher(MatchingRegex, Wording{Minor: []string{`(?m)^feat:`}})
		assert.NoError(t, err)
		assert.NotPanics(t, func() {
			got := m.Matches("feat: add widget", []string{"^never-compiled:"}, nil)
			assert.False(t, got)
		})
	})
}

func TestFuzzyMatcherMatches(t *testing.T) {
	InitLogger(false)

	originalFuzzyFind := FuzzyFind
	defer func() { FuzzyFind = originalFuzzyFind }()
	FuzzyFind = func(needle string, haystack []string) []string {
		for _, h := range haystack {
			if h == needle {
				return []string{h}
			}
		}
		return nil
	}

	m := FuzzyMatcher{}
	assert.True(t, m.Matches("update dependencies", []string{"update", "fix"}, nil))
	assert.False(t, m.Matches("chore dependencies", []string{"update", "fix"}, nil))
	assert.False(t, m.Matches("update dependencies skip-ci", []string{"update", "fix"}, []string{"skip-ci"}))
}
