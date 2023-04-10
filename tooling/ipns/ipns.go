package ipns

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/ipfs/gateway-conformance/tooling/fixtures"
)

/**
 * Extracts the public key from the path of an IPNS record.
 * The path is expected to be in the format of:
 * some/path/then/[pubkey](_anything)?.ipns-record
 */
func extractPubkeyFromPath(path string) (string, error) {
	filename := filepath.Base(path)
	r := regexp.MustCompile(`^(.+?)(_.*|)\.ipns-record$`)
	matches := r.FindStringSubmatch(filename)

	if len(matches) < 2 {
		return "", fmt.Errorf("failed to extract pubkey from path: %s", path)
	}

	return matches[1], nil
}


func OpenIPNSRecord(absPath string) (*IpnsRecord, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	r, err := UnmarshalIpnsRecord(data)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func OpenIPNSRecordWithKey(absPath string) (*IpnsRecord, error) {
	// name is [pubkey](_anything)?.ipns-record
	pubkey, err := extractPubkeyFromPath(absPath)
	if err != nil {
		return nil, err
	}

	r, err := OpenIPNSRecord(absPath)
	if err != nil {
		return nil, err
	}
	
	return r.WithKey(pubkey), nil
}

func MustOpenIPNSRecordWithKey(file string) *IpnsRecord {
	fixturePath := path.Join(fixtures.Dir(), file)
	
	ipnsRecord, err := OpenIPNSRecordWithKey(fixturePath)
	if err != nil {
		panic(err)
	}

	return ipnsRecord
}

// func MustOpenDNSLink(file string) *DNSLinks {
// 	fixturePath := path.Join(fixtures.Dir(), file)
// 	dnsLinks, err := OpenDNSLink(fixturePath)
// 	if err != nil {
// 		panic(err)
// 	}

// 	return dnsLinks
// }

// func (d *DNSLinks) Get(id string) string {
// 	dnsLink, ok := d.DNSLinks[id]
// 	if !ok {
// 		panic(fmt.Errorf("dnslink %s not found", id))
// 	}
// 	return dnsLink.Subdomain
// }
