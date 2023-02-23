package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipfs/go-unixfsnode/file"
	"github.com/ipld/go-car/v2/blockstore"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	basicnode "github.com/ipld/go-ipld-prime/node/basic"
)

var ErrNotDir = fmt.Errorf("not a directory")

func main() {
	if err := ExtractCar(os.Args[1]); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// ExtractCar pulls files and directories out of a car
func ExtractCar(file string) error {
	bs, err := blockstore.OpenReadOnly(file)
	if err != nil {
		return err
	}

	ls := cidlink.DefaultLinkSystem()
	ls.TrustedStorage = true
	ls.StorageReadOpener = func(_ ipld.LinkContext, l ipld.Link) (io.Reader, error) {
		cid, err := getCid(l)
		if err != nil {
			return nil, err
		}
		blk, err := bs.Get(context.Background(), cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(blk.RawData()), nil
	}

	roots, err := bs.Roots()
	if err != nil {
		return err
	}

	for _, root := range roots {
		if err := extractRoot(&ls, root); err != nil {
			return err
		}
	}

	return nil
}

func getCid(l ipld.Link) (cid.Cid, error) {
	cl, ok := l.(cidlink.Link)
	if !ok {
		return cid.Undef, fmt.Errorf("not a cidlink")
	}

	return cl.Cid, nil
}


func extractRoot(ls *ipld.LinkSystem, root cid.Cid) error {
	if root.Prefix().Codec == cid.Raw {
		fmt.Printf("skipping raw root %s\n", root)
		return nil
	}

	raw, err := ls.StorageReadOpener(ipld.LinkContext{}, cidlink.Link{Cid: root})
	if err != nil {
		return err
	}

	fmt.Printf("path: %s\n", "/")
	fmt.Printf("  cid: %s\n", root)
	fmt.Printf("  raw: %v\n", raw.(*bytes.Buffer).Bytes())

	pbn, err := ls.Load(ipld.LinkContext{}, cidlink.Link{Cid: root}, dagpb.Type.PBNode)
	if err != nil {
		return err
	}
	pbnode := pbn.(dagpb.PBNode)

	ufn, err := unixfsnode.Reify(ipld.LinkContext{}, pbnode, ls)
	if err != nil {
		return err
	}

	if err := extractDir(ls, ufn, "/"); err != nil {
		if !errors.Is(err, ErrNotDir) {
			return fmt.Errorf("%s: %w", root, err)
		}
		ufsData, err := pbnode.LookupByString("Data")
		if err != nil {
			return err
		}
		ufsBytes, err := ufsData.AsBytes()
		if err != nil {
			return err
		}
		ufsNode, err := data.DecodeUnixFSData(ufsBytes)
		if err != nil {
			return err
		}
		if ufsNode.DataType.Int() == data.Data_File || ufsNode.DataType.Int() == data.Data_Raw {
			if err := extractFile(ls, pbnode, "unknown"); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func extractDir(ls *ipld.LinkSystem, n ipld.Node, outputPath string) error {
	if n.Kind() == ipld.Kind_Map {
		mi := n.MapIterator()
		for !mi.Done() {
			key, val, err := mi.Next()
			if err != nil {
				return err
			}
			ks, err := key.AsString()
			if err != nil {
				return err
			}
			nextRes := path.Join(outputPath, ks)

			if val.Kind() != ipld.Kind_Link {
				return fmt.Errorf("unexpected map value for %s at %s", ks, outputPath)
			}
			// a directory may be represented as a map of name:<link> if unixADL is applied
			vl, err := val.AsLink()
			if err != nil {
				return err
			}

			cid, err := getCid(vl)
			if err != nil {
				return err
			}
			raw, err := ls.StorageReadOpener(ipld.LinkContext{}, cidlink.Link{Cid: cid})
			if err != nil {
				return err
			}

			fmt.Printf("path: %s\n", nextRes)
			fmt.Printf("  cid: %s\n", cid)
			fmt.Printf("  raw: %v\n", raw.(*bytes.Buffer).Bytes())

			dest, err := ls.Load(ipld.LinkContext{}, vl, basicnode.Prototype.Any)
			if err != nil {
				return err
			}
			// degenerate files are handled here.
			if dest.Kind() == ipld.Kind_Bytes {
				if err := extractFile(ls, dest, nextRes); err != nil {
					return err
				}
				continue
			} else {
				// dir / pbnode
				pbb := dagpb.Type.PBNode.NewBuilder()
				if err := pbb.AssignNode(dest); err != nil {
					return err
				}
				dest = pbb.Build()
			}
			pbnode := dest.(dagpb.PBNode)

			// interpret dagpb 'data' as unixfs data and look at type.
			ufsData, err := pbnode.LookupByString("Data")
			if err != nil {
				return err
			}
			ufsBytes, err := ufsData.AsBytes()
			if err != nil {
				return err
			}
			ufsNode, err := data.DecodeUnixFSData(ufsBytes)
			if err != nil {
				return err
			}
			if ufsNode.DataType.Int() == data.Data_Directory || ufsNode.DataType.Int() == data.Data_HAMTShard {
				ufn, err := unixfsnode.Reify(ipld.LinkContext{}, pbnode, ls)
				if err != nil {
					return err
				}

				if err := extractDir(ls, ufn, nextRes); err != nil {
					return err
				}
			} else if ufsNode.DataType.Int() == data.Data_File || ufsNode.DataType.Int() == data.Data_Raw {
				if err := extractFile(ls, pbnode, nextRes); err != nil {
					return err
				}
			} else if ufsNode.DataType.Int() == data.Data_Symlink {
				data := ufsNode.Data.Must().Bytes()
				if err := os.Symlink(string(data), nextRes); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return ErrNotDir
}

func extractFile(ls *ipld.LinkSystem, n ipld.Node, outputName string) error {
	node, err := file.NewUnixFSFile(context.Background(), n, ls)
	if err != nil {
		return err
	}
	nlr, err := node.AsLargeBytes()
	if err != nil {
		return err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, nlr)
	if err != nil {
		return err
	}

	fmt.Printf("  str: %s\n", buf.String())

	return nil
}
