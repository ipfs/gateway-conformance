all: dag.car issue-54.car

dag.car:
	npx ipfs-car pack --no-wrap ./dir --output ./dag.car

issue-54.car:
	mkdir -p dir-54/sub1
	echo "abc123" > dir-54/sub1/hello.txt
	mkdir -p dir-54/sub2
	echo "xyz456" > dir-54/sub2/hello.txt
	ipfs dag export `ipfs add -Qr --chunker=size-5 dir-54` > ./issue-54.car


file-3k-and-3-blocks-missing-block.car:
	@dd if=/dev/urandom of="/tmp/file-3k-and-3-blocks.bin" bs=1024 count=3; 	\
	CID=$$(ipfs add /tmp/file-3k-and-3-blocks.bin --chunker=size-1024 -q); 	\
	ipfs dag export $$CID > /tmp/file-3k-and-3-blocks.car; 						\
	REMOVE_BLOCK=$$(ipfs dag get $$CID | jq '.Links[1].Hash["/"]' -r); 			\
	echo $$REMOVE_BLOCK | car filter --version 1 --inverse /tmp/file-3k-and-3-blocks.car ./file-3k-and-3-blocks-missing-block.car;
