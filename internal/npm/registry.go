package npm

type NpmPackageJSON struct {
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

type Package struct {
	Name    string
	Version string
}

type (
	Dependencies    []Package
	DevDependencies []Package
)
