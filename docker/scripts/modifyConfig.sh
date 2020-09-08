#!/bin/bash

if [ "$1" == "-g" ]; then
  echo "Is genesis node"
  genesis=true
else
  echo "Is not genesis node"
  genesis=false
fi

modify() {
  sed "s/^$1 =.*/$1 = $2/g" ../.axelarD/config/config.toml >../.axelarD/config/config.toml.tmp &&
    mv ../.axelarD/config/config.toml.tmp ../.axelarD/config/config.toml
}

modify "timeout_commit" "\"1s\""
if ! $genesis; then
  peers=$(<peers.txt)
  modify "persistent_peers" "\"$peers\""
fi

modify "prometheus" true
