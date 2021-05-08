package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/lukaszraczylo/zero"
	"github.com/spf13/viper"
)

var (
	err error
)

type Wording struct {
	Patch []string
	Minor []string
	Major []string
}

type Force struct {
	Patch int
	Minor int
	Major int
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

func (s *Setup) CalculateSemver() {
	for _, commit := range s.Commits {
		s.Semver.Patch++
		commitSlice := strings.Split(commit.Message, " ")
		matchPatch := checkMatches(commitSlice, s.Wording.Patch)
		matchMinor := checkMatches(commitSlice, s.Wording.Minor)
		matchMajor := checkMatches(commitSlice, s.Wording.Major)
		// fmt.Println(matchPatch, matchMinor, matchMajor, commit.Message, s.Wording.Patch, s.Wording.Minor, s.Wording.Major)
		if matchPatch {
			s.Semver.Patch++
			fmt.Println("Patch version bumped:", commit.Message)
		}
		if matchMinor {
			s.Semver.Minor++
			s.Semver.Patch = 1
			fmt.Println("Minor version bumped:", commit.Message)
		}
		if matchMajor {
			s.Semver.Major++
			s.Semver.Minor = 0
			s.Semver.Patch = 1
			fmt.Println("Major version bumped:", commit.Message)
		}
	}
}

func (s *Setup) ListCommits() {
	ref, _ := s.RepositoryHandler.Head()
	commitsList, err := s.RepositoryHandler.Log(&git.LogOptions{From: ref.Hash(), Order: git.LogOrderBSF})
	if err != nil {
		panic(err)
	}
	commitsList.ForEach(func(c *object.Commit) error {
		s.Commits = append(s.Commits, CommitDetails{Hash: c.Hash.String(), Author: c.Author.String(), Message: c.Message, Timestamp: c.Author.When})
		sort.Slice(s.Commits, func(i, j int) bool { return s.Commits[i].Timestamp.Unix() < s.Commits[j].Timestamp.Unix() })
		return nil
	})
}

func (s *Setup) Prepare() {
	u, _ := url.Parse(s.RepositoryName)
	s.RepositoryLocalPath = fmt.Sprintf("/tmp/foo/%s", u.Path)
	os.RemoveAll(s.RepositoryLocalPath)
	s.RepositoryHandler, err = git.PlainClone(s.RepositoryLocalPath, false, &git.CloneOptions{
		URL: s.RepositoryName,
	})
	if err != nil {
		panic(err)
	}
}

func (s *Setup) ForcedVersioning() {
	if !zero.IsZero(s.Force.Major) {
		s.Semver.Major = s.Force.Major
	}
	if !zero.IsZero(s.Force.Minor) {
		s.Semver.Minor = s.Force.Minor
	}
	if !zero.IsZero(s.Force.Patch) {
		s.Semver.Patch = s.Force.Patch
	}
}

func (s *Setup) ReadConfig(file string) {
	viper.SetConfigFile(file)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viper.UnmarshalKey("wording", &s.Wording)
	viper.UnmarshalKey("force", &s.Force)
}

func (s *Setup) getSemver() string {
	return fmt.Sprintf("%d.%d.%d", s.Semver.Major, s.Semver.Minor, s.Semver.Patch)
}

func main() {
	repo := &Setup{
		RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client",
	}
	repo.ReadConfig("config.yaml")
	repo.Prepare()
	fmt.Println("Repo local path:", repo.RepositoryLocalPath)
	repo.ListCommits()
	repo.ForcedVersioning()
	repo.CalculateSemver()
	fmt.Println("Calculated semver:", repo.getSemver())
}
