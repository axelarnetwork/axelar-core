#!/bin/bash

for node in "$@"
do
  echo "==== Making $node a validator ===="
  docker exec "$node" bash makeValidator.sh "$node"
done