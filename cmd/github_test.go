package cmd

import (
	"testing"

	"github.com/lukaszraczylo/semver-generator/cmd/utils"
)

func Test_checkLatestRelease(t *testing.T) {
	utils.InitLogger(true)
	tests := []struct {
		name  string
		want  string
		want1 bool
	}{
		{
			name:  "Check latest release",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got1 := checkLatestRelease()
			if got1 != tt.want1 {
				t.Errorf("checkLatestRelease() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_updatePackage(t *testing.T) {
	utils.InitLogger(true)
	if testing.Short() {
		t.Skip("Skipping test in short / CI mode")
	}
	tests := []struct {
		name string
		want bool
	}{
		{
			name: "Run autoupdater",
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updatePackage(); got != tt.want {
				t.Errorf("updatePackage() = %v, want %v", got, tt.want)
			}
		})
	}
}
