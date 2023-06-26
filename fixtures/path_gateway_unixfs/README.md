# Path Gateway Fixtures

## Recipes

### [symlink.car](./symlink.car)

```sh
# using ipfs version 0.18.1
mkdir testfiles &&
echo "content" > testfiles/foo &&
ln -s foo testfiles/bar &&
ROOT_DIR_CID=$(ipfs add -Qr testfiles) &&
ipfs dag export $ROOT_DIR_CID > symlink.car
```
