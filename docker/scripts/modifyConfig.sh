#!/bin/bash

if [ "$1" == "-g" ]; then
  echo "Is genesis node"
  genesis=true
else
  echo "Is not genesis node"
  genesis=false
fi

modify() {
  sed "s/^$1 =.*/$1 = $2/g" ../.scavengeD/config/config.toml >../.scavengeD/config/config.toml.tmp &&
    mv ../.scavengeD/config/config.toml.tmp ../.scavengeD/config/config.toml
}

modify "timeout_commit" "\"1s\""
if ! $genesis; then
  peers=$(<peers.txt)
  modify "persistent_peers" "\"$peers\""
fi

modify "prometheus" true
