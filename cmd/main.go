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
	"flag"
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
	"github.com/lukaszraczylo/pandati"
	"github.com/spf13/viper"
)

var (
	err         error
	repo        *Setup
	PKG_VERSION string
)

type Wording struct {
	Patch   []string
	Minor   []string
	Major   []string
	Release []string
}

type Force struct {
	Patch    int
	Minor    int
	Major    int
	Commit   string
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
	RepositoryName      string
	RepositoryLocalPath string
	RepositoryHandler   *git.Repository
	Commits             []CommitDetails
	Tags                []TagDetails
	Semver              SemVer
	Wording             Wording
	Force               Force
	Generate            bool
	LocalConfigFile     string
	UseLocal            bool
}

type CommitDetails struct {
	Hash      string
	Author    string
	Message   string
	Timestamp time.Time
}

type TagDetails struct {
	Name string
	Hash string
}

func checkMatches(content []string, targets []string) bool {
	if fuzzy.MatchNormalizedFold(strings.Join(content, " "), "Merge branch") {
		debugPrint(fmt.Sprintln("Merge detected, ignoring commits within:", content))
		return false
	}
	var r []string
	for _, tgt := range targets {
		r = fuzzy.FindNormalizedFold(tgt, content)
		if len(r) > 0 {
			debugPrint(fmt.Sprintln("Found match for ", tgt, "|", strings.Join(r, ",")))
			return true
		}
	}
	return false
}

func debugPrint(content string) {
	if params.varDebug && flag.Lookup("test.v") == nil {
		fmt.Println("DEBUG:", content)
	}
}

var extractNumber = regexp.MustCompile("[0-9]+")

func parseExistingSemver(tagName string, currentSemver SemVer) (semanticVersion SemVer) {
	debugPrint(fmt.Sprintln("Parsing existing semver:", tagName))
	var tagNameParts []string
	tagNameParts = strings.Split(tagName, ".")
	if len(tagNameParts) < 3 {
		debugPrint("Unable to parse incompatible semver ( non x.y.z )")
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
					debugPrint(fmt.Sprintln("Found existing tag:", tagHash.Name, "related to", commit.Message))
					s.Semver = parseExistingSemver(tagHash.Name, s.Semver)
					continue
				}
			}
		}

		if !params.varStrict && !s.Force.Strict {
			s.Semver.Patch++
			debugPrint(fmt.Sprintln("Incrementing patch (DEFAULT) on ", strings.TrimSuffix(commit.Message, "\n"), "| Semver:", s.getSemver()))
		}
		commitSlice := strings.Fields(commit.Message)
		matchPatch := checkMatches(commitSlice, s.Wording.Patch)
		matchMinor := checkMatches(commitSlice, s.Wording.Minor)
		matchMajor := checkMatches(commitSlice, s.Wording.Major)
		matchReleaseCandidate := checkMatches(commitSlice, s.Wording.Release)
		if matchPatch {
			s.Semver.Patch++
			debugPrint(fmt.Sprintln("Incrementing patch (WORDING) on ", strings.TrimSuffix(commit.Message, "\n"), "| Semver:", s.getSemver()))
			continue
		}
		if matchReleaseCandidate {
			s.Semver.Release++
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = true
			debugPrint(fmt.Sprintln("Incrementing release candidate (WORDING) on ", strings.TrimSuffix(commit.Message, "\n"), "| Semver:", s.getSemver()))
			continue
		}
		if matchMinor {
			s.Semver.Minor++
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = false
			s.Semver.Release = 0
			debugPrint(fmt.Sprintln("Incrementing minor (WORDING) on ", strings.TrimSuffix(commit.Message, "\n"), "| Semver:", s.getSemver()))
			continue
		}
		if matchMajor {
			s.Semver.Major++
			s.Semver.Minor = 0
			s.Semver.Patch = 1
			s.Semver.EnableReleaseCandidate = false
			s.Semver.Release = 0
			debugPrint(fmt.Sprintln("Incrementing major (WORDING) on ", strings.TrimSuffix(commit.Message, "\n"), "| Semver:", s.getSemver()))
			continue
		}
	}
	return s.Semver
}

func (s *Setup) ListExistingTags() {
	debugPrint(fmt.Sprintln("Listing existing tags"))
	refs, err := s.RepositoryHandler.Tags()
	if err != nil {
		panic(err)
	}
	if err := refs.ForEach(func(ref *plumbing.Reference) error {
		s.Tags = append(s.Tags, TagDetails{Name: ref.Name().Short(), Hash: ref.Hash().String()})
		debugPrint(fmt.Sprintln("Found tag:", ref.Name().Short(), ref.Hash().String()))
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

	debugPrint(fmt.Sprintln("\n---COMMITS BEFORE CUT---\n", s.Commits))

	for commitId, cmt := range tmpResults {
		if s.Force.Commit != "" && cmt.Hash == s.Force.Commit {
			debugPrint(fmt.Sprintln(">>>> FOUND MATCH", len(s.Commits), len(tmpResults[commitId:])))
			s.Commits = tmpResults[commitId:]
			break
		} else {
			s.Commits = tmpResults
		}
	}

	debugPrint(fmt.Sprintln("\n---COMMITS AFTER CUT---\n", s.Commits))
	return s.Commits, err
}

func (s *Setup) Prepare() error {
	if !repo.UseLocal {
		u, err := url.Parse(s.RepositoryName)
		if err != nil {
			fmt.Println("Unable to parse repository URL", s.RepositoryName, "Error:", err.Error())
			return err
		}
		s.RepositoryLocalPath = fmt.Sprintf("/tmp/semver/%s", u.Path)
		os.RemoveAll(s.RepositoryLocalPath)
		s.RepositoryHandler, err = git.PlainClone(s.RepositoryLocalPath, false, &git.CloneOptions{
			URL: s.RepositoryName,
			Auth: &http.BasicAuth{
				Username: os.Getenv("GITHUB_USERNAME"),
				Password: os.Getenv("GITHUB_TOKEN"),
			},
			Tags: git.AllTags,
		})
		if err != nil {
			fmt.Println("Unable to reach repository", s.RepositoryName, "Error:", err.Error())
			return err
		}
	} else {
		s.RepositoryLocalPath = "./"
		s.RepositoryHandler, err = git.PlainOpen(s.RepositoryLocalPath)
		if err != nil {
			fmt.Println("Unable to reach repository", s.RepositoryName, "Error:", err.Error())
			return err
		}
	}
	os.Chdir(s.RepositoryLocalPath)
	return err
}

func (s *Setup) ForcedVersioning() {
	if !pandati.IsZero(s.Force.Major) {
		debugPrint(fmt.Sprintln("Forced versioning (MAJOR)", s.Force.Major))
		s.Semver.Major = s.Force.Major
	}
	if !pandati.IsZero(s.Force.Minor) {
		debugPrint(fmt.Sprintln("Forced versioning (MINOR)", s.Force.Minor))
		s.Semver.Minor = s.Force.Minor
	}
	if !pandati.IsZero(s.Force.Patch) {
		debugPrint(fmt.Sprintln("Forced versioning (PATCH)", s.Force.Patch))
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
	if params.varShowVersion {
		var outdatedMsg string
		latestRelease, latestRelaseOk := checkLatestRelease()
		if PKG_VERSION != latestRelease && latestRelaseOk {
			outdatedMsg = fmt.Sprintf("(Latest available: %s)", latestRelease)
		}
		fmt.Println("semver-gen", PKG_VERSION, "", outdatedMsg, "\tMore information: https://github.com/lukaszraczylo/semver-generator")
		if outdatedMsg != "" {
			fmt.Println("You can update automatically with: semver-gen -u")
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
			fmt.Println("Unable to find config file semver.yaml. Using defaults and flags.", repo.LocalConfigFile)
		}
		err = repo.Prepare()
		if err != nil {
			fmt.Println("Unable to prepare repository")
			os.Exit(1)
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
