#!/usr/bin/env bash

set -eo pipefail

protoc_gen_go() {
  if ! grep "github.com/gogo/protobuf => github.com/regen-network/protobuf" go.mod &>/dev/null ; then
    echo -e "\tPlease run this command from somewhere inside the axelar-core folder."
    return 1
  fi
}

protoc_gen_go

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # shellcheck disable=SC2046
  buf protoc \
  -I "proto" \
  -I "third_party/proto" \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  --grpc-gateway_out=logtostderr=true:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto') # this needs to remain unquoted because we want word splitting
done

# command to generate docs using protoc-gen-doc
# shellcheck disable=SC2046
buf protoc \
-I "proto" \
-I "third_party/proto" \
--doc_out=./docs/proto \
--doc_opt=./docs/protodoc-markdown.tmpl,proto-docs.md \
$(find "$(pwd)/proto" -maxdepth 5 -name '*.proto') # this needs to remain unquoted because we want word splitting

# move proto files to the right places
cp -r github.com/axelarnetwork/axelar-core/* ./
rm -rf github.com
