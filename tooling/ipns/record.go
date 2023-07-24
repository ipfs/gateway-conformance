package ipns

import (
	"strings"
	"time"

	"github.com/ipfs/boxo/ipns"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	mbase "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
)

type IpnsRecord struct {
	rec      *ipns.Record
	key      string
	value    string
	name     ipns.Name
	validity time.Time
}

func UnmarshalIpnsRecord(data []byte, pubKey string) (*IpnsRecord, error) {
	pb, err := ipns.UnmarshalRecord(data)
	if err != nil {
		return nil, err
	}

	validity, err := pb.Validity()
	if err != nil {
		return nil, err
	}

	value, err := pb.Value()
	if err != nil {
		return nil, err
	}

	id, err := peer.Decode(pubKey)
	if err != nil {
		return nil, err
	}

	return &IpnsRecord{
		rec:      pb,
		key:      pubKey,
		name:     ipns.NameFromPeer(id),
		validity: validity,
		value:    value.String(),
	}, nil
}

func (i *IpnsRecord) Value() string {
	return i.value
}

func (i *IpnsRecord) Key() string {
	return i.key
}

func (i *IpnsRecord) Validity() time.Time {
	return i.validity
}

func (i *IpnsRecord) Valid() error {
	return ipns.ValidateWithName(i.rec, i.name)
}

func (i *IpnsRecord) idV1(codec multicodec.Code, base mbase.Encoding) (string, error) {
	c := i.name.Cid()
	c = cid.NewCidV1(uint64(codec), c.Hash())
	s, err := c.StringOfBase(base)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (i *IpnsRecord) ToCID(codec multicodec.Code, base mbase.Encoding) string {
	s, err := i.idV1(codec, base)
	if err != nil {
		panic(err)
	}
	return s
}

func (i *IpnsRecord) IdV0() string {
	if strings.HasPrefix(i.key, "Qm") || strings.HasPrefix(i.key, "1") {
		return i.key
	}

	panic("not a v0 id")
}

func (i *IpnsRecord) IdV1() string {
	return i.ToCID(cid.Libp2pKey, mbase.Base36)
}

func (i *IpnsRecord) B58MH() string {
	return i.name.Peer().String()
}
