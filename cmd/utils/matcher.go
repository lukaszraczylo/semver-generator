// Package utils implements semver-gen's configuration loading, git access,
// commit-keyword matching (fuzzy and regex) and version calculation.
package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Matching mode values accepted by the "matching" config key.
const (
	MatchingFuzzy = "fuzzy"
	MatchingRegex = "regex"
)

// ErrInvalidMatchingConfig is wrapped by any error NewKeywordMatcher returns.
// Callers use errors.Is to tell a bad "matching" value or a bad regex keyword
// apart from other, non-fatal config problems.
var ErrInvalidMatchingConfig = errors.New("invalid matching configuration")

// KeywordMatcher decides whether a commit message matches one of a wording
// list's keywords, honoring the blacklist. CalculateSemver picks one
// implementation, chosen once from config, instead of branching on the
// matching mode itself.
type KeywordMatcher interface {
	Matches(commitMessage string, keywords []string, blacklist []string) bool
}

// FuzzyMatcher reproduces the original matching behaviour: the commit message
// is split into whitespace-separated words and each keyword is fuzzy-matched
// (letters in order, case-insensitive) against those words via FuzzyFind.
type FuzzyMatcher struct{}

// Matches implements KeywordMatcher.
func (FuzzyMatcher) Matches(commitMessage string, keywords []string, blacklist []string) bool {
	return CheckMatches(strings.Fields(commitMessage), keywords, blacklist)
}

// RegexMatcher matches keywords, each a Go RE2 regular expression compiled
// once up front, against the full raw commit message (subject and body,
// newlines preserved).
type RegexMatcher struct {
	compiled map[string]*regexp.Regexp
}

// Matches implements KeywordMatcher.
func (m RegexMatcher) Matches(commitMessage string, keywords []string, blacklist []string) bool {
	hasMatch := false
	for _, keyword := range keywords {
		re, ok := m.compiled[keyword]
		if !ok {
			// NewKeywordMatcher compiles every keyword of every wording list up
			// front, so this means Matches was called with a keyword that
			// wasn't part of the wording the matcher was built from. Log it
			// loudly instead of silently skipping, and treat it as no-match
			// rather than panicking on a nil regexp.
			Error("Regex keyword was never compiled; treating as no-match", map[string]interface{}{
				"keyword": keyword,
			})
			continue
		}
		if re.MatchString(commitMessage) {
			hasMatch = true
			break
		}
	}

	if !hasMatch {
		return false
	}

	// Match blacklist semantics exactly the same way CheckMatches (fuzzy mode)
	// does: whitespace-normalized content, so a blacklist term split across a
	// newline or tab in the raw commit message still suppresses the match.
	contentStr := strings.Join(strings.Fields(commitMessage), " ")
	return !isBlacklisted(contentStr, blacklist)
}

// NewKeywordMatcher builds the KeywordMatcher selected by matching ("" and
// "fuzzy" both mean fuzzy, "regex" means regex; anything else is an error).
// In regex mode every keyword across all four wording lists is compiled
// exactly once here, so a bad pattern is caught at config-load time instead
// of failing mid-run. Compile and validation errors wrap
// ErrInvalidMatchingConfig so callers can fail loudly on them.
func NewKeywordMatcher(matching string, wording Wording) (KeywordMatcher, error) {
	switch matching {
	case "", MatchingFuzzy:
		return FuzzyMatcher{}, nil
	case MatchingRegex:
		wordingLists := []struct {
			name     string
			keywords []string
		}{
			{"patch", wording.Patch},
			{"minor", wording.Minor},
			{"major", wording.Major},
			{"release", wording.Release},
		}

		compiled := make(map[string]*regexp.Regexp)
		for _, list := range wordingLists {
			for _, keyword := range list.keywords {
				if _, alreadyCompiled := compiled[keyword]; alreadyCompiled {
					continue
				}
				re, err := regexp.Compile(keyword)
				if err != nil {
					return nil, fmt.Errorf("%w: invalid regex in wording.%s keyword %q: %v", ErrInvalidMatchingConfig, list.name, keyword, err)
				}
				compiled[keyword] = re
			}
		}
		return RegexMatcher{compiled: compiled}, nil
	default:
		return nil, fmt.Errorf("%w: invalid matching value %q, allowed values are %q and %q", ErrInvalidMatchingConfig, matching, MatchingFuzzy, MatchingRegex)
	}
}
