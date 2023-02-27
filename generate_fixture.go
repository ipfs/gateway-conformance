package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/ipfs/go-cid"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/ipld/go-car/v2/blockstore"
)

/**
 * list all `*.car` file in the basePath directory, recursively
 */
func listAllCarFile(basePath string) []string {
	var carFiles []string

	filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
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

	return carFiles
}

func GetAll(ctx context.Context, bs blockstore.ReadOnly) []cid.Cid {
	var cids []cid.Cid

	c, err := bs.AllKeysChan(ctx)
	if err != nil {
		panic(err)
	}

	for c := range c {
		cids = append(cids, c)
	}

	return cids
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	rout, err := blockstore.OpenReadWrite("./fixtures.car", []cid.Cid{})
	if err != nil {
		panic(err)
	}

	carFiles := listAllCarFile("./fixtures")
	for _, f := range carFiles {
		robs, err := blockstore.OpenReadOnly(f,
			blockstore.UseWholeCIDs(true),
			carv2.ZeroLengthSectionAsEOF(true),
		)

		if err != nil {
			panic(err)
		}

		cids, err := robs.AllKeysChan(ctx)
		if err != nil {
			panic(err)
		}

		for c := range cids {
			block, err := robs.Get(ctx, c)
			if err != nil {
				panic(err)
			}

			rout.Put(ctx, block)
		}
	}

	err = rout.Finalize()
	if err != nil {
		panic(err)
	}
}
