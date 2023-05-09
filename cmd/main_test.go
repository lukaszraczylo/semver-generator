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
	"os"
	"strings"
	"testing"

	git "github.com/go-git/go-git/v5"
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
	os.Chdir(testCurrentPath)
	assert = assertions.New(suite.T())
	params.varDebug = true
	params.varRepoBranch = "main"
}

func TestSuite(t *testing.T) {
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

func (suite *Tests) TestSetup_Prepare() {
	type fields struct {
		RepositoryName      string
		RepositoryLocalPath string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Test repository lukaszraczylo/simple-gql-client",
			fields: fields{
				RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client",
			},
			wantErr: true,
		},
		{
			name: "Test non-existing repository",
			fields: fields{
				RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client-dead",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{
				RepositoryName: tt.fields.RepositoryName,
			}
			s.Prepare()
			if _, err := os.Stat(s.RepositoryLocalPath); os.IsNotExist(err) {
				if !tt.wantErr {
					assert.NoError(err, "Error should not be present in "+tt.name)
				} else {
					assert.Error(err, "Error should be present in "+tt.name)
				}
			}
		})
	}
}

func (suite *Tests) TestSetup_ReadConfig() {
	type fields struct {
		Wording Wording
		Force   Force
	}
	type args struct {
		file string
	}
	tests := []struct {
		name         string
		args         args
		fields       fields
		wordingEmpty bool
		wantErr      bool
	}{
		{
			name: "Test non-existent config file",
			args: args{
				file: "random-file-name.yaml",
			},
			wordingEmpty: true,
			wantErr:      true,
		},
		{
			name: "Test existing config file",
			args: args{
				file: "../config.yaml",
			},
			wordingEmpty: false,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{}
			err := s.ReadConfig(tt.args.file)
			if !tt.wantErr {
				assert.NoError(err, "Error should not be present in "+tt.name)
			} else {
				assert.Error(err, "Error should be present in "+tt.name)
			}
			assert.Equal(tt.wordingEmpty, pandati.IsZero(s.Wording), "Unexpected wording count "+tt.name+":", s.Wording)
		})
	}
}

func (suite *Tests) Test_checkMatches() {
	type args struct {
		content []string
		targets []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "No match",
			args: args{
				content: strings.Fields("Fields splits the string s around each instance of one or more consecutive white space characters, as defined by unicode.IsSpace, returning a slice of substrings of s or an empty slice if s contains only white space"),
				targets: []string{"github", "repository", "test"},
			},
			want: false,
		},
		{
			name: "Match",
			args: args{
				content: strings.Fields("Fields splits the string s around each instance of one or more consecutive white space characters, as defined by unicode.IsSpace, returning a slice of substrings of s or an empty slice if s contains only white space"),
				targets: []string{"github", "repository", "instance"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got := checkMatches(tt.args.content, tt.args.targets)
			assert.Equal(tt.want, got, "Unexpected result in "+tt.name)
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
		Force               Force
		Commits             []CommitDetails
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

func (suite *Tests) TestSetup_CalculateSemver() {
	type fields struct {
		RepositoryName  string
		BranchName      string
		LocalConfigFile string
		Force           Force
	}
	type wantSemver struct {
		Major int
		Minor int
		Patch int
	}
	tests := []struct {
		name           string
		fields         fields
		wantSemver     wantSemver
		strictMatching bool
	}{
		{
			name: "Test on existing repository",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				BranchName:      "main",
			},
			strictMatching: false,
			wantSemver: wantSemver{
				Major: 0,
				Minor: 0,
				Patch: 7,
			},
		},
		{
			name: "Test on existing repository with strict matching",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				BranchName:      "main",
			},
			strictMatching: true,
			wantSemver: wantSemver{
				Major: 2,
				Minor: 4,
				Patch: 1,
			},
		},
		{
			name: "Test on existing repository, starting with certain hash",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				BranchName:      "main",
				Force: Force{
					Major:  1,
					Minor:  1,
					Commit: "45f9a23cec39e94503841638aee3efecd45111cf",
				},
			},
			strictMatching: false,
			wantSemver: wantSemver{
				Major: 1,
				Minor: 5,
				Patch: 1,
			},
		},
		{
			name: "Test on existing repository, starting with different hash",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				BranchName:      "main",
				Force: Force{
					Major:  1,
					Minor:  1,
					Commit: "48564920d88a8a16df607736b438947309ffb8c6",
				},
			},
			strictMatching: false,
			wantSemver: wantSemver{
				Major: 1,
				Minor: 4,
				Patch: 1,
			},
		},
		{
			name: "Test on non-existing repository",
			fields: fields{
				RepositoryName: "https://github.com/lukaszraczylo/semver-generator-test-repo-dead",
			},
			wantSemver: wantSemver{
				Major: 1, // 1 because config file enforces MAJOR version
				Minor: 1, // 1 because config file enforces MINOR version
				Patch: 0,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Setup{}
			s.ReadConfig(tt.fields.LocalConfigFile)
			s.RepositoryName = tt.fields.RepositoryName
			s.RepositoryBranch = tt.fields.BranchName
			s.Prepare()
			s.ForcedVersioning()
			s.Force = tt.fields.Force
			s.ListCommits()
			params.varStrict = tt.strictMatching
			semver := s.CalculateSemver()
			assert.Equal(tt.wantSemver.Major, semver.Major, "Unexpected MAJOR semver result in "+tt.name)
			assert.Equal(tt.wantSemver.Minor, semver.Minor, "Unexpected MINOR semver result in "+tt.name)
			assert.Equal(tt.wantSemver.Patch, semver.Patch, "Unexpected PATCH semver result in "+tt.name)
		})
	}
}

func (suite *Tests) Test_debugPrint() {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test debug print",
			args: args{
				content: "Test debug",
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			debugPrint(tt.args.content)
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

func (suite *Tests) Test_parseExistingSemver() {
	type args struct {
		tagName string
	}
	tests := []struct {
		name                string
		args                args
		wantSemanticVersion SemVer
	}{
		{
			name: "Test parsing existing semver",
			args: args{
				tagName: "1.2.3",
			},
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
			wantSemanticVersion: SemVer{
				Major:   1,
				Minor:   2,
				Patch:   5,
				Release: 7,
			},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			got := parseExistingSemver(tt.args.tagName, SemVer{
				Major: 1,
				Minor: 1,
				Patch: 1,
			})
			assert.Equal(tt.wantSemanticVersion.Major, got.Major, "Unexpected MAJOR semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Minor, got.Minor, "Unexpected MINOR semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Patch, got.Patch, "Unexpected PATCH semver result in "+tt.name)
			assert.Equal(tt.wantSemanticVersion.Release, got.Release, "Unexpected RELEASE semver result in "+tt.name)
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
		Force               Force
		Commits             []CommitDetails
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
