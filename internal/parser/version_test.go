package parser

import (
	"testing"
)

func TestSanitizeVersion(t *testing.T) {
	tests := []struct {
		name, version, expected string
	}{
		{"no prefix", "1.2.3", "1.2.3"},
		{"with prefix ^", "^1.2.3", "1.2.3"},
		{"with prefix ~", "~1.2.3", "1.2.3"},
		{"with prefix >=", ">=1.2.3", "1.2.3"},
		{"with prefix <=", "<=1.2.3", "1.2.3"},
		{"with prefix >", ">1.2.3", "1.2.3"},
		{"with prefix <", "<1.2.3", "1.2.3"},
		{"with prefix =", "=1.2.3", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeVersion(tt.version)
			if result != tt.expected {
				t.Errorf("isReleaseCandidate(%s) = %s; expected %s", tt.version, result, tt.expected)
			}
		})
	}
}

func TestIsRealeseCandidate(t *testing.T) {
	tests := []struct {
		name, version string
		expected      bool
	}{
		{"not release candidate", "19.2.312", false},
		{"is release candidate (-rc)", "19.2.312-rc", true},
		{"is release candidate (.rc)", "19.2.312.rc", true},
		{"is alpha release", "19.2.312-alpha", true},
		{"is beta release", "19.2.312-beta", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isReleaseCandidate(tt.version)
			if result != tt.expected {
				t.Errorf("isReleaseCandidate(%s) = %t; expected %t", tt.version, result, tt.expected)
			}
		})
	}
}

func TestParseVersion_ValidVersions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantMaj int
		wantMin int
		wantPat int
	}{
		{"simple", "1.2.3", 1, 2, 3},
		{"with v prefix", "v10.20.30", 10, 20, 30},
		{"with pre-release -rc", "2.4.6-rc1", 2, 4, 6},
		{"with pre-release -alpha", "3.5.7-alpha", 3, 5, 7},
		{"with pre-release -beta", "4.6.8-beta.2", 4, 6, 8},
		{"with spaces", " 5.7.9 ", 5, 7, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maj, min, pat, err := parseVersion(tt.input)
			if err != nil {
				t.Errorf("parseVersion(%q) returned error: %v", tt.input, err)
			}
			if maj != tt.wantMaj || min != tt.wantMin || pat != tt.wantPat {
				t.Errorf("parseVersion(%q) = %d,%d,%d; want %d,%d,%d", tt.input, maj, min, pat, tt.wantMaj, tt.wantMin, tt.wantPat)
			}
		})
	}
}

func TestParseVersion_InvalidVersions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"too few parts", "1.2"},
		{"too many parts", "1.2.3.4"},
		{"non-numeric major", "x.2.3"},
		{"non-numeric minor", "1.y.3"},
		{"non-numeric patch", "1.2.z"},
		{"empty string", ""},
		{"only pre-release", "-rc"},
		{"random string", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, _, err := parseVersion(tt.input)
			if err == nil {
				t.Errorf("parseVersion(%q) expected error, got nil", tt.input)
			}
		})
	}
}

func TestGetVersionDiff(t *testing.T) {
	tests := []struct {
		name, current, latest string
		expected              VersionDiff
	}{
		{"latest is RC", "1.2.3", "1.2.3-rc", LatestIsRC},
		{"current is latest", "1.2.3", "1.2.3", CurrentIsLatest},
		{"major is diff", "1.2.3", "2.0.0", MajorDiff},
		{"minor is diff", "1.2.3", "1.3.0", MinorDiff},
		{"patch is diff", "1.2.3", "1.2.4", PatchDiff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := getVersionDiff(tt.current, tt.latest)
			if result != tt.expected {
				t.Errorf("isReleaseCandidate(%s, %s) = %d; expected %d", tt.current, tt.latest, result, tt.expected)
			}
		})
	}
}
