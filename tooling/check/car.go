package check

import (
	"fmt"

	"github.com/ipfs/go-cid"
)

type CheckIsCarFile struct {
	blockCIDs []cid.Cid
	rootCIDs  []cid.Cid
	isExact   bool
	isOrdered bool
}

func IsCar() *CheckIsCarFile {
	return &CheckIsCarFile{
		blockCIDs: []cid.Cid{},
		rootCIDs:  []cid.Cid{},
		isExact:   false,
		isOrdered: false,
	}
}

func decoded(cidStr string) cid.Cid {
	cid, err := cid.Decode(cidStr)
	if err != nil {
		panic(fmt.Errorf("invalid CID: %w", err))
	}
	return cid
}

func (c CheckIsCarFile) HasBlock(cidStr string) CheckIsCarFile {
	c.blockCIDs = append(c.blockCIDs, decoded(cidStr))
	return c
}

func (c CheckIsCarFile) HasBlocks(cidStrs ...string) CheckIsCarFile {
	for _, cidStr := range cidStrs {
		c.blockCIDs = append(c.blockCIDs, decoded(cidStr))
	}
	return c
}

func (c CheckIsCarFile) HasRoot(cidStr string) CheckIsCarFile {
	c.rootCIDs = append(c.rootCIDs, decoded(cidStr))
	return c
}

func (c CheckIsCarFile) HasRoots(cidStrs ...string) CheckIsCarFile {
	for _, cidStr := range cidStrs {
		c.rootCIDs = append(c.rootCIDs, decoded(cidStr))
	}
	return c
}

func (c CheckIsCarFile) Exactly() CheckIsCarFile {
	c.isExact = true
	return c
}

func (c CheckIsCarFile) InThatOrder() CheckIsCarFile {
	c.isOrdered = true
	return c
}

func (c *CheckIsCarFile) Check(carContent []byte) CheckOutput {
	gotCIDs, err := listAllCids(carContent)
	if err != nil {
		return CheckOutput{
			Success: false,
			Reason:  fmt.Sprintf("failed to list all cids: %v", err),
		}
	}

	cmp := CidSetContains

	if c.isExact {
		if c.isOrdered {
			cmp = CidArrayEquals
		} else {
			cmp = CidSetEquals
		}
	} else {
		if c.isOrdered {
			cmp = CidOrderedSubsetContains
		}
	}

	output := cmp(gotCIDs, c.blockCIDs)

	if !output.Success {
		return output
	}

	if len(c.rootCIDs) > 0 || c.isExact {
		gotRoots, err := listAllRoots(carContent)
		if err != nil {
			return CheckOutput{
				Success: false,
				Reason:  fmt.Sprintf("failed to list all roots: %v", err),
			}
		}

		output = cmp(gotRoots, c.rootCIDs)

		if !output.Success {
			return output
		}
	}

	return CheckOutput{
		Success: true,
	}
}
