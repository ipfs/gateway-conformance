# Fixtures

This folder contains all the fixtures using during the tests, with
the recipes to re-create them.

## Recipes

### [gateway-raw-block.car](./gateway-raw-block.car)

Generated with:

```sh
# using Kubo CLI version 0.18.1 (https://dist.ipfs.tech/kubo/v0.18.1/)
mkdir -p dir &&
echo "hello application/vnd.ipld.raw" > dir/ascii.txt &&
ROOT_DIR_CID=$(ipfs add -Qrw --cid-version 1 dir) &&
FILE_CID=$(ipfs resolve -r /ipfs/$ROOT_DIR_CID/dir/ascii.txt | cut -d "/" -f3) &&
ipfs dag export $ROOT_DIR_CID > gateway-raw-block.car

echo ROOT_DIR_CID=${ROOT_DIR_CID} # ./
echo FILE_CID=${FILE_CID} # ./dir/ascii.txt
```