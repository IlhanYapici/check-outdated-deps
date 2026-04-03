package npm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func GetPackageMetadata(pkgName string) (map[string]interface{}, error) {
	var data map[string]interface{}
	var err error

	res, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", pkgName))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch package metadata for %s: %w", pkgName, err)
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body for %s: %w", pkgName, err)
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse package metadata for %s: %w", pkgName, err)
	}

	return data, nil
}
