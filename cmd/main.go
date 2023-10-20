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
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/lithammer/fuzzysearch/fuzzy"
	libpack_logger "github.com/lukaszraczylo/graphql-monitoring-proxy/logging"
	"github.com/lukaszraczylo/pandati"
	"github.com/spf13/viper"
)

var (
	err         error
	repo        *Setup
	PKG_VERSION string
	logger      *libpack_logger.LogConfig
)

type Wording struct {
	Patch   []string
	Minor   []string
	Major   []string
	Release []string
}

type Force struct {
	Commit   string
	Patch    int
	Minor    int
	Major    int
	Existing bool
	Strict   bool
}

type SemVer struct {
	Patch                  int
	Minor                  int
	Major                  int
	Release                int
	EnableReleaseCandidate bool
}

type Setup struct {
	RepositoryHandler   *git.Repository
	RepositoryName      string
	RepositoryBranch    string
	RepositoryLocalPath string
	LocalConfigFile     string
	Wording             Wording
	Force               Force
	Commits             []CommitDetails
	Tags                []TagDetails
	Semver              SemVer
	Generate            bool
	UseLocal            bool
}

type CommitDetails struct {
	Timestamp time.Time
	Hash      string
	Author    string
	Message   string
}

type TagDetails struct {
	Name string
	Hash string
}

func checkMatches(content []string, targets []string) bool {
	if fuzzy.MatchNormalizedFold(strings.Join(content, " "), "Merge branch") {
		logger.Debug("semver-gen", map[string]interface{}{"message": "Merge detected, ignoring commits within", "content": strings.Join(content, " ")})
		return false
	}
	var r []string
	for _, tgt := range targets {
		r = fuzzy.FindNormalizedFold(tgt, content)
		if len(r) > 0 {
			logger.Debug("semver-gen", map[string]interface{}{"message": "Found match", "target": tgt, "match": strings.Join(r, ","), "content": strings.Join(content, " ")})
			return true
		}
	}
	return false
}

var extractNumber = regexp.MustCompile("[0-9]+")

func parseExistingSemver(tagName string, currentSemver SemVer) (semanticVersion SemVer) {
	logger.Debug("semver-gen", map[string]interface{}{"message": "Parsing existing semver", "tag": tagName})
	var tagNameParts []string
	tagNameParts = strings.Split(tagName, ".")
	if len(tagNameParts) < 3 {
		logger.Debug("semver-gen", map[string]interface{}{"message": "Unable to parse incompatible semver ( non x.y.z )"})
		return currentSemver
	}
	semanticVersion.Major, _ = strconv.Atoi(extractNumber.FindAllString(tagNameParts[0], -1)[0])
	semanticVersion.Minor, _ = strconv.Atoi(extractNumber.FindAllString(tagNameParts[1], -1)[0])
	semanticVersion.Patch, _ = strconv.Atoi(extractNumber.FindAllString(tagNameParts[2], -1)[0])
	if len(tagNameParts) > 3 {
		semanticVersion.Release, _ = strconv.Atoi(extractNumber.FindAllString(tagNameParts[3], -1)[0])
		semanticVersion.EnableReleaseCandidate = true
	}
	return
}

func (s *Setup) CalculateSemver() SemVer {
	for _, commit := range s.Commits {
		if params.varExisting || s.Force.Existing {
			for _, tagHash := range s.Tags {
				if commit.Hash == tagHash.Hash {
					logger.Debug("semver-gen", map[string]interface{}{"message": "Found existing tag", "tag": tagHash.Name, "commit": strings.TrimSuffix(commit.Message, "\n")})
					s.Semver = parseExistingSemver(tagHash.Name, s.Semver)
					continue
				}
			}
		}

		if !params.varStrict && !s.Force.Strict {
			s.Semver.Patch++
			logger.Debug("semver-gen", map[string]interface{}{"message": "Incrementing patch (DEFAULT)", "commit": strings.TrimSuffix(commit.Message, "\n"), "semver": s.getSemver()})
		}
		commitSlice := strings.Fields(commit.Message)
		matchPatch := checkMatches(commitSlice, s.Wording.Patch)
		matchMinor := checkMatches(commitSlice, s.Wording.Minor)
		matchMajor := checkMatches(commitSlice, s.Wording.Major)
		matchReleaseCandidate := checkMatches(commitSlice, s.Wording.Release)
		if matchPatch {
			s.Semver.Patch++
			logger.Debug("semver-gen", map[string]interface{}{"message": "Incrementing patch (WORDING)", "commit": strings.TrimSuffix(commit.Message, "\n"), "semver": s.getSemver()})
			continue
		}
		if matchReleaseCandidate {
			s.Semver.Release++
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = true
			logger.Debug("semver-gen", map[string]interface{}{"message": "Incrementing release candidate (WORDING)", "commit": strings.TrimSuffix(commit.Message, "\n"), "semver": s.getSemver()})
			continue
		}
		if matchMinor {
			s.Semver.Minor++
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = false
			s.Semver.Release = 0
			logger.Debug("semver-gen", map[string]interface{}{"message": "Incrementing minor (WORDING)", "commit": strings.TrimSuffix(commit.Message, "\n"), "semver": s.getSemver()})
			continue
		}
		if matchMajor {
			s.Semver.Major++
			s.Semver.Minor = 0
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = false
			s.Semver.Release = 0
			logger.Debug("semver-gen", map[string]interface{}{"message": "Incrementing major (WORDING)", "commit": strings.TrimSuffix(commit.Message, "\n"), "semver": s.getSemver()})
			continue
		}
	}
	return s.Semver
}

func (s *Setup) ListExistingTags() {
	logger.Debug("semver-gen", map[string]interface{}{"message": "Listing existing tags"})
	refs, err := s.RepositoryHandler.Tags()
	if err != nil {
		panic(err)
	}
	if err := refs.ForEach(func(ref *plumbing.Reference) error {
		s.Tags = append(s.Tags, TagDetails{Name: ref.Name().Short(), Hash: ref.Hash().String()})
		logger.Debug("semver-gen", map[string]interface{}{"message": "Found tag", "tag": ref.Name().Short(), "hash": ref.Hash().String()})
		return nil
	}); err != nil {
		panic(err)
	}
}

func (s *Setup) ListCommits() ([]CommitDetails, error) {
	var ref *plumbing.Reference
	var err error

	ref, err = s.RepositoryHandler.Head()
	if err != nil {
		return []CommitDetails{}, err
	}
	commitsList, err := s.RepositoryHandler.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return []CommitDetails{}, err
	}

	var tmpResults []CommitDetails
	commitsList.ForEach(func(c *object.Commit) error {
		tmpResults = append(tmpResults, CommitDetails{Hash: c.Hash.String(), Author: c.Author.String(), Message: c.Message, Timestamp: c.Author.When})
		sort.Slice(tmpResults, func(i, j int) bool { return tmpResults[i].Timestamp.Unix() < tmpResults[j].Timestamp.Unix() })
		return nil
	})

	logger.Debug("semver-gen", map[string]interface{}{"message": "Commits before cut", "commits": tmpResults})
	for commitId, cmt := range tmpResults {
		if s.Force.Commit != "" && cmt.Hash == s.Force.Commit {
			logger.Debug("semver-gen", map[string]interface{}{"message": "Found commit match", "commit": cmt.Hash, "index": commitId})
			s.Commits = tmpResults[commitId:]
			break
		} else {
			s.Commits = tmpResults
		}
	}

	logger.Debug("semver-gen", map[string]interface{}{"message": "Commits after cut", "commits": s.Commits})
	return s.Commits, err
}

func (s *Setup) Prepare() error {
	if !repo.UseLocal {
		u, err := url.Parse(s.RepositoryName)
		if err != nil {
			logger.Error("semver-gen", map[string]interface{}{"message": "Unable to parse repository URL", "error": err.Error(), "url": s.RepositoryName})
			return err
		}
		s.RepositoryLocalPath = fmt.Sprintf("/tmp/semver/%s/%s", u.Path, s.RepositoryBranch)
		os.RemoveAll(s.RepositoryLocalPath)
		s.RepositoryHandler, err = git.PlainClone(s.RepositoryLocalPath, false, &git.CloneOptions{
			URL:           s.RepositoryName,
			ReferenceName: plumbing.NewBranchReferenceName(s.RepositoryBranch),
			SingleBranch:  true,
			Auth: &http.BasicAuth{
				Username: os.Getenv("GITHUB_USERNAME"),
				Password: os.Getenv("GITHUB_TOKEN"),
			},
			Tags: git.AllTags,
		})
		if err != nil {
			logger.Error("semver-gen", map[string]interface{}{"message": "Unable to clone repository", "error": err.Error(), "url": s.RepositoryName})
			return err
		}
	} else {
		s.RepositoryLocalPath = "./"
		s.RepositoryHandler, err = git.PlainOpen(s.RepositoryLocalPath)
		if err != nil {
			logger.Error("semver-gen", map[string]interface{}{"message": "Unable to open local repository", "error": err.Error(), "path": s.RepositoryLocalPath})
			return err
		}
	}
	os.Chdir(s.RepositoryLocalPath)
	return err
}

func (s *Setup) ForcedVersioning() {
	if !pandati.IsZero(s.Force.Major) {
		logger.Debug("semver-gen", map[string]interface{}{"message": "Forced versioning (MAJOR)", "major": s.Force.Major})
		s.Semver.Major = s.Force.Major
	}
	if !pandati.IsZero(s.Force.Minor) {
		logger.Debug("semver-gen", map[string]interface{}{"message": "Forced versioning (MINOR)", "minor": s.Force.Minor})
		s.Semver.Minor = s.Force.Minor
	}
	if !pandati.IsZero(s.Force.Patch) {
		logger.Debug("semver-gen", map[string]interface{}{"message": "Forced versioning (PATCH)", "patch": s.Force.Patch})
		s.Semver.Patch = s.Force.Patch
	}
}

func (s *Setup) ReadConfig(file string) error {
	viper.SetConfigFile(file)
	err := viper.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("Fatal error config file: %s \n", err)
		return err
	}
	viper.UnmarshalKey("wording", &s.Wording)
	viper.UnmarshalKey("force", &s.Force)
	return err
}

func (s *Setup) getSemver() (semverReturned string) {
	semverReturned = fmt.Sprintf("%d.%d.%d", s.Semver.Major, s.Semver.Minor, s.Semver.Patch)
	if s.Semver.EnableReleaseCandidate {
		semverReturned = fmt.Sprintf("%s-rc.%d", semverReturned, s.Semver.Release)
	}
	return semverReturned
}

func main() {
	logger = libpack_logger.NewLogger()
	if params.varShowVersion {
		var outdatedMsg string
		latestRelease, latestRelaseOk := checkLatestRelease()
		if PKG_VERSION != latestRelease && latestRelaseOk {
			outdatedMsg = fmt.Sprintf("(Latest available: %s)", latestRelease)
		}
		logger.Info("semver-gen", map[string]interface{}{"version": PKG_VERSION, "outdated": outdatedMsg})
		if outdatedMsg != "" {
			logger.Info("semver-gen", map[string]interface{}{"message": "You can update automatically with: semver-gen -u"})
		}
		return
	}
	if params.varUpdate {
		updatePackage()
		return
	}
	if repo.Generate || params.varGenerateInTest {
		err := repo.ReadConfig(repo.LocalConfigFile)
		if err != nil {
			logger.Error("semver-gen", map[string]interface{}{"message": "Unable to find config file semver.yaml. Using defaults and flags.", "file": repo.LocalConfigFile})
		}
		err = repo.Prepare()
		if err != nil {
			logger.Critical("semver-gen", map[string]interface{}{"message": "Unable to prepare repository", "error": err.Error()})
		}
		repo.ListCommits()
		if params.varExisting || repo.Force.Existing {
			repo.ListExistingTags()
		}
		repo.ForcedVersioning()
		repo.CalculateSemver()
		fmt.Println("SEMVER", repo.getSemver())
	}
}
