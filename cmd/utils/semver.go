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
) SemVer {
	semver := initialSemver

	for _, commit := range commits {
		// Check for existing tags if enabled
		if respectExisting {
			for _, tagHash := range tags {
				if commit.Hash == tagHash.Hash {
					Debug("Found existing tag", map[string]interface{}{
						"tag": tagHash.Name, 
						"commit": strings.TrimSuffix(commit.Message, "\n"),
					})
					semver = ParseExistingSemver(tagHash.Name, semver)
					continue
				}
			}
		}

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