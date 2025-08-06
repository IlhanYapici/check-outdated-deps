package config

import (
	"check-outdated-deps/pkg/npm"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func LoadPackageJson(filename string) (*npm.NpmPackageJson, error) {
	var parsedFile npm.NpmPackageJson

	jsonFile, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &parsedFile)

	return &parsedFile, nil
}

// DetectPackageManagerFromLockfiles detects package manager based on lockfiles in the given directory
func DetectPackageManagerFromLockfiles(dir string) (npm.PackageManager, error) {
	lockfiles := map[string]npm.PackageManager{
		"pnpm-lock.yaml":    npm.PNPM,
		"yarn.lock":         npm.YARN,
		"package-lock.json": npm.NPM,
	}

	for lockfile, manager := range lockfiles {
		lockfilePath := filepath.Join(dir, lockfile)
		if _, err := os.Stat(lockfilePath); err == nil {
			return manager, nil
		}
	}

	// No lockfile found, default to npm
	return npm.NPM, nil
}

// GetPackageManagerFromString parses package manager string (like "npm@8.19.2")
func GetPackageManagerFromString(pkgManager string) (npm.PackageManager, error) {
	if pkgManager == "" {
		return "", fmt.Errorf("empty package manager string")
	}

	if strings.HasPrefix(pkgManager, "npm") {
		return npm.NPM, nil
	} else if strings.HasPrefix(pkgManager, "pnpm") {
		return npm.PNPM, nil
	} else if strings.HasPrefix(pkgManager, "yarn") {
		return npm.YARN, nil
	} else {
		return "", fmt.Errorf("unknown package manager: %s", pkgManager)
	}
}

// DeterminePackageManager uses hybrid approach: packageManager field first, then lockfile detection
func DeterminePackageManager(packageJsonPath string, parsedPackageJson *npm.NpmPackageJson) (npm.PackageManager, error) {
	// First, try to get package manager from package.json
	if parsedPackageJson.PackageManager != "" {
		manager, err := GetPackageManagerFromString(parsedPackageJson.PackageManager)

		if err == nil {
			return manager, nil
		}

		// If parsing fails, log the issue but don't fail completely
		fmt.Printf("Warning: failed to parse packageManager field '%s': %v\n", parsedPackageJson.PackageManager, err)
	}

	// Fall back to lockfile detection
	dir := filepath.Dir(packageJsonPath)
	manager, err := DetectPackageManagerFromLockfiles(dir)
	if err != nil {
		return "", fmt.Errorf("failed to detect package manager from lockfiles: %w", err)
	}

	fmt.Printf("Package manager detected from lockfiles: %s\n", manager)
	return manager, nil
}

func CheckPkgManagerExists(pkgManager string) bool {
	_, err := exec.LookPath(pkgManager)

	return err == nil
}
