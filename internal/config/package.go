package config

import (
	"encoding/json"
	"io"
	"os"

	"check-outdated-deps/internal/npm"
)

func LoadPackageJSON(filename string) (*npm.NpmPackageJSON, error) {
	var parsedFile npm.NpmPackageJSON

	jsonFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	byteValue, _ := io.ReadAll(jsonFile)

	json.Unmarshal(byteValue, &parsedFile)

	return &parsedFile, nil
}
