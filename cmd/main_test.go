package cmd

import (
	"os"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
	libpack_logging "github.com/lukaszraczylo/graphql-monitoring-proxy/logging"
	"github.com/lukaszraczylo/pandati"
	assertions "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type Tests struct {
	suite.Suite
}

var (
	assert          *assertions.Assertions
	testCurrentPath string
)

func (suite *Tests) SetupTest() {
	err := os.Chdir(testCurrentPath)
	if err != nil {
		logger.Critical(&libpack_logging.LogMessage{Message: "Unable to change directory to test directory", Pairs: map[string]any{"error": err}})
	}
	assert = assertions.New(suite.T())
	params.varDebug = true
	params.varRepoBranch = "main"
}

func TestSuite(t *testing.T) {
	logger = libpack_logging.New()
	testCurrentPath, _ = os.Getwd()
	suite.Run(t, new(Tests))
}

func (suite *Tests) TestSetup_getSemver() {
	type fields struct {
		Semver SemVer
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "Return 1.3.7",
			fields: fields{
				Semver: SemVer{
					Major: 1,
					Minor: 3,
					Patch: 7,
				},
			},
			want: "1.3.7",
		},
		{
			name: "Return 1.3.7-rc.2",
			fields: fields{
				Semver: SemVer{
					Major:                  1,
					Minor:                  3,
					Patch:                  7,
					Release:                2,
					EnableReleaseCandidate: true,
				},
			},
			want: "1.3.7-rc.2",
		},
		{
			name: "Return 1.3.9",
			fields: fields{
				Semver: SemVer{
					Major:                  1,
					Minor:                  3,
					Patch:                  9,
					Release:                2,
					EnableReleaseCandidate: false,
				},
			},
			want: "1.3.9",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{
				Semver: tt.fields.Semver,
			}
			got := s.getSemver()
			assert.Equal(tt.want, got, "Unexpected result in "+tt.name)
		})
	}
}

func (suite *Tests) TestSetup_ForcedVersioning() {
	type fields struct {
		Force  Force
		Semver SemVer
	}
	tests := []struct {
		name   string
		want   string
		fields fields
	}{
		{
			name: "No versioning",
			want: "0.0.0",
		},
		{
			name: "Major version set",
			fields: fields{
				Force: Force{
					Major: 2,
				},
			},
			want: "2.0.0",
		},
		{
			name: "Minor version set",
			fields: fields{
				Force: Force{
					Minor: 3,
				},
			},
			want: "0.3.0",
		},
		{
			name: "Patch version set",
			fields: fields{
				Force: Force{
					Patch: 7,
				},
			},
			want: "0.0.7",
		},
		{
			name: "All versions set",
			fields: fields{
				Force: Force{
					Major: 2,
					Minor: 3,
					Patch: 4,
				},
			},
			want: "2.3.4",
		},
		{
			name: "Major and Minor set",
			fields: fields{
				Force: Force{
					Major: 2,
					Minor: 3,
				},
			},
			want: "2.3.0",
		},
		{
			name: "Minor and Patch set",
			fields: fields{
				Force: Force{
					Minor: 3,
					Patch: 4,
				},
			},
			want: "0.3.4",
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{
				Semver: tt.fields.Semver,
				Force:  tt.fields.Force,
			}
			s.ForcedVersioning()
			got := s.getSemver()
			assert.Equal(tt.want, got, "Unexpected result in "+tt.name)
		})
	}
}

func (suite *Tests) Test_checkMatches() {
	type args struct {
		content []string
		targets []string
	}
	tests := []struct {
		name      string
		args      args
		blacklist []string
		want      bool
	}{
		{
			name: "No match",
			args: args{
				content: strings.Fields("Fields splits the string s around each instance of one or more consecutive white space characters"),
				targets: []string{"github", "repository", "test"},
			},
			want: false,
		},
		{
			name: "Match",
			args: args{
				content: strings.Fields("Fields splits the string s around each instance of one or more consecutive white space characters"),
				targets: []string{"github", "repository", "instance"},
			},
			want: true,
		},
		{
			name: "Match but blacklisted",
			args: args{
				content: strings.Fields("feat: add new feature with breaking changes"),
				targets: []string{"feat", "feature"},
			},
			blacklist: []string{"breaking"},
			want:      false,
		},
		{
			name: "Match with empty blacklist",
			args: args{
				content: strings.Fields("feat: add new feature"),
				targets: []string{"feat", "feature"},
			},
			blacklist: []string{},
			want:      true,
		},
		{
			name: "No match with blacklist",
			args: args{
				content: strings.Fields("chore: update dependencies"),
				targets: []string{"feat", "feature"},
			},
			blacklist: []string{"skip-ci"},
			want:      false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			repo.Blacklist = tt.blacklist
			got := checkMatches(tt.args.content, tt.args.targets)
			assert.Equal(tt.want, got, "Unexpected result in "+tt.name)
		})
	}
}

func (suite *Tests) Test_parseExistingSemver() {
	type args struct {
		tagName string
	}
	tests := []struct {
		name                string
		args                args
		currentSemver      SemVer
		wantSemanticVersion SemVer
	}{
		{
			name: "Test parsing existing semver",
			args: args{
				tagName: "1.2.3",
			},
			currentSemver: SemVer{Major: 1, Minor: 1, Patch: 1},
			wantSemanticVersion: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name: "Test parsing existing semver with v",
			args: args{
				tagName: "v1.2.3",
			},
			currentSemver: SemVer{Major: 1, Minor: 1, Patch: 1},
			wantSemanticVersion: SemVer{
				Major: 1,
				Minor: 2,
				Patch: 3,
			},
		},
		{
			name: "Test parsing existing semver with rc",
			args: args{
				tagName: "1.2.5-rc.7",
			},
			currentSemver: SemVer{Major: 1, Minor: 1, Patch: 1},
			wantSemanticVersion: SemVer{
				Major:   1,
				Minor:   2,
				Patch:   5,
				Release: 7,
			},
		},
		{
			name: "Test invalid semver format",
			args: args{
				tagName: "invalid",
			},
			currentSemver: SemVer{Major: 2, Minor: 3, Patch: 4},
			wantSemanticVersion: SemVer{
				Major: 2,
				Minor: 3,
				Patch: 4,
			},
		},
		{
			name: "Test partial semver",
			args: args{
				tagName: "1.2",
			},
			currentSemver: SemVer{Major: 2, Minor: 3, Patch: 4},
			wantSemanticVersion: SemVer{
				Major: 2,
				Minor: 3,
				Patch: 4,
			},
		},
		{
			name: "Test empty tag",
			args: args{
				tagName: "",
			},
			currentSemver: SemVer{Major: 2, Minor: 3, Patch: 4},
			wantSemanticVersion: SemVer{
				Major: 2,
				Minor: 3,
				Patch: 4,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got := parseExistingSemver(tt.args.tagName, tt.currentSemver)
			assert.Equal(tt.wantSemanticVersion.Major, got.Major, "Unexpected MAJOR semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Minor, got.Minor, "Unexpected MINOR semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Patch, got.Patch, "Unexpected PATCH semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Release, got.Release, "Unexpected RELEASE semver result in "+tt.name)
		})
	}
}

func (suite *Tests) TestSetup_ListCommits() {
	type fields struct {
		RepositoryHandler   *git.Repository
		RepositoryName      string
		RepositoryBranch    string
		RepositoryLocalPath string
		LocalConfigFile     string
		Wording             Wording
		Commits             []CommitDetails
		Force               Force
		Semver              SemVer
	}

	tests := []struct {
		name      string
		fields    fields
		noCommits bool
		wantErr   bool
	}{
		{
			name: "List commits from existing repository",
			fields: fields{
				RepositoryName:   "https://github.com/lukaszraczylo/simple-gql-client",
				RepositoryBranch: "master",
			},
			noCommits: false,
			wantErr:   false,
		},
		{
			name: "List commits from non-existing repository",
			fields: fields{
				RepositoryName:   "https://github.com/lukaszraczylo/simple-gql-client-dead",
				RepositoryBranch: "main",
			},
			noCommits: true,
			wantErr:   true,
		},
		{
			name: "List commits starting with certain hash",
			fields: fields{
				RepositoryName:   "https://github.com/lukaszraczylo/simple-gql-client",
				RepositoryBranch: "master",
				Force: Force{
					Commit: "f6ee82113afb32ee95eac892d1155582a2f85166",
				},
			},
			noCommits: false,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{}
			s.ReadConfig(tt.fields.LocalConfigFile)
			s.RepositoryName = tt.fields.RepositoryName
			s.RepositoryBranch = tt.fields.RepositoryBranch
			s.Force = tt.fields.Force
			s.Prepare()
			listOfCommits, err := s.ListCommits()
			if !tt.wantErr {
				assert.NoError(err, "Error should not be present in "+tt.name)
			} else {
				assert.Error(err, "Error should be present in "+tt.name)
			}
			assert.Equal(tt.noCommits, pandati.IsZero(listOfCommits), "Unexpected commits count"+tt.name)
		})
	}
}

func (suite *Tests) TestSetup_ListExistingTags() {
	type fields struct {
		RepositoryHandler   *git.Repository
		RepositoryName      string
		RepositoryBranch    string
		RepositoryLocalPath string
		LocalConfigFile     string
		Wording             Wording
		Commits             []CommitDetails
		Force               Force
		Semver              SemVer
	}

	tests := []struct {
		name   string
		fields fields
		noTags bool
	}{
		{
			name: "List tags from existing repository",
			fields: fields{
				RepositoryName:   "https://github.com/lukaszraczylo/simple-gql-client",
				RepositoryBranch: "master",
			},
			noTags: false,
		},
		{
			name: "List tags from non-existing repository",
			fields: fields{
				RepositoryName:   "https://github.com/lukaszraczylo/simple-gql-client-dead",
				RepositoryBranch: "master",
			},
			noTags: true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{}
			s.ReadConfig(tt.fields.LocalConfigFile)
			s.RepositoryName = tt.fields.RepositoryName
			s.RepositoryBranch = tt.fields.RepositoryBranch
			s.Force = tt.fields.Force
			s.Prepare()
			s.ListExistingTags()
			if tt.noTags {
				assert.Equal(len(s.Tags), 0, "Unexpected number of tags in "+tt.name)
			} else {
				assert.GreaterOrEqual(len(s.Tags), 1, "Unexpected number of tags in "+tt.name)
			}
		})
	}
}

func (suite *Tests) Test_main() {
	type vars struct {
		varRepoName       string
		varRepoBranch     string
		varLocalCfg       string
		varUseLocal       bool
		varShowVersion    bool
		varDebug          bool
		varUpdate         bool
		varStrict         bool
		varGenerateInTest bool
		varExisting       bool
	}
	tests := []struct {
		name string
		vars vars
	}{
		{
			name: "Test printing version",
			vars: vars{
				varShowVersion: true,
			},
		},
		{
			name: "Test update switch",
			vars: vars{
				varUpdate: true,
			},
		},
		{
			name: "Test main",
			vars: vars{
				varGenerateInTest: false,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			params = myParams(tt.vars)
			repo = &Setup{}
			repo.LocalConfigFile = "../config.yaml"
			repo.UseLocal = true
			main()
		})
	}
}
