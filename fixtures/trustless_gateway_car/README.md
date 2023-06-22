# Fixture for Trustless Gateway Car Testing

## Recipes

### [file-3k-and-3-blocks-missing-block.car](./file-3k-and-3-blocks-missing-block.car)

We generate a random file, chunked into 3 * 1024 blocks. Then use `go-car` to remove the
middle block.

```sh
dd if=/dev/urandom of="file-3k-and-3-blocks.bin" bs=1024 count=3
CID=$(ipfs add ./file-3k-and-3-blocks.bin --chunker=size-1024 -q)
ipfs dag export $CID > file-3k-and-3-blocks.car
REMOVE_BLOCK=$(ipfs dag get $CID | jq '.Links[1].Hash["/"]' -r)
echo $REMOVE_BLOCK | car filter --version 1 --inverse ./file-3k-and-3-blocks.car ./file-3k-and-3-blocks-missing-block.car 
```
