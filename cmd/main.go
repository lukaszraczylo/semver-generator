// Project: semver-generator
/*
Copyright © 2021 LUKASZ RACZYLO <lukasz$raczylo,com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/lukaszraczylo/semver-generator/cmd/utils"
)

var (
	err         error
	repo        *Setup
	PKG_VERSION string
)

// Setup represents the application setup
type Setup struct {
	RepositoryName      string
	RepositoryBranch    string
	LocalConfigFile     string
	Generate            bool
	UseLocal            bool
	GitRepo             utils.GitRepository
	Config              *utils.Config
	Semver              utils.SemVer
}

// Initialize the fuzzy search function in the utils package
func init() {
	utils.InitLogger(false) // Will be updated in main based on debug flag
	
	// Set the fuzzy search function
	utils.FuzzyFind = fuzzy.FindNormalizedFold
}

// getSemver returns the semantic version as a string
func (s *Setup) getSemver() string {
	return utils.FormatSemver(s.Semver)
}

// main is the entry point for the application
func main() {
	// Initialize logger
	if params.varDebug {
		utils.InitLogger(true)
	} else {
		utils.InitLogger(false)
	}

	// Show version if requested
	if params.varShowVersion {
		var outdatedMsg string
		latestRelease, latestReleaseOk := utils.CheckLatestRelease()
		if PKG_VERSION != latestRelease && latestReleaseOk {
			outdatedMsg = fmt.Sprintf("(Latest available: %s)", latestRelease)
		}
		
		utils.Info("semver-gen", map[string]interface{}{
			"version": PKG_VERSION,
			"outdated": outdatedMsg,
		})
		
		if outdatedMsg != "" {
			utils.Info("semver-gen", map[string]interface{}{
				"message": "You can update automatically with: semver-gen -u",
			})
		}
		return
	}

	// Update package if requested
	if params.varUpdate {
		utils.UpdatePackage()
		return
	}

	// Generate semantic version
	if repo.Generate || params.varGenerateInTest {
		// Read configuration
		config, err := utils.ReadConfig(repo.LocalConfigFile)
		if err != nil {
			utils.Error("Unable to find config file. Using defaults and flags.", map[string]interface{}{
				"file": repo.LocalConfigFile,
			})
		}
		repo.Config = config

		// Setup git repository
		gitRepo := utils.GitRepository{
			Name:        repo.RepositoryName,
			Branch:      repo.RepositoryBranch,
			UseLocal:    repo.UseLocal,
			StartCommit: repo.Config.Force.Commit,
		}
		repo.GitRepo = gitRepo

		// Prepare repository
		err = utils.PrepareRepository(&repo.GitRepo)
		if err != nil {
			utils.Critical("Unable to prepare repository", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}

		// List commits
		utils.ListCommits(&repo.GitRepo)

		// List existing tags if needed
		if params.varExisting || repo.Config.Force.Existing {
			utils.ListExistingTags(&repo.GitRepo)
		}

		// Apply forced versioning
		utils.ApplyForcedVersioning(repo.Config.Force, &repo.Semver)

		// Calculate semantic version
		repo.Semver = utils.CalculateSemver(
			repo.GitRepo.Commits,
			repo.GitRepo.Tags,
			repo.Config.Wording,
			repo.Config.Blacklist,
			repo.Semver,
			params.varExisting || repo.Config.Force.Existing,
			params.varStrict || repo.Config.Force.Strict,
		)

		// Print semantic version
		fmt.Println("SEMVER", repo.getSemver())
	}
}
