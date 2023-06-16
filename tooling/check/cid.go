package check

import (
	"bytes"
	"fmt"
	"io"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car"
	"github.com/ipld/go-car/v2/blockstore"
)

func CidSetContains(a, b []cid.Cid) CheckOutput {
	s1 := cid.NewSet()
	for _, cid := range a {
		s1.Add(cid)
	}

	// for each cid in b, check if it's in a
	for _, cid := range b {
		if !s1.Has(cid) {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("missing CID %s", cid),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}

func CidSetEquals(a, b []cid.Cid) CheckOutput {
	t1 := CidSetContains(a, b)

	if !t1.Success {
		return t1
	}

	return CidSetContains(b, a)
}

func CidArrayEquals(a, b []cid.Cid) CheckOutput {
	if len(a) != len(b) {
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("length mismatch: %d != %d", len(a), len(b)),
		}
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("mismatch at index %d: %s != %s", i, a[i], b[i]),
			}
		}
	}

	return CheckOutput{
		Success: true,
	}
}

func CidOrderedSubsetContains(a, b []cid.Cid) CheckOutput {
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			i++
			j++
			continue
		}

		i++
	}

	if j != len(b) {
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("missing CID %s", b[j]),
		}
	}

	return CheckOutput{
		Success: true,
	}
}

func listAllCids(carContent []byte) ([]cid.Cid, error) {
	reader := bytes.NewReader(carContent)
	cr, err := car.NewCarReader(reader)
	if err != nil {
		return nil, err
	}

	// aggregate all blocks, ordered
	var gotCIDs []cid.Cid
	for {
		block, err := cr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		gotCIDs = append(gotCIDs, block.Cid())
	}

	return gotCIDs, nil
}

func listAllRoots(carContent []byte) ([]cid.Cid, error) {
	reader := bytes.NewReader(carContent)
	bs, err := blockstore.NewReadOnly(reader, nil)

	if err != nil {
		return nil, err
	}

	return bs.Roots()
}
