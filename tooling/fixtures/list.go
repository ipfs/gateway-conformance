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

type Fixtures struct {
	CarFiles    []string
	ConfigFiles []string
	IPNSRecords []string
}

func List() (*Fixtures, error) {
	var carFiles []string
	var yamlFiles []string
	var ipnsRecords []string

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
		// if we have a yaml file, append:
		if filepath.Ext(path) == ".yml" {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			yamlFiles = append(yamlFiles, path)
		}
		// if we have an ipns-record file, append:
		if filepath.Ext(path) == ".ipns-record" {
			path, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			ipnsRecords = append(ipnsRecords, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Fixtures{
		CarFiles:    carFiles,
		ConfigFiles: yamlFiles,
		IPNSRecords: ipnsRecords,
	}, nil
}
