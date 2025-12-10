package utils

import (
	"regexp"
	"strconv"
	"strings"
)

// SemVer represents a semantic version
type SemVer struct {
	Patch                  int
	Minor                  int
	Major                  int
	Release                int
	EnableReleaseCandidate bool
}

// FormatSemver formats a semantic version as a string
func FormatSemver(semver SemVer) string {
	result := strings.TrimSpace(
		strings.Join(
			[]string{
				strconv.Itoa(semver.Major),
				strconv.Itoa(semver.Minor),
				strconv.Itoa(semver.Patch),
			},
			".",
		),
	)

	if semver.EnableReleaseCandidate {
		result = strings.TrimSpace(
			strings.Join(
				[]string{
					result,
					strings.Join(
						[]string{
							"rc",
							strconv.Itoa(semver.Release),
						},
						".",
					),
				},
				"-",
			),
		)
	}

	return result
}

var extractNumber = regexp.MustCompile("[0-9]+")

// StripTagPrefix removes configured prefixes from a tag name
// The "v" prefix is always stripped by default (e.g., v1.2.3 -> 1.2.3)
func StripTagPrefix(tagName string, prefixes []string) string {
	result := tagName

	// Always strip "v" prefix by default
	if strings.HasPrefix(result, "v") && len(result) > 1 {
		// Only strip if followed by a digit (to avoid stripping "version-1.0.0")
		if result[1] >= '0' && result[1] <= '9' {
			result = result[1:]
			Debug("Stripped default 'v' prefix from tag", map[string]interface{}{
				"original": tagName,
				"result":   result,
			})
		}
	}

	// Then strip any user-configured prefixes
	for _, prefix := range prefixes {
		if strings.HasPrefix(result, prefix) {
			result = strings.TrimPrefix(result, prefix)
			Debug("Stripped prefix from tag", map[string]interface{}{
				"original": tagName,
				"prefix":   prefix,
				"result":   result,
			})
			break // Only strip one prefix
		}
	}
	return result
}

// ParseExistingSemver parses a semantic version from a tag name
func ParseExistingSemver(tagName string, currentSemver SemVer, prefixes []string) SemVer {
	Debug("Parsing existing semver", map[string]interface{}{"tag": tagName})

	// Strip configured prefixes before parsing
	cleanTagName := StripTagPrefix(tagName, prefixes)

	// Check for release candidate pattern (-rc.X) before splitting
	isReleaseCandidate := false
	rcVersion := 0
	if idx := strings.Index(cleanTagName, "-rc."); idx != -1 {
		isReleaseCandidate = true
		rcPart := cleanTagName[idx+4:] // Get everything after "-rc."
		rcMatches := extractNumber.FindAllString(rcPart, 1)
		if len(rcMatches) > 0 {
			rcVersion, _ = strconv.Atoi(rcMatches[0])
		}
		// Remove the RC suffix for version parsing
		cleanTagName = cleanTagName[:idx]
		Debug("Detected release candidate", map[string]interface{}{
			"rc_version":     rcVersion,
			"clean_tag_name": cleanTagName,
		})
	}

	tagNameParts := strings.Split(cleanTagName, ".")
	if len(tagNameParts) < 3 {
		Debug("Unable to parse incompatible semver (non x.y.z)", map[string]interface{}{"tag": tagName})
		return currentSemver
	}

	semanticVersion := SemVer{}

	// Extract major version
	majorMatches := extractNumber.FindAllString(tagNameParts[0], -1)
	if len(majorMatches) > 0 {
		semanticVersion.Major, _ = strconv.Atoi(majorMatches[0])
	}

	// Extract minor version
	minorMatches := extractNumber.FindAllString(tagNameParts[1], -1)
	if len(minorMatches) > 0 {
		semanticVersion.Minor, _ = strconv.Atoi(minorMatches[0])
	}

	// Extract patch version
	patchMatches := extractNumber.FindAllString(tagNameParts[2], -1)
	if len(patchMatches) > 0 {
		semanticVersion.Patch, _ = strconv.Atoi(patchMatches[0])
	}

	// Set release candidate if detected
	if isReleaseCandidate {
		semanticVersion.Release = rcVersion
		semanticVersion.EnableReleaseCandidate = true
	}

	return semanticVersion
}

// CheckMatches checks if any of the targets match the content
func CheckMatches(content []string, targets []string, blacklist []string) bool {
	contentStr := strings.Join(content, " ")

	// First check if any target matches
	hasMatch := false
	for _, tgt := range targets {
		matches := FuzzyFind(tgt, content)
		if len(matches) > 0 {
			hasMatch = true
			Debug("Found match", map[string]interface{}{
				"target":  tgt,
				"match":   strings.Join(matches, ","),
				"content": contentStr,
			})
			break
		}
	}

	// If we have a match, check against blacklist
	if hasMatch && len(blacklist) > 0 {
		for _, blacklistTerm := range blacklist {
			if strings.Contains(strings.ToLower(contentStr), strings.ToLower(blacklistTerm)) {
				Debug("Blacklisted term detected, ignoring commit", map[string]interface{}{
					"content":        contentStr,
					"blacklist_term": blacklistTerm,
				})
				return false
			}
		}
	}

	return hasMatch
}

// FuzzyFind is a wrapper for the fuzzy search library to make it easier to mock in tests
var FuzzyFind = func(needle string, haystack []string) []string {
	// This will be replaced with the actual implementation in main.go
	return nil
}
