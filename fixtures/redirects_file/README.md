# Fixture for Redirects File Testing

## Recipes

### [dnslink.yml](./dnslink.yml)

See comments in the yml file.


### [redirects.car](./redirects.car)

Fixtures based on [specs.ipfs.tech/http-gateways/web-redirects-file/#test-fixtures](https://specs.ipfs.tech/http-gateways/web-redirects-file/#test-fixtures) and [IPIP-0002](https://specs.ipfs.tech/ipips/ipip-0002/)

### [redirects-spa.car](./redirects-spa.car)

```sh
ipfs version
# ipfs version 0.22.0
REDIRECTS=$(cat <<-EOF
# Map SPA routes to the main index HTML file.
/* /index.html 200
EOF
)
REDIRECTS_CID=$(echo $REDIRECTS | ipfs add --cid-version=1 -q)
HELLO_CID=$(echo "hello world" | ipfs add --cid-version=1 -q)
ipfs files mkdir -p --cid-version 1 /redirects-spa
ipfs files cp /ipfs/$REDIRECTS_CID "/redirects-spa/_redirects"
ipfs files cp /ipfs/$HELLO_CID "/redirects-spa/index.html"
ipfs files ls -l
# Manually CID of "redirects-spa" and then...
ipfs dag export $CID
```
