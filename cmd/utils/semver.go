package utils

import (
	"strings"
)

// CalculateSemver calculates the semantic version based on commit messages
func CalculateSemver(
	commits []CommitDetails,
	tags []TagDetails,
	wording Wording,
	blacklist []string,
	initialSemver SemVer,
	respectExisting bool,
	strictMode bool,
	tagPrefixes []string,
) SemVer {
	semver := initialSemver
	startIndex := 0

	// If respecting existing tags, find the latest tagged commit and start from there
	if respectExisting && len(tags) > 0 {
		latestTagIndex := -1
		var latestTagName string

		// Find the latest tagged commit (highest index since commits are oldest-first)
		for i, commit := range commits {
			for _, tag := range tags {
				if commit.Hash == tag.Hash {
					if i > latestTagIndex {
						latestTagIndex = i
						latestTagName = tag.Name
					}
				}
			}
		}

		// If we found a tagged commit, use its version and start processing after it
		if latestTagIndex >= 0 {
			Debug("Found latest existing tag", map[string]interface{}{
				"tag":    latestTagName,
				"commit": strings.TrimSuffix(commits[latestTagIndex].Message, "\n"),
			})
			semver = ParseExistingSemver(latestTagName, semver, tagPrefixes)
			startIndex = latestTagIndex + 1
		}
	}

	for _, commit := range commits[startIndex:] {
		// In non-strict mode, increment patch by default
		if !strictMode {
			semver.Patch++
			Debug("Incrementing patch (DEFAULT)", map[string]interface{}{
				"commit": strings.TrimSuffix(commit.Message, "\n"),
				"semver": FormatSemver(semver),
			})
		}

		// Check for keyword matches
		commitSlice := strings.Fields(commit.Message)
		matchPatch := CheckMatches(commitSlice, wording.Patch, blacklist)
		matchMinor := CheckMatches(commitSlice, wording.Minor, blacklist)
		matchMajor := CheckMatches(commitSlice, wording.Major, blacklist)
		matchReleaseCandidate := CheckMatches(commitSlice, wording.Release, blacklist)

		// Apply version changes based on matches
		if matchMajor {
			semver.Major++
			semver.Minor = 0
			semver.Patch = 1
			semver.EnableReleaseCandidate = false
			semver.Release = 0
			Debug("Incrementing major (WORDING)", map[string]interface{}{
				"commit": strings.TrimSuffix(commit.Message, "\n"),
				"semver": FormatSemver(semver),
			})
			continue
		}

		if matchMinor {
			semver.Minor++
			semver.Patch = 1
			semver.EnableReleaseCandidate = false
			semver.Release = 0
			Debug("Incrementing minor (WORDING)", map[string]interface{}{
				"commit": strings.TrimSuffix(commit.Message, "\n"),
				"semver": FormatSemver(semver),
			})
			continue
		}

		if matchReleaseCandidate {
			semver.Release++
			semver.Patch = 1
			semver.EnableReleaseCandidate = true
			Debug("Incrementing release candidate (WORDING)", map[string]interface{}{
				"commit": strings.TrimSuffix(commit.Message, "\n"),
				"semver": FormatSemver(semver),
			})
			continue
		}

		if matchPatch {
			semver.Patch++
			Debug("Incrementing patch (WORDING)", map[string]interface{}{
				"commit": strings.TrimSuffix(commit.Message, "\n"),
				"semver": FormatSemver(semver),
			})
			continue
		}
	}

	return semver
}
