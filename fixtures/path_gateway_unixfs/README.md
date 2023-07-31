# Path Gateway Fixtures

## Recipes

### [symlink.car](./symlink.car)

```sh
# using Kubo CLI version 0.18.1 (https://dist.ipfs.tech/kubo/v0.18.1/
mkdir testfiles &&
echo "content" > testfiles/foo &&
ln -s foo testfiles/bar &&
ROOT_DIR_CID=$(ipfs add -Qr testfiles) &&
ipfs dag export $ROOT_DIR_CID > symlink.car
```

### [dir-with-files.car](./dir-with-files.car)

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
ipfs files mkdir -p --cid-version 1 /dir-with-files
ipfs files cp /ipfs/$ASCII_CID /dir-with-files/ascii-copy.txt
ipfs files cp /ipfs/$ASCII_CID /dir-with-files/ascii.txt
ipfs files cp /ipfs/$HELLO_CID /dir-with-files/hello.txt
ipfs files cp /ipfs/$MULTIBLOCK_CID /dir-with-files/multiblock.txt
ipfs files ls -l
# Manually CID of "dir-with-files" and then...
ipfs dag export $CID
```
