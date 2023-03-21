package fixtures

import (
	"os"
	"path"
	"path/filepath"

	"github.com/ipfs/gateway-conformance/tooling"
)

func Dir() string {
	home := tooling.Home()
	return path.Join(home, "fixtures")
}

func List() ([]string, error) {
	var carFiles []string

	err := filepath.WalkDir(Dir(), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".car" {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			carFiles = append(carFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return carFiles, nil
}
