package ipns

import (
	"testing"
	"time"

	mbase "github.com/multiformats/go-multibase"
	"github.com/multiformats/go-multicodec"
	"github.com/stretchr/testify/assert"
)

var YEAR_100 time.Time

func init() {
	YEAR_100 = time.Time(time.Date(2123, time.March, 01, 01, 0, 0, 801257000, time.UTC))
}

func TestExtractPath(t *testing.T) {
	path := "a/b/c/key_suffix.ipns-record"
	k, err := extractPubkeyFromPath(path)

	assert.Nil(t, err)
	assert.Equal(t, "key", k)

	path = "a/b/c/something.ipns-record"
	k, err = extractPubkeyFromPath(path)

	assert.Nil(t, err)
	assert.Equal(t, "something", k)

	path = "a/b/c/brokenrecord.ipns"
	k, err = extractPubkeyFromPath(path)

	assert.NotNil(t, err)
	assert.Equal(t, "", k)
}

func TestLoadIPNSRecord(t *testing.T) {
	path := "./_fixtures/k51qzi5uqu5dh71qgwangrt6r0nd4094i88nsady6qgd1dhjcyfsaqmpp143ab.ipns-record"
	ipns, err := OpenIPNSRecordWithKey(path)

	assert.Nil(t, err)
	assert.Equal(t, "k51qzi5uqu5dh71qgwangrt6r0nd4094i88nsady6qgd1dhjcyfsaqmpp143ab", ipns.Key())
	assert.Equal(t, ipns.Value(), "/ipfs/bafkreidfdrlkeq4m4xnxuyx6iae76fdm4wgl5d4xzsb77ixhyqwumhz244")
	assert.True(t, ipns.Validity().After(YEAR_100))

	err = ipns.Valid()
	assert.NoError(t, err)
}

func TestLoadTestRecord(t *testing.T) {
	// Test record created with (100 years):
	// ipfs name publish --allow-offline -t 876000h --key=self "/ipfs/$( echo "helloworld" | ipfs add --inline -q )"
	// ipfs routing get /ipns/${K} > ${K}.ipns-record

	path := "./_fixtures/k51qzi5uqu5dgh7y9l90nqs6tvnzcm9erbt8fhzg3fu79p5qt9zb2izvfu51ki.ipns-record"
	ipns, err := OpenIPNSRecordWithKey(path)

	assert.Nil(t, err)
	assert.Equal(t, "k51qzi5uqu5dgh7y9l90nqs6tvnzcm9erbt8fhzg3fu79p5qt9zb2izvfu51ki", ipns.Key())
	assert.Equal(t, ipns.Value(), "/ipfs/bafyaaeykceeaeeqlnbswy3dpo5xxe3debimaw")
	assert.True(t, ipns.Validity().After(YEAR_100))

	err = ipns.Valid()
	assert.NoError(t, err)
}

func TestIPNSFixtureVersionsConversion(t *testing.T) {
	path := "./_fixtures/12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d.ipns-record"
	record, err := OpenIPNSRecordWithKey(path)

	assert.Nil(t, err)

	// 12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d is a ED25519 key, which is using the identity hash.
	assert.Equal(t, "12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d", record.Key())
	assert.Equal(t, "12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d", record.IdV0())
	assert.Equal(t, "k51qzi5uqu5dk3v4rmjber23h16xnr23bsggmqqil9z2gduiis5se8dht36dam", record.IdV1())
	assert.Equal(t, "k50rm9yjlt0jey4fqg6wafvqprktgbkpgkqdg27tpqje6iimzxewnhvtin9hhq", record.ToCID(multicodec.DagPb, mbase.Base36))
	assert.Equal(t, "12D3KooWLQzUv2FHWGVPXTXSZpdHs7oHbXub2G5WC8Tx4NQhyd2d", record.B58MH())
	assert.Equal(t, "k51qzi5uqu5dk3v4rmjber23h16xnr23bsggmqqil9z2gduiis5se8dht36dam", record.ToCID(multicodec.Libp2pKey, mbase.Base36))

	path = "./_fixtures/QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3.ipns-record"
	record, err = OpenIPNSRecordWithKey(path)

	assert.Nil(t, err)

	// QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3 is a RSA key, which is using sha256 hash.
	assert.Equal(t, "QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3", record.Key())
	assert.Equal(t, "QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3", record.IdV0())
	assert.Equal(t, "k2k4r8m7xvggw5pxxk3abrkwyer625hg01hfyggrai7lk1m63fuihi7w", record.IdV1())
	assert.Equal(t, "k2jmtxu61bnhrtj301lw7zizknztocdbeqhxgv76l2q9t36fn9jbzipo", record.ToCID(multicodec.DagPb, mbase.Base36))
	assert.Equal(t, "QmVujd5Vb7moysJj8itnGufN7MEtPRCNHkKpNuA4onsRa3", record.B58MH())
}
