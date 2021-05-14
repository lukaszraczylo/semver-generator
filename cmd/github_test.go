package cmd

import "testing"

func Test_checkLatestRelease(t *testing.T) {
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
