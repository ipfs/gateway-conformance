# Fixture for Trustless Gateway Car Testing

## Recipes

### [file-3k-and-3-blocks-missing-block.car](./file-3k-and-3-blocks-missing-block.car)

We generate a random file, chunked into 3 * 1024 blocks. Then use `go-car` to remove the
middle block. At the end, we nuke the repo to make sure there are no providers for the removed block.

This is necessary for testing edge case described in [gateway-conformance#75](https://github.com/ipfs/gateway-conformance/issues/75).

```sh
dd if=/dev/urandom of="file-3k-and-3-blocks.bin" bs=1024 count=3
CID=$(ipfs add ./file-3k-and-3-blocks.bin --chunker=size-1024 -q)
ipfs dag export $CID > file-3k-and-3-blocks.car
REMOVE_BLOCK=$(ipfs dag get $CID | jq '.Links[1].Hash["/"]' -r)
echo $REMOVE_BLOCK | car filter --version 1 --inverse ./file-3k-and-3-blocks.car ./file-3k-and-3-blocks-missing-block.car
ipfs pin rm $CID; ipfs repo gc
# First and third outputted CIDs are used in the missing blocks tests.
ipfs dag get $CID  | jq .Links | jq -r '.[].Hash."/"'
```

### [dir-with-duplicate-files.car](./dir-with-duplicate-files.car)

```sh
ipfs version
# ipfs version 0.21.0
TEXT=$(cat <<-EOF 
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nunc non imperdiet nunc. Proin ac quam ut nibh eleifend aliquet. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Sed ligula dolor, imperdiet sagittis arcu et, semper tincidunt urna. Donec et tempor augue, quis sollicitudin metus. Curabitur semper ullamcorper aliquet. Mauris hendrerit sodales lectus eget fermentum. Proin sollicitudin vestibulum commodo. Vivamus nec lectus eu augue aliquet dignissim nec condimentum justo. In hac habitasse platea dictumst. Mauris vel sem neque.

Vivamus finibus, enim at lacinia semper, arcu erat gravida lacus, sit amet gravida magna orci sit amet est. Sed non leo lacus. Nullam viverra ipsum a tincidunt dapibus. Nulla pulvinar ligula sit amet ante ultrices tempus. Proin purus urna, semper sed lobortis quis, gravida vitae ipsum. Aliquam mi urna, pulvinar eu bibendum quis, convallis ac dolor. In gravida justo sed risus ullamcorper, vitae luctus massa hendrerit. Pellentesque habitant amet.
EOF
)

ASCII_CID=$(echo "hello application/vnd.ipld.car" | ipfs add --cid-version=1 -q)
HELLO_CID=$(echo "hello world" | ipfs add --cid-version=1 -q)
MULTIBLOCK_CID=$(echo -n $TEXT | ipfs add --cid-version=1 --chunker=size-256 -q)
# Print the Multiblock CIDs (required for some tests)
ipfs dag get $MULTIBLOCK_CID  | jq .Links | jq -r '.[].Hash."/"'
ipfs files mkdir -p --cid-version 1 /dir-with-duplicate-files
ipfs files cp /ipfs/$ASCII_CID /dir-with-duplicate-files/ascii-copy.txt
ipfs files cp /ipfs/$ASCII_CID /dir-with-duplicate-files/ascii.txt
ipfs files cp /ipfs/$HELLO_CID /dir-with-duplicate-files/hello.txt
ipfs files cp /ipfs/$MULTIBLOCK_CID /dir-with-duplicate-files/multiblock.txt
ipfs files ls -l
# Manually CID of "dir-with-duplicate-files" and then...
ipfs dag export $CID
```
