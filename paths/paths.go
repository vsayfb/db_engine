package paths

import (
	"path/filepath"
	"runtime"
)

var (
	ProjectRoot string
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	ProjectRoot = filepath.Join(filepath.Dir(filename), "..")
}

func GetPath(relativePath string) string {
	return filepath.Join(ProjectRoot, relativePath)
}
