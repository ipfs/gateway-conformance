package ipns

import (
	"strings"
	"time"

	"github.com/ipfs/boxo/ipns"
	ipns_pb "github.com/ipfs/boxo/ipns/pb"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	mbase "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
)

type IpnsRecord struct {
	pb       *ipns_pb.IpnsEntry
	key      string
	id       peer.ID
	validity time.Time
}

func UnmarshalIpnsRecord(data []byte, pubKey string) (*IpnsRecord, error) {
	pb, err := ipns.UnmarshalIpnsEntry(data)
	if err != nil {
		return nil, err
	}

	validity, err := ipns.GetEOL(pb)
	if err != nil {
		return nil, err
	}

	id, err := peer.Decode(pubKey)
	if err != nil {
		return nil, err
	}

	return &IpnsRecord{pb: pb, key: pubKey, id: id, validity: validity}, nil
}

func (i *IpnsRecord) Value() string {
	return string(i.pb.Value)
}

func (i *IpnsRecord) Key() string {
	return i.key
}

func (i *IpnsRecord) Validity() time.Time {
	return i.validity
}

func (i *IpnsRecord) Valid() error {
	return ipns.ValidateWithPeerID(i.id, i.pb)
}

func (i *IpnsRecord) idV1(codec multicodec.Code, base mbase.Encoding) (string, error) {
	c := peer.ToCid(i.id)
	c = cid.NewCidV1(uint64(codec), c.Hash())
	s, err := c.StringOfBase(base)
	if err != nil {
		return "", err
	}
	return s, nil
}

func (i *IpnsRecord) IntoCID(codec multicodec.Code, base mbase.Encoding) string {
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
	return i.IntoCID(cid.Libp2pKey, mbase.Base36)
}

func (i *IpnsRecord) B58MH() string {
	return i.id.String()
}
