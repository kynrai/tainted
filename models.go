package chooser

const (
	gitCommand = "git"
	goCommand  = "go"
)

// Package represents a code package with its imports and path to import
type Package struct {
	ImportPath string
	RootDir    string
	Imports    []string
}
