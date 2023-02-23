# ipfs add --recursive --raw-leaves --cid-version=1 --wrap-with-directory fixtures/dir

added bafkreibyniihaqnhs536mbvc3wx66vbfttisc6zrvea6kzeku3w4as6lum dir/ascii.txt
added bafybeiaxmpwa4nz2mekiiosayvkfgfpbubm7pch7tspx5syhzzsvkoewc4 dir
added bafybeie7u7bbgvdqbur6inl6fpl4yw3xcg34bjjmnnzkmpwb5v6x3xz3hy

# ipfs dag export bafybeie7u7bbgvdqbur6inl6fpl4yw3xcg34bjjmnnzkmpwb5v6x3xz3hy > fixtures/dir.car

# go install github.com/ipld/go-car/cmd/car@latest

# car list fixtures/dir.car

bafybeie7u7bbgvdqbur6inl6fpl4yw3xcg34bjjmnnzkmpwb5v6x3xz3hy
bafybeiaxmpwa4nz2mekiiosayvkfgfpbubm7pch7tspx5syhzzsvkoewc4
bafkreibyniihaqnhs536mbvc3wx66vbfttisc6zrvea6kzeku3w4as6lum

# car extract --file fixtures/dir.car --verbose

.../gateway-conformance/dir
.../gateway-conformance/dir/ascii.txt

# go run extract.go fixtures/dir.car

/dir
/dir/ascii.txt
goodbye application/vnd.ipld.raw
