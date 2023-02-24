package car

import (
	"bytes"
	"context"
	"fmt"
	"io"
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

var errNotDir = fmt.Errorf("not a directory")

type node struct {
	Path   string
	Cid    cid.Cid
	Raw    []byte
	String string
}

func findNode(nodes []node, p string) *node {
	for _, node := range nodes {
		if node.Path == p {
			return &node
		}
	}
	return nil
}

// extractCar pulls files and directories out of a car
func extractCar(file string) ([]node, error) {
	ns, err := blockstore.OpenReadOnly(file)
	if err != nil {
		return nil, err
	}

	ls := cidlink.DefaultLinkSystem()
	ls.TrustedStorage = true
	ls.StorageReadOpener = func(_ ipld.LinkContext, l ipld.Link) (io.Reader, error) {
		cid, err := getCid(l)
		if err != nil {
			return nil, err
		}
		blk, err := ns.Get(context.Background(), cid)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(blk.RawData()), nil
	}

	roots, err := ns.Roots()
	if err != nil {
		return nil, err
	}

	nodes := []node{}
	for _, root := range roots {
		ns, err := extractRoot(&ls, root)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, ns...)
	}

	return nodes, nil
}

func getCid(l ipld.Link) (cid.Cid, error) {
	cl, ok := l.(cidlink.Link)
	if !ok {
		return cid.Undef, fmt.Errorf("not a cidlink")
	}

	return cl.Cid, nil
}

func extractRoot(ls *ipld.LinkSystem, root cid.Cid) ([]node, error) {
	if root.Prefix().Codec == cid.Raw {
		fmt.Printf("skipping raw root %s\n", root)
		return nil, nil
	}

	raw, err := ls.StorageReadOpener(ipld.LinkContext{}, cidlink.Link{Cid: root})
	if err != nil {
		return nil, err
	}

	pbn, err := ls.Load(ipld.LinkContext{}, cidlink.Link{Cid: root}, dagpb.Type.PBNode)
	if err != nil {
		return nil, err
	}
	pbnode := pbn.(dagpb.PBNode)

	ufn, err := unixfsnode.Reify(ipld.LinkContext{}, pbnode, ls)
	if err != nil {
		return nil, err
	}

	nodes := []node{
		{
			Path:   "/",
			Cid:    root,
			Raw:    raw.(*bytes.Buffer).Bytes(),
			String: "",
		},
	}

	ns, err := extractDir(ls, ufn, "/")
	if err != nil {
		return nil, err
	}
	nodes = append(nodes, ns...)

	return nodes, nil
}

func extractDir(ls *ipld.LinkSystem, n ipld.Node, outputPath string) ([]node, error) {
	if n.Kind() == ipld.Kind_Map {
		mi := n.MapIterator()
		nodes := []node{}
		for !mi.Done() {
			key, val, err := mi.Next()
			if err != nil {
				return nil, err
			}
			ks, err := key.AsString()
			if err != nil {
				return nil, err
			}
			nextRes := path.Join(outputPath, ks)

			if val.Kind() != ipld.Kind_Link {
				return nil, fmt.Errorf("unexpected map value for %s at %s", ks, outputPath)
			}
			// a directory may be represented as a map of name:<link> if unixADL is applied
			vl, err := val.AsLink()
			if err != nil {
				return nil, err
			}

			cid, err := getCid(vl)
			if err != nil {
				return nil, err
			}
			raw, err := ls.StorageReadOpener(ipld.LinkContext{}, cidlink.Link{Cid: cid})
			if err != nil {
				return nil, err
			}

			dest, err := ls.Load(ipld.LinkContext{}, vl, basicnode.Prototype.Any)
			if err != nil {
				return nil, err
			}

			// degenerate files are handled here.
			if dest.Kind() == ipld.Kind_Bytes {
				str, err := extractFile(ls, dest, nextRes)
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, node{
					Path:   nextRes,
					Cid:    cid,
					Raw:    raw.(*bytes.Buffer).Bytes(),
					String: str,
				})
				continue
			} else {
				// dir / pbnode
				pbb := dagpb.Type.PBNode.NewBuilder()
				if err := pbb.AssignNode(dest); err != nil {
					return nil, err
				}
				dest = pbb.Build()
			}
			pbnode := dest.(dagpb.PBNode)

			// interpret dagpb 'data' as unixfs data and look at type.
			ufsData, err := pbnode.LookupByString("Data")
			if err != nil {
				return nil, err
			}
			ufsBytes, err := ufsData.AsBytes()
			if err != nil {
				return nil, err
			}
			ufsNode, err := data.DecodeUnixFSData(ufsBytes)
			if err != nil {
				return nil, err
			}
			if ufsNode.DataType.Int() == data.Data_Directory || ufsNode.DataType.Int() == data.Data_HAMTShard {
				ufn, err := unixfsnode.Reify(ipld.LinkContext{}, pbnode, ls)
				if err != nil {
					return nil, err
				}

				nodes = append(nodes, node{
					Path:   nextRes,
					Cid:    cid,
					Raw:    raw.(*bytes.Buffer).Bytes(),
					String: "",
				})

				ns, err := extractDir(ls, ufn, nextRes)
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, ns...)
			} else if ufsNode.DataType.Int() == data.Data_File || ufsNode.DataType.Int() == data.Data_Raw {
				str, err := extractFile(ls, pbnode, nextRes)
				if err != nil {
					return nil, err
				}
				nodes = append(nodes, node{
					Path:   nextRes,
					Cid:    cid,
					Raw:    raw.(*bytes.Buffer).Bytes(),
					String: str,
				})
			} else {
				return nil, fmt.Errorf("unknown unixfs type: %d", ufsNode.DataType.Int())
			}
		}
		return nodes, nil
	}
	return nil, errNotDir
}

func extractFile(ls *ipld.LinkSystem, n ipld.Node, outputName string) (string, error) {
	node, err := file.NewUnixFSFile(context.Background(), n, ls)
	if err != nil {
		return "", err
	}
	nlr, err := node.AsLargeBytes()
	if err != nil {
		return "", err
	}

	buf := new(strings.Builder)
	_, err = io.Copy(buf, nlr)
	if err != nil {
		return "", err
	}

	str := buf.String()
	return str, nil
}
