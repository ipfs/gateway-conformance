all: dag.car issue-54.car

dag.car:
	npx ipfs-car pack --no-wrap ./dir --output ./dag.car

issue-54.car:
	mkdir -p dir-54/sub1
	echo "abc123" > dir-54/sub1/hello.txt
	mkdir -p dir-54/sub2
	echo "xyz456" > dir-54/sub2/hello.txt
	ipfs dag export `ipfs add -Qr --chunker=size-5 dir-54` > ./issue-54.car