package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ipfs/boxo/ipns"
	ipns_pb "github.com/ipfs/boxo/ipns/pb"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-test/random"
	ic "github.com/libp2p/go-libp2p/core/crypto"
	"google.golang.org/protobuf/proto"
)

func makeRawPath(str string) path.Path {
	prefix := cid.Prefix{
		Version: 1,
		Codec:   0x55,
	}

	cid, err := prefix.Sum([]byte(str))
	panicOnErr(err)

	return path.FromCid(cid)
}

var (
	seq = uint64(0)
	eol = time.Now().Add(time.Hour * 876000) // 100 years
	ttl = time.Minute * 30
)

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func makeKeyPair() (ic.PrivKey, ic.PubKey, ipns.Name) {
	pid, sk, pk := random.Identity()
	return sk, pk, ipns.NameFromPeer(pid)
}

func saveToFile(data []byte, filename string, value path.Path) {
	err := os.WriteFile(filename, data, 0666)
	panicOnErr(err)
	fmt.Printf("%s -> %s\n", filename, value.String())
}

func makeV1Only() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v1-only record")

	// Create working record
	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(true))
	panicOnErr(err)

	// Marshal
	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	// Unmarshal into raw structure
	pb := ipns_pb.IpnsRecord{}
	err = proto.Unmarshal(raw, &pb)
	panicOnErr(err)

	// Make it V1-only
	pb.Data = nil
	pb.SignatureV2 = nil

	// Marshal again and store it
	raw, err = proto.Marshal(&pb)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v1.ipns-record", v)
}

func makeV1V2() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v1+v2 record")

	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(true))
	panicOnErr(err)

	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v1-v2.ipns-record", v)
}

func makeV1V2WithBrokenValue() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v1+v2 record with broken value")

	// Create working record
	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(true))
	panicOnErr(err)

	// Marshal
	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	// Unmarshal into raw structure
	pb := ipns_pb.IpnsRecord{}
	err = proto.Unmarshal(raw, &pb)
	panicOnErr(err)

	// Make Value different
	pb.Value = []byte("/ipfs/bafkqaglumvzxi2lom4qgeyleebuxa3ttebzgky3pojshgcq")

	// Marshal again and store it
	raw, err = proto.Marshal(&pb)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v1-v2-broken-v1-value.ipns-record", v)
}

func makeV1V2WithBrokenSignatureV1() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v1+v2 with broken signature v1")

	// Create working record
	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(true))
	panicOnErr(err)

	// Marshal
	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	// Unmarshal into raw structure
	pb := ipns_pb.IpnsRecord{}
	err = proto.Unmarshal(raw, &pb)
	panicOnErr(err)

	// Break Signature V1
	pb.SignatureV1 = []byte("invalid stuff")

	// Marshal again and store it
	raw, err = proto.Marshal(&pb)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v1-v2-broken-signature-v1.ipns-record", v)
}

func makeV1V2WithBrokenSignatureV2() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v1+v2 with broken signature v2")

	// Create working record
	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(true))
	panicOnErr(err)

	// Marshal
	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	// Unmarshal into raw structure
	pb := ipns_pb.IpnsRecord{}
	err = proto.Unmarshal(raw, &pb)
	panicOnErr(err)

	// Break Signature V2
	pb.SignatureV2 = []byte("invalid stuff")

	// Marshal again and store it
	raw, err = proto.Marshal(&pb)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v1-v2-broken-signature-v2.ipns-record", v)
}

func makeV2Only() {
	sk, _, name := makeKeyPair()

	v := makeRawPath("v2-only record")

	rec, err := ipns.NewRecord(sk, v, seq, eol, ttl, ipns.WithV1Compatibility(false))
	panicOnErr(err)

	raw, err := ipns.MarshalRecord(rec)
	panicOnErr(err)

	saveToFile(raw, name.String()+"_v2.ipns-record", v)
}

func main() {
	makeV1Only()
	makeV1V2()
	makeV1V2WithBrokenValue()
	makeV1V2WithBrokenSignatureV1()
	makeV1V2WithBrokenSignatureV2()
	makeV2Only()
}
