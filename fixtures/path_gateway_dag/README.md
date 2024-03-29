# Path Gateway Dag Fixtures

## Recipes

### Legacy fixtures from Kubo

- dag-cbor-traversal.car
- dag-json-traversal.car
- dag-pb.car
- dag-pb.json

### [gateway-json-cbor.car](./gateway-json-cbor.car) and ipns-records

Generated with:

```sh
# using Kubo CLI version 0.21.0-rc3 (https://dist.ipfs.tech/kubo/v0.21.0-rc3/)

mkdir -p rootDir/ipfs &&
mkdir -p rootDir/ipns &&
mkdir -p rootDir/api &&
mkdir -p rootDir/ą/ę &&
echo "{ \"test\": \"i am a plain json file\" }" > rootDir/ą/ę/t.json &&
echo "I am a txt file on path with utf8" > rootDir/ą/ę/file-źł.txt &&
echo "I am a txt file in confusing /api dir" > rootDir/api/file.txt &&
echo "I am a txt file in confusing /ipfs dir" > rootDir/ipfs/file.txt &&
echo "I am a txt file in confusing /ipns dir" > rootDir/ipns/file.txt &&
DIR_CID=$(ipfs add -Qr --cid-version 1 rootDir) &&
FILE_JSON_CID=$(ipfs files stat --enc=json /ipfs/$DIR_CID/ą/ę/t.json | jq -r .Hash) &&
FILE_CID=$(ipfs files stat --enc=json /ipfs/$DIR_CID/ą/ę/file-źł.txt | jq -r .Hash) &&
FILE_SIZE=$(ipfs files stat --enc=json /ipfs/$DIR_CID/ą/ę/file-źł.txt | jq -r .Size)
echo "$FILE_CID / $FILE_SIZE"

echo DIR_CID=${DIR_CID} # ./rootDir
echo FILE_JSON_CID=${FILE_JSON_CID} # ./rootDir/ą/ę/t.json
echo FILE_CID=${FILE_CID} # ./rootDir/ą/ę/file-źł.txt
echo FILE_SIZE=${FILE_SIZE}

ipfs dag export ${DIR_CID} > gateway-json-cbor.car

DAG_CBOR_TRAVERSAL_CID="bafyreibs4utpgbn7uqegmd2goqz4bkyflre2ek2iwv743fhvylwi4zeeim"
DAG_JSON_TRAVERSAL_CID="baguqeeram5ujjqrwheyaty3w5gdsmoz6vittchvhk723jjqxk7hakxkd47xq"
DAG_PB_CID="bafybeiegxwlgmoh2cny7qlolykdf7aq7g6dlommarldrbm7c4hbckhfcke"

test_native_dag() {
  NAME=$1
  CID=$2

  IPNS_ID=$(ipfs key gen --ipns-base=base36 --type=ed25519 ${NAME}_test_key | head -n1 | tr -d "\n")
  ipfs name publish --key ${NAME}_test_key --allow-offline --ttl=876600h --lifetime=876600h -Q "/ipfs/${CID}" > name_publish_out

  ipfs routing get /ipns/${IPNS_ID} > ${IPNS_ID}.ipns-record

  echo "IPNS_ID_${NAME}=${IPNS_ID}"
}

test_native_dag "DAG_JSON" "$DAG_JSON_TRAVERSAL_CID"
test_native_dag "DAG_CBOR" "$DAG_CBOR_TRAVERSAL_CID"

# DIR_CID=bafybeiafyvqlazbbbtjnn6how5d6h6l6rxbqc4qgpbmteaiskjrffmyy4a # ./rootDir
# FILE_JSON_CID=bafkreibrppizs3g7axs2jdlnjua6vgpmltv7k72l7v7sa6mmht6mne3qqe # ./rootDir/ą/ę/t.json
# FILE_CID=bafkreialihlqnf5uwo4byh4n3cmwlntwqzxxs2fg5vanqdi3d7tb2l5xkm # ./rootDir/ą/ę/file-źł.txt
# FILE_SIZE=34
# IPNS_ID_DAG_JSON=k51qzi5uqu5dhjghbwdvbo6mi40htrq6e2z4pwgp15pgv3ho1azvidttzh8yy2
# IPNS_ID_DAG_CBOR=k51qzi5uqu5dghjous0agrwavl8vzl64xckoqzwqeqwudfr74kfd11zcyk3b7l
```
