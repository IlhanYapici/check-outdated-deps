package npm

type PackageManager string

const (
	NPM  PackageManager = "npm"
	PNPM PackageManager = "pnpm"
	YARN PackageManager = "yarn"
)

type NpmPackageJson struct {
	PackageManager  string            `json:"packageManager,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

type Package struct {
	Name    string
	Version string
}

type Dependencies []Package
type DevDependencies []Package
