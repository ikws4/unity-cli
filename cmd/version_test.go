package cmd

import "testing"

func TestDefaultVersion(t *testing.T) {
	if DefaultVersion != "0.8.1" {
		t.Fatalf("DefaultVersion = %q, want 0.8.1; increment this assertion with each feature", DefaultVersion)
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name      string
		candidate string
		current   string
		want      bool
	}{
		{"newer patch", "v0.8.1", "0.8.0", true},
		{"newer minor", "0.9.0", "v0.8.9", true},
		{"same with v prefix", "v0.8.0", "0.8.0", false},
		{"older version", "v0.3.22", "0.8.0", false},
		{"empty candidate", "", "0.8.0", false},
		{"release newer than prerelease", "0.8.0", "0.8.0-beta.1", true},
		{"prerelease older than release", "0.8.0-beta.1", "0.8.0", false},
		{"newer prerelease", "0.8.0-beta.2", "0.8.0-beta.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNewerVersion(tt.candidate, tt.current); got != tt.want {
				t.Fatalf("isNewerVersion(%q, %q) = %v, want %v", tt.candidate, tt.current, got, tt.want)
			}
		})
	}
}
