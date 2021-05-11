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
	"sort"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/lukaszraczylo/zero"
	"github.com/spf13/viper"
)

var (
	err         error
	repo        *Setup
	PKG_VERSION string
)

type Wording struct {
	Patch []string
	Minor []string
	Major []string
}

type Force struct {
	Patch  int
	Minor  int
	Major  int
	Commit string
}

type SemVer struct {
	Patch int
	Minor int
	Major int
}

type Setup struct {
	RepositoryName      string
	RepositoryLocalPath string
	RepositoryHandler   *git.Repository
	Commits             []CommitDetails
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

func checkMatches(content []string, targets []string) bool {
	var r []string
	for _, tgt := range targets {
		r = fuzzy.FindFold(tgt, content)
	}
	return (len(r) > 0)
}

func (s *Setup) CalculateSemver() SemVer {
	for _, commit := range s.Commits {
		s.Semver.Patch++
		if varDebug {
			fmt.Println("DEBUG: Incrementing patch (DEFAULT) on ", commit.Message)
		}
		commitSlice := strings.Split(commit.Message, " ")
		matchPatch := checkMatches(commitSlice, s.Wording.Patch)
		matchMinor := checkMatches(commitSlice, s.Wording.Minor)
		matchMajor := checkMatches(commitSlice, s.Wording.Major)
		if matchPatch {
			if varDebug {
				fmt.Println("DEBUG: Incrementing patch (WORDING) on ", commit.Message)
			}
			s.Semver.Patch++
		}
		if matchMinor {
			if varDebug {
				fmt.Println("DEBUG: Incrementing minor (WORDING) on ", commit.Message)
			}
			s.Semver.Minor++
			s.Semver.Patch = 1
		}
		if matchMajor {
			if varDebug {
				fmt.Println("DEBUG: Incrementing major (WORDING) on ", commit.Message)
			}
			s.Semver.Major++
			s.Semver.Minor = 0
			s.Semver.Patch = 1
		}
	}
	return s.Semver
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

	if varDebug {
		fmt.Println("\n---COMMITS BEFORE CUT---\n", tmpResults)
	}

	for commitId, cmt := range tmpResults {
		if s.Force.Commit != "" && cmt.Hash == s.Force.Commit {
			s.Commits = tmpResults[commitId:]
		} else {
			s.Commits = tmpResults
		}
	}

	if varDebug {
		fmt.Println("\n---COMMITS AFTER CUT---\n", s.Commits)
	}

	return s.Commits, err
}

func (s *Setup) Prepare() error {
	if !repo.UseLocal {
		u, err := url.Parse(s.RepositoryName)
		if err != nil {
			fmt.Println("Unable to parse repository URL", err.Error())
			return err
		}
		s.RepositoryLocalPath = fmt.Sprintf("/tmp/foo/%s", u.Path)
		os.RemoveAll(s.RepositoryLocalPath)
		s.RepositoryHandler, err = git.PlainClone(s.RepositoryLocalPath, false, &git.CloneOptions{
			URL: s.RepositoryName,
		})
		if err != nil {
			fmt.Println("Unable to reach repository", err.Error())
			return err
		}
	} else {
		s.RepositoryHandler, err = git.PlainOpen(s.RepositoryLocalPath)
		if err != nil {
			fmt.Println("Unable to reach repository", err.Error())
			return err
		}
	}
	return err
}

func (s *Setup) ForcedVersioning() {
	if !zero.IsZero(s.Force.Major) {
		if varDebug {
			fmt.Println("DEBUG: Forced versioning (MAJOR)", s.Force.Major)
		}
		s.Semver.Major = s.Force.Major
	}
	if !zero.IsZero(s.Force.Minor) {
		if varDebug {
			fmt.Println("DEBUG: Forced versioning (MINOR)", s.Force.Minor)
		}
		s.Semver.Minor = s.Force.Minor
	}
	if !zero.IsZero(s.Force.Patch) {
		if varDebug {
			fmt.Println("DEBUG: Forced versioning (PATCH)", s.Force.Patch)
		}
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

func (s *Setup) getSemver() string {
	return fmt.Sprintf("%d.%d.%d", s.Semver.Major, s.Semver.Minor, s.Semver.Patch)
}

func main() {
	if varShowVersion {
		checkLatestRelease()
		fmt.Println("semver-gen", PKG_VERSION, "\tMore information: https://github.com/lukaszraczylo/semver-generator")
		return
	}
	if repo.Generate {
		err := repo.ReadConfig(repo.LocalConfigFile)
		if err != nil {
			panic(err)
		}
		err = repo.Prepare()
		if err != nil {
			panic(err)
		}
		repo.ListCommits()
		repo.ForcedVersioning()
		repo.CalculateSemver()
		fmt.Println("SEMVER", repo.getSemver())
	}
}
