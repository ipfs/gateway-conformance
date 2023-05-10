/**
 * This was copied from Kubo's `name` command
 * TODO: discuss with the team about setting up a reusable abstraction somewhere.
 **/
package ipns

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/boxo/ipns"
	ipns_pb "github.com/ipfs/boxo/ipns/pb"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/libp2p/go-libp2p/core/peer"
	mbase "github.com/multiformats/go-multibase"
)

// IpnsInspectEntry contains the deserialized values from an IPNS Entry:
// https://github.com/ipfs/specs/blob/main/ipns/IPNS.md#record-serialization-format
type IpnsInspectEntry struct {
	Value        string                          `json:"value"`
	ValidityType *ipns_pb.IpnsEntry_ValidityType `json:"validityType"`
	Validity     *time.Time                      `json:"validity"`
	Sequence     uint64                          `json:"sequence"`
	TTL          *uint64                         `json:"ttl"`
	PublicKey    string                          `json:"publicKey"`
	SignatureV1  string                          `json:"signatureV1"`
	SignatureV2  string                          `json:"signatureV2"`
	Data         interface{}                     `json:"data"`
}

func unmarshalIPNSEntry(data []byte) (*ipns_pb.IpnsEntry, error) {
	var entry ipns_pb.IpnsEntry
	err := proto.Unmarshal(data, &entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func unmarshalIPNSRecord(entry *ipns_pb.IpnsEntry) (*IpnsInspectEntry, error) {
	encoder, err := mbase.EncoderByName("base64")
	if err != nil {
		return nil, err
	}

	result := IpnsInspectEntry{
		Value:        string(entry.Value),
		ValidityType: entry.ValidityType,
		Sequence:     *entry.Sequence,
		TTL:          entry.Ttl,
		PublicKey:    encoder.Encode(entry.PubKey),
		SignatureV1:  encoder.Encode(entry.SignatureV1),
		SignatureV2:  encoder.Encode(entry.SignatureV2),
		Data:         nil,
	}

	if len(entry.Data) != 0 {
		// This is hacky. The variable node (datamodel.Node) doesn't directly marshal
		// to JSON. Therefore, we need to first decode from DAG-CBOR, then encode in
		// DAG-JSON and finally unmarshal it from JSON. Since DAG-JSON is a subset
		// of JSON, that should work. Then, we can store the final value in the
		// result.Entry.Data for further inspection.
		node, err := ipld.Decode(entry.Data, dagcbor.Decode)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		err = dagjson.Encode(node, &buf)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(buf.Bytes(), &result.Data)
		if err != nil {
			return nil, err
		}
	}

	validity, err := ipns.GetEOL(entry)
	if err == nil {
		result.Validity = &validity
	}

	return &result, nil
}

type IpnsInspectValidation struct {
	Valid     bool
	Reason    string
	PublicKey peer.ID
}
