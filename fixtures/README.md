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

path: /
  cid: bafybeie7u7bbgvdqbur6inl6fpl4yw3xcg34bjjmnnzkmpwb5v6x3xz3hy
  raw: [18 45 10 36 1 112 18 32 23 99 236 14 55 58 97 20 132 58 64 197 84 83 21 225 160 89 247 136 255 156 159 126 203 7 206 101 85 56 150 23 18 3 100 105 114 24 90 10 2 8 1]
path: /dir
  cid: bafybeiaxmpwa4nz2mekiiosayvkfgfpbubm7pch7tspx5syhzzsvkoewc4
  raw: [18 51 10 36 1 85 18 32 56 106 16 112 65 167 151 119 230 6 162 221 175 239 84 37 156 209 33 123 49 169 1 229 100 138 166 237 192 75 203 163 18 9 97 115 99 105 105 46 116 120 116 24 33 10 2 8 1]
path: /dir/ascii.txt
  cid: bafkreibyniihaqnhs536mbvc3wx66vbfttisc6zrvea6kzeku3w4as6lum
  raw: [103 111 111 100 98 121 101 32 97 112 112 108 105 99 97 116 105 111 110 47 118 110 100 46 105 112 108 100 46 114 97 119 10]
  str: goodbye application/vnd.ipld.raw
