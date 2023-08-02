package tooling

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	Version = "dev"
)

func Home() string {
	home := os.Getenv("GATEWAY_CONFORMANCE_HOME")
	if home == "" {
		_, filename, _, _ := runtime.Caller(0)
		basePath := filepath.Dir(filename)
		return filepath.Join(basePath, "..")
	}
	return home
}
