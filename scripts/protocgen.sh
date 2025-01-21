#!/usr/bin/env bash

set -eo pipefail

cd proto

proto_dirs=$(find ./axelar -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yml $file
    fi
  done
done

cd ..

# move proto files to the right places
cp -r github.com/axelarnetwork/axelar-core/* ./
rm -rf github.com
