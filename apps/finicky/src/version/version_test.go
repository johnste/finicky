package version

import "testing"

func TestIsUpdateAvailable(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
		wantErr  bool
	}{
		{"same version", "1.0.0", "1.0.0", false, false},
		{"newer version available", "1.0.0", "1.0.1", true, false},
		{"older version", "1.0.1", "1.0.0", false, false},
		{"v-prefixed versions", "v1.0.0", "v1.0.1", true, false},
		{"mixed prefixes", "1.0.0", "v1.0.1", true, false},
		{"pre-release vs stable", "1.0.0-beta", "1.0.0", true, false},
		{"invalid current version", "invalid", "1.0.0", false, true},
		{"invalid latest version", "1.0.0", "invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isUpdateAvailable(tt.current, tt.latest)
			if (err != nil) != tt.wantErr {
				t.Errorf("isUpdateAvailable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("isUpdateAvailable() = %v, want %v", got, tt.expected)
			}
		})
	}
}
