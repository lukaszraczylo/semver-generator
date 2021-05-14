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
	"github.com/lukaszraczylo/zero"
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
	varDebug = true
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
		fields fields
		want   string
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
		Semver SemVer
		Force  Force
	}
	tests := []struct {
		name   string
		fields fields
		want   string
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
		fields       fields
		args         args
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
			assert.Equal(tt.wordingEmpty, zero.IsZero(s.Wording), "Unexpected wording count "+tt.name+":", s.Wording)
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
		RepositoryName      string
		RepositoryLocalPath string
		RepositoryHandler   *git.Repository
		LocalConfigFile     string
		Commits             []CommitDetails
		Semver              SemVer
		Wording             Wording
		Force               Force
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
				RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client",
			},
			noCommits: false,
			wantErr:   false,
		},
		{
			name: "List commits from non-existing repository",
			fields: fields{
				RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client-dead",
			},
			noCommits: true,
			wantErr:   true,
		},
		{
			name: "List commits starting with certain hash",
			fields: fields{
				RepositoryName: "https://github.com/lukaszraczylo/simple-gql-client",
				Force: Force{
					Commit: "97d3682ed94168600926f9ff6da650403d1f3317",
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
			s.Force = tt.fields.Force
			s.Prepare()
			listOfCommits, err := s.ListCommits()
			if !tt.wantErr {
				assert.NoError(err, "Error should not be present in "+tt.name)
			} else {
				assert.Error(err, "Error should be present in "+tt.name)
			}
			assert.Equal(tt.noCommits, zero.IsZero(listOfCommits), "Unexpected commits count"+tt.name)
		})
	}
}

func (suite *Tests) TestSetup_CalculateSemver() {
	type fields struct {
		RepositoryName  string
		Force           Force
		LocalConfigFile string
	}
	type wantSemver struct {
		Major int
		Minor int
		Patch int
	}
	tests := []struct {
		name       string
		fields     fields
		wantSemver wantSemver
	}{
		{
			name: "Test on existing repository",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				Force: Force{
					Commit: "",
				},
			},
			wantSemver: wantSemver{
				Major: 0,
				Minor: 0,
				Patch: 7,
			},
		},
		{
			name: "Test on existing repository, starting with certain hash",
			fields: fields{
				RepositoryName:  "https://github.com/lukaszraczylo/semver-generator-test-repo",
				LocalConfigFile: "meta.yaml",
				Force: Force{
					Commit: "45f9a23cec39e94503841638aee3efecd45111cf",
				},
			},
			wantSemver: wantSemver{
				Major: 2,
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
			s.Prepare()
			s.ForcedVersioning()
			s.Force = tt.fields.Force
			s.ListCommits()
			semver := s.CalculateSemver()
			assert.Equal(tt.wantSemver.Major, semver.Major, "Unexpected MAJOR semver result in "+tt.name)
			assert.Equal(tt.wantSemver.Minor, semver.Minor, "Unexpected MINOR semver result in "+tt.name)
			assert.Equal(tt.wantSemver.Patch, semver.Patch, "Unexpected PATCH semver result in "+tt.name)
		})
	}
}
