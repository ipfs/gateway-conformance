# Subdomain Gateway Fixtures

## Recipes

### [fixtures.car](./fixtures.car)

```sh
# using ipfs version 0.21.0-dev (03a98280e3e642774776cd3d0435ab53e5dfa867)

# CIDv0to1 is necessary because raw-leaves are enabled by default during
# "ipfs add" with CIDv1 and disabled with CIDv0
CID_VAL="hello"
CIDv1=$(echo $CID_VAL | ipfs add --cid-version 1 -Q)
CIDv0=$(echo $CID_VAL | ipfs add --cid-version 0 -Q)
CIDv0to1=$(echo "$CIDv0" | ipfs cid base32)
# sha512 will be over 63char limit, even when represented in Base36
CIDv1_TOO_LONG=$(echo $CID_VAL | ipfs add --cid-version 1 --hash sha2-512 -Q)

echo CID_VAL=${CID_VAL}
echo CIDv1=${CIDv1}
echo CIDv0=${CIDv0}
echo CIDv0to1=${CIDv0to1}
echo CIDv1_TOO_LONG=${CIDv1_TOO_LONG}

# Directory tree crafted to test for edge cases like "/ipfs/ipfs/ipns/bar"
mkdir -p testdirlisting/ipfs/ipns &&
echo "hello" > testdirlisting/hello &&
echo "text-file-content" > testdirlisting/ipfs/ipns/bar &&
mkdir -p testdirlisting/api &&
mkdir -p testdirlisting/ipfs &&
echo "I am a txt file" > testdirlisting/api/file.txt &&
echo "I am a txt file" > testdirlisting/ipfs/file.txt &&
DIR_CID=$(ipfs add -Qr --cid-version 1 testdirlisting)

echo DIR_CID=${DIR_CID} # ./testdirlisting

ipfs files mkdir /t0114/
ipfs files cp /ipfs/${CIDv1} /t0114/
ipfs files cp /ipfs/${CIDv0} /t0114/
ipfs files cp /ipfs/${CIDv0to1} /t0114/
ipfs files cp /ipfs/${DIR_CID} /t0114/
ipfs files cp /ipfs/${CIDv1_TOO_LONG} /t0114/

ROOT=`ipfs files stat /t0114/ --hash`

ipfs dag export ${ROOT} > ./fixtures.car
```