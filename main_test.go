package main

import (
	"testing"

	git "github.com/go-git/go-git/v5"
)

func TestSetup_getSemver(t *testing.T) {
	type fields struct {
		RepositoryName      string
		RepositoryLocalPath string
		RepositoryHandler   *git.Repository
		Commits             []CommitDetails
		Semver              SemVer
		Wording             Wording
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
		t.Run(tt.name, func(t *testing.T) {
			s := &Setup{
				RepositoryName:      tt.fields.RepositoryName,
				RepositoryLocalPath: tt.fields.RepositoryLocalPath,
				RepositoryHandler:   tt.fields.RepositoryHandler,
				Commits:             tt.fields.Commits,
				Semver:              tt.fields.Semver,
				Wording:             tt.fields.Wording,
			}
			if got := s.getSemver(); got != tt.want {
				t.Errorf("Setup.getSemver() = %v, want %v", got, tt.want)
			}
		})
	}
}
