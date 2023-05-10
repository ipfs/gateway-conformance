package ipns

import (
	"time"

	"github.com/ipfs/boxo/ipns"
	ipns_pb "github.com/ipfs/boxo/ipns/pb"
	"github.com/libp2p/go-libp2p/core/peer"
)

type IpnsRecord struct {
	pb  *ipns_pb.IpnsEntry
	key string
	validity time.Time
}

func UnmarshalIpnsRecord(data []byte, pubKey string) (*IpnsRecord, error) {
	pb, err := unmarshalIPNSEntry(data)
	if err != nil {
		return nil, err
	}

	validity, err := ipns.GetEOL(pb)
	if err != nil {
		return nil, err
	}

	return &IpnsRecord{pb: pb, key: pubKey, validity: validity}, nil
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
	id, err := peer.Decode(i.key)
	if err != nil {
		return err
	}

	return ipns.ValidateWithPeerID(id, i.pb)
}
