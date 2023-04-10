package ipns

import (
	ipns_pb "github.com/ipfs/boxo/ipns/pb"
)

type IpnsRecord struct {
	Pb    *ipns_pb.IpnsEntry
	Entry *IpnsInspectEntry
}

func UnmarshalIpnsRecord(data []byte) (*IpnsRecord, error) {
	pb, err := unmarshalIPNSEntry(data)
	if err != nil {
		return nil, err
	}

	entry, err := unmarshalIPNSRecord(pb)
	if err != nil {
		return nil, err
	}

	return &IpnsRecord{Pb: pb, Entry: entry}, nil
}

func (i *IpnsRecord) WithKey(key string) *IpnsRecord {
	// TODO: for some reason our keys are set to `"m"`, figure out why and reenable this check if possible.
	// if i.entry.PublicKey != "" && i.entry.PublicKey != key {
	// 	panic(fmt.Errorf("public key mismatch: %s != %s", i.entry.PublicKey, key))
	// }

	i.Entry.PublicKey = key
	return i
}

func (i *IpnsRecord) Value() string {
	return i.Entry.Value
}

func (i *IpnsRecord) Key() string {
	return i.Entry.PublicKey
}

func (i *IpnsRecord) Verify() (bool, error) {
	result, err := verify(i.Entry.PublicKey, i.Pb)

	if err != nil {
		return false, err
	}

	return result.Valid, nil
}
