# Subdomain Gateway Fixtures

## Recipes

### [dnslink.yml](./dnslink.yml)

See comments in the yml file.

### [fixtures.car](./fixtures.car)

```sh
# using Kubo CLI version 0.21.0-rc3 (https://dist.ipfs.tech/kubo/v0.21.0-rc3/)

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
ipfs files cp /ipfs/${CIDv1} /t0114/hello-CIDv1
ipfs files cp /ipfs/${CIDv0} /t0114/hello-CIDv0
ipfs files cp /ipfs/${CIDv0to1} /t0114/hello-CIDv0to1
ipfs files cp /ipfs/${DIR_CID} /t0114/testdirlisting
ipfs files cp /ipfs/${CIDv1_TOO_LONG} /t0114/hello-CIDv1_TOO_LONG

ROOT=`ipfs files stat /t0114/ --hash`

ipfs dag export ${ROOT} > ./fixtures.car
```

### [ipns-records](./)

We create ipns-records pointing to the CIDv1 generated above:

```sh
KEY_NAME=test_key_rsa_$RANDOM
RSA_KEY=$(ipfs key gen --ipns-base=b58mh --type=rsa --size=2048 ${KEY_NAME} | head -n1 | tr -d "\n")
RSA_IPNS_IDv0=$(echo "$RSA_KEY" | ipfs cid format -v 0)
RSA_IPNS_IDv1=$(echo "$RSA_KEY" | ipfs cid format -v 1 --mc libp2p-key -b base36)
RSA_IPNS_IDv1_DAGPB=$(echo "$RSA_IPNS_IDv0" | ipfs cid format -v 1 -b base36)

# publish a record valid for a 100 years
ipfs name publish --key ${KEY_NAME} --allow-offline -Q --ttl=876600h --lifetime=876600h "/ipfs/$CIDv1"
ipfs routing get /ipns/${RSA_KEY} > ${RSA_KEY}.ipns-record

echo RSA_KEY=${RSA_KEY}
echo RSA_IPNS_IDv0=${RSA_IPNS_IDv0}
echo RSA_IPNS_IDv1=${RSA_IPNS_IDv1}
echo RSA_IPNS_IDv1_DAGPB=${RSA_IPNS_IDv1_DAGPB}

KEY_NAME=test_key_ed25519_$RANDOM
ED25519_KEY=$(ipfs key gen --ipns-base=b58mh --type=ed25519 ${KEY_NAME} | head -n1 | tr -d "\n")
ED25519_IPNS_IDv0=$ED25519_KEY
ED25519_IPNS_IDv1=$(ipfs key list -l --ipns-base=base36 | grep ${KEY_NAME} | cut -d " " -f1 | tr -d "\n")
ED25519_IPNS_IDv1_DAGPB=$(echo "$ED25519_IPNS_IDv1" | ipfs cid format -v 1 -b base36 --mc dag-pb)

# ed25519 fits under 63 char limit when represented in base36
IPNS_ED25519_B58MH=$(ipfs key list -l --ipns-base b58mh | grep $KEY_NAME | cut -d" " -f1 | tr -d "\n")
IPNS_ED25519_B36CID=$(ipfs key list -l --ipns-base base36 | grep $KEY_NAME | cut -d" " -f1 | tr -d "\n")

# publish a record valid for a 100 years
ipfs name publish --key ${KEY_NAME} --allow-offline -Q --ttl=876600h --lifetime=876600h "/ipfs/$CIDv1"
ipfs routing get /ipns/${ED25519_KEY} > ${ED25519_KEY}.ipns-record

echo ED25519_KEY=${ED25519_KEY}
echo ED25519_IPNS_IDv0=${ED25519_IPNS_IDv0}
echo ED25519_IPNS_IDv1=${ED25519_IPNS_IDv1}
echo ED25519_IPNS_IDv1_DAGPB=${ED25519_IPNS_IDv1_DAGPB}
echo IPNS_ED25519_B58MH=${IPNS_ED25519_B58MH}
echo IPNS_ED25519_B36CID=${IPNS_ED25519_B36CID}
```
