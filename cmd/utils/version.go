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

// ParseExistingSemver parses a semantic version from a tag name
func ParseExistingSemver(tagName string, currentSemver SemVer) SemVer {
	Debug("Parsing existing semver", map[string]interface{}{"tag": tagName})
	
	tagNameParts := strings.Split(tagName, ".")
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
	
	// Extract release candidate version if present
	if len(tagNameParts) > 3 {
		releaseMatches := extractNumber.FindAllString(tagNameParts[3], -1)
		if len(releaseMatches) > 0 {
			semanticVersion.Release, _ = strconv.Atoi(releaseMatches[0])
			semanticVersion.EnableReleaseCandidate = true
		}
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
				"target": tgt,
				"match": strings.Join(matches, ","),
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
					"content": contentStr, 
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