package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// ANSI color codes
const (
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Reset  = "\033[0m"
)

// VersionDiff represents the type of version difference
type VersionDiff int

const (
	NoDiff VersionDiff = iota
	PatchDiff
	MinorDiff
	MajorDiff
	CurrentIsLatest
	LatestIsRC
)

func SanitizeVersion(version string) string {
	version = strings.TrimSpace(version)

	prefixes := []string{"^", "~", ">=", "<=", ">", "<", "="}

	for _, prefix := range prefixes {
		if strings.HasPrefix(version, prefix) {
			version = strings.TrimPrefix(version, prefix)
			break
		}
	}

	return strings.TrimSpace(version)
}

// isReleaseCandidate checks if a version is a release candidate
func isReleaseCandidate(version string) bool {
	version = strings.ToLower(version)
	return strings.Contains(version, "-rc") ||
		strings.Contains(version, ".rc") ||
		strings.Contains(version, "-alpha") ||
		strings.Contains(version, "-beta")
}

// parseVersion parses a semantic version string (e.g., "1.2.3")
func parseVersion(version string) (major, minor, patch int, err error) {
	// Remove leading and trailing spaces
	version = strings.TrimSpace(version)
	// Remove 'v' prefix if present
	version = strings.TrimPrefix(version, "v")

	// Remove pre-release identifiers for parsing
	if idx := strings.Index(version, "-"); idx != -1 {
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
	}

	patch, err = strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
	}

	return major, minor, patch, nil
}

// getVersionDiff determines the highest level of difference between two versions
func getVersionDiff(currentVersion, latestVersion string) (VersionDiff, error) {
	// Check if latest is a release candidate first
	if isReleaseCandidate(latestVersion) {
		return LatestIsRC, nil
	}

	currMajor, currMinor, currPatch, err := parseVersion(currentVersion)
	if err != nil {
		return NoDiff, fmt.Errorf("error parsing current version: %w", err)
	}

	latestMajor, latestMinor, latestPatch, err := parseVersion(latestVersion)
	if err != nil {
		return NoDiff, fmt.Errorf("error parsing latest version: %w", err)
	}

	// Check if current is same as latest
	if currMajor == latestMajor && currMinor == latestMinor && currPatch == latestPatch {
		return CurrentIsLatest, nil
	}

	if currMajor != latestMajor {
		return MajorDiff, nil
	}
	if currMinor != latestMinor {
		return MinorDiff, nil
	}
	if currPatch != latestPatch {
		return PatchDiff, nil
	}

	return NoDiff, nil
}

// formatVersionWithHighlight formats the latest version with color highlighting
// based on the difference type
func formatVersionWithHighlight(latestVersion string, diff VersionDiff) (string, error) {
	switch diff {
	case CurrentIsLatest:
		// Highlight entire latest version in green (same as current)
		major, minor, patch, err := parseVersion(latestVersion)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s%d.%d.%d%s (up to date)", Green, major, minor, patch, Reset), nil
	case LatestIsRC:
		// No highlighting for release candidates
		return fmt.Sprintf("%s%s%s (pre-release)", Yellow, latestVersion, Reset), nil
	case PatchDiff:
		// Highlight only patch in green
		major, minor, patch, err := parseVersion(latestVersion)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d.%d.%s%d%s", major, minor, Green, patch, Reset), nil
	case MinorDiff:
		// Highlight minor and patch in yellow
		major, minor, patch, err := parseVersion(latestVersion)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d.%s%d.%d%s", major, Yellow, minor, patch, Reset), nil
	case MajorDiff:
		// Highlight entire version in red
		major, minor, patch, err := parseVersion(latestVersion)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s%d.%d.%d%s", Red, major, minor, patch, Reset), nil
	default:
		// No difference, return latest as-is
		return latestVersion, nil
	}
}

// FormatVersionComparison is the main function that compares versions and returns
// only the formatted latest version with appropriate color highlighting
func FormatVersionComparison(currentVersion, latestVersion string) (string, error) {
	diff, err := getVersionDiff(currentVersion, latestVersion)
	if err != nil {
		return "", err
	}

	return formatVersionWithHighlight(latestVersion, diff)
}
