#!/usr/bin/env bash

set -eo pipefail

SWAGGER_DIR=./swagger-proto

# configure buf workspace settings
mkdir -p "$SWAGGER_DIR/proto"
printf "version: v1\ndirectories:\n  - proto\n  - third_party" > "$SWAGGER_DIR/buf.work.yaml"
printf "version: v1\nname: buf.build/axelarnetwork/axelar-core\n" > "$SWAGGER_DIR/proto/buf.yaml"
cp ./proto/buf.gen.swagger.yaml "$SWAGGER_DIR/buf.gen.swagger.yaml"

# copy existing proto files
cp -r ./proto/axelar "$SWAGGER_DIR/proto"

# create temporary folder to store intermediate results from `buf generate`
mkdir -p ./tmp-swagger-gen

cd "$SWAGGER_DIR"

# Get the path of the cosmos-sdk repo from go/pkg/mod
proto_dirs=$(find ./proto ./third_party -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  proto_files=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ -z "$proto_files" ]]; then
      continue
  fi

  for proto_file in ${proto_files}; do
    buf generate --template buf.gen.swagger.yaml $proto_file
  done
done

cd ..

swagger-combine ./client/docs/config.json -o o ./client/docs/static/swagger/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf "$SWAGGER_DIR"
rm -rf ./tmp-swagger-gen