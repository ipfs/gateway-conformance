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

### [dir-with-tricky-filenames.car](./dir-with-tricky-filenames.car)

UnixFS directory with filenames that require percent-encoding when returned
in the `Ipfs-Uri` response header (space, `%`, `#`, `?`, non-ASCII), plus
filenames in major scripts. All names are NFC-normalized UTF-8.

```sh
ipfs version
# ipfs version 0.44.0-dev
mkdir dir-with-tricky-filenames &&
printf 'plain ascii name\n' > 'dir-with-tricky-filenames/plain.txt' &&
printf 'name with space\n' > 'dir-with-tricky-filenames/with space.txt' &&
printf 'name with percent sign\n' > 'dir-with-tricky-filenames/100% sure.txt' &&
printf 'name with hash and question mark\n' > 'dir-with-tricky-filenames/a#b?c.txt' &&
printf 'name with latin diacritics\n' > 'dir-with-tricky-filenames/łódź.txt' &&
printf 'name with emoji\n' > 'dir-with-tricky-filenames/emoji🚀.txt' &&
printf 'name in chinese\n' > 'dir-with-tricky-filenames/你好.txt' &&
printf 'name in japanese\n' > 'dir-with-tricky-filenames/ファイル.txt' &&
printf 'name in korean\n' > 'dir-with-tricky-filenames/파일.txt' &&
printf 'name in arabic\n' > 'dir-with-tricky-filenames/ملف.txt' &&
printf 'name in hebrew\n' > 'dir-with-tricky-filenames/קובץ.txt' &&
printf 'name in cyrillic\n' > 'dir-with-tricky-filenames/файл.txt' &&
printf 'name in greek\n' > 'dir-with-tricky-filenames/αρχείο.txt' &&
printf 'name in devanagari\n' > 'dir-with-tricky-filenames/नमस्ते.txt' &&
printf 'name in thai\n' > 'dir-with-tricky-filenames/ไฟล์.txt' &&
ROOT_DIR_CID=$(ipfs add -Qr --cid-version 1 dir-with-tricky-filenames) &&
ipfs dag export $ROOT_DIR_CID > dir-with-tricky-filenames.car

echo ROOT_DIR_CID=${ROOT_DIR_CID} # bafybeiflhd5aimv4xavauidbetge3v3uadbqfibu5lfo4b26yyieqvnvae
```

### [dir-with-tricky-nested-filenames.car](./dir-with-tricky-nested-filenames.car)

UnixFS directory with a subdirectory whose name needs percent-encoding, and
a file name using the sub-delims that platform URL encoders leave raw
(`!'()*`) plus an unreserved `~` that some over-encode.

```sh
ipfs version
# ipfs version 0.43.0
mkdir -p 'dir-with-tricky-nested-filenames/sub dir' &&
printf 'name with encoder traps\n' > "dir-with-tricky-nested-filenames/sub dir/file!'()*~.txt" &&
ROOT_DIR_CID=$(ipfs add -Qr --cid-version 1 dir-with-tricky-nested-filenames) &&
ipfs dag export $ROOT_DIR_CID > dir-with-tricky-nested-filenames.car

echo ROOT_DIR_CID=${ROOT_DIR_CID} # bafybeiab2yboxqoybfufvfjatsmqu2mkzf6piueex7xwsk74tdxobyuw6u
```

### [dir-with-slash-in-filename.car](./dir-with-slash-in-filename.car)

Hand-assembled dag-pb directory holding a real subdirectory `a` with
`b.txt` inside, plus a sibling link literally named `a/b.txt`. Such a link
is legal at the dag-pb level but effectively illegal in UnixFS: the most
popular interfaces, HTTP gateways and `ipfs://` URIs, cannot address it,
because the path `a/b.txt` always resolves through the subdirectory.
[IPIP-548](https://github.com/ipfs/specs/pull/548) locks this behavior
down. Regular tooling refuses to create names containing `/`, hence
`ipfs dag put`:

```sh
ipfs version
# ipfs version 0.43.0
mkdir -p a && printf 'real nested file\n' > a/b.txt &&
SUBDIR=$(ipfs add -Qr --cid-version 1 a) &&
DECOY=$(printf 'not path-addressable\n' | ipfs add -Q --cid-version 1) &&
echo '{"Data":{"/":{"bytes":"CAE"}},"Links":[
  {"Hash":{"/":"'$SUBDIR'"},"Name":"a","Tsize":70},
  {"Hash":{"/":"'$DECOY'"},"Name":"a/b.txt","Tsize":21}]}' |
ipfs dag put --store-codec dag-pb --input-codec dag-json
# bafybeihuqitp4tzukehqfyaozl6zexd7szyzeywojuynfxopnn7dqjepv4
ipfs dag export bafybeihuqitp4tzukehqfyaozl6zexd7szyzeywojuynfxopnn7dqjepv4 > dir-with-slash-in-filename.car
```

### [dir-with-percent-encoded-filename.car](./dir-with-percent-encoded-name-file.car)

```sh
ipfs version
# ipfs version 0.22.0
CID=$(echo "hello from a percent encoded filename" | ipfs add --cid-version=1 -q)
ipfs files mkdir -p --cid-version 1 /dir-with-percent-encoded-filename
ipfs files cp /ipfs/$CID "/dir-with-percent-encoded-filename/Portugal%2C+España=Peninsula Ibérica.txt"
ipfs files ls -l
# Manually CID of "dir-with-percent-encoded-filename" and then...
ipfs dag export $CID
```