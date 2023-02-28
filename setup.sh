
set -x
set -e

# CID=`ipfs add -r ./fixtures -Q`
# echo "CID: $CID"
# ipfs dag export $CID > ./fixtures.dag

# ASCII_CID=`ipfs files stat --hash "/ipfs/${CID}/dir/ascii.txt"`
# echo "ASCII_CID: $ASCII_CID"

export IPFS_NS_MAP="dnslink-test.example.com:/ipfs/${ASCII_CID},12D3KooWQbpsnyzdBcxw6GUMbijV8WgXE4L8EtfnbcQWLfyxBKho:/ipfs/${CID},bafzaajaiaejcbpltl72da5f3y7ojrtsa7hsfn5bbnkjbkwyesziqqtdry6vjilku:/ipfs/${CID}"
export IPFS_GATEWAY="http://localhost:8040"

# key: k51qzi5uqu5dlnojhwrggtpty9c0cp5hvnkdozowth4eqb726jvoros8k9niyu => b58 encode: 12D3KooWQbpsnyzdBcxw6GUMbijV8WgXE4L8EtfnbcQWLfyxBKho
# ipfs dag export $CID > ./fixtures.car

echo $IPFS_NS_MAP

# open http://localhost:8040/ipfs/${ASCII_CID}
open http://localhost:8040/ipns/dnslink-test.example.com
open http://localhost:8040/ipns/k51qzi5uqu5dlnojhwrggtpty9c0cp5hvnkdozowth4eqb726jvoros8k9niyu
open http://localhost:8040/ipns/12D3KooWNZuG8phqhoNK9KWcUhwfzA3biDKNCUNVWEaJgigr6Acj

# open "http://localhost:8040/ipfs/QmYBhLYDwVFvxos9h8CGU2ibaY66QNgv8hpfewxaQrPiZj"

# assumes you `go install` the car example
export GOLOG_LOG_LEVEL="debug,namesys=debug"
car -c ./fixtures.car
