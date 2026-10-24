package kanjiconv

import "testing"

func TestGoVersionWarning(t *testing.T) {
	tests := []struct {
		version string
		warn    bool
	}{
		{"go1.27.3", false},
		{"go1.26.10", false},
		{"go1.27.4", false},
		{"go1.26.11", false},
		{"go1.27.2", true},
		{"go1.27.0", true},
		{"go1.26.9", true},
		{"go1.25.12", true},
		{"go1.24.1", true},
		{"go1.28.0", true},
		{"go1.27rc1", true},
		{"devel go1.28-abcdef", true},
		{"go1.27.3 X:boringcrypto", true},
	}
	for _, tt := range tests {
		got := goVersionWarning(tt.version)
		if (got != "") != tt.warn {
			t.Errorf("goVersionWarning(%q) = %q, want warning: %v", tt.version, got, tt.warn)
		}
	}
}
