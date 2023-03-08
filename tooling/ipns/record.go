package ipns

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/gogo/protobuf/proto"
	ipns_pb "github.com/ipfs/go-ipns/pb"
)

type IPNSRecord struct {
	entry *ipns_pb.IpnsEntry
}

func (r *IPNSRecord) Verify(bytes []byte) bool {
	// TODO: see kubo/core/commands/name/name.go
	return false
}

func (r *IPNSRecord) Key() string {
	// TODO: see kubo/core/commands/name/name.go
	return string(r.entry.GetPubKey())
}

func (r *IPNSRecord) TTL() uint64 {
	return r.entry.GetTtl()
}

func newRecordFromPath(path string) (*IPNSRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var b bytes.Buffer

	_, err = io.Copy(&b, file)
	if err != nil {
		return nil, err
	}

	var entry ipns_pb.IpnsEntry
	err = proto.Unmarshal(b.Bytes(), &entry)
	if err != nil {
		return nil, err
	}

	return &IPNSRecord{entry: &entry}, nil
}

func MustOpenRecord(file string) *IPNSRecord {
	_, filename, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(filename)
	fixturePath := path.Join(basePath, "..", "..", "fixtures", file)

	record, err := newRecordFromPath(fixturePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return record
}
