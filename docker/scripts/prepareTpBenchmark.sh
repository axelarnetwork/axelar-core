#!/bin/bash

if [ -z "$1" ]; then
  echo "define the number of test accounts to create"
  exit 1
fi

for ((i = 0; i < $1; i++)); do
  axelarCLI keys add "test$i"
  axelarCLI tx send "$(axelarCLI keys show treasury -a)" "$(axelarCLI keys show "test$i" -a)" 10000foo -y
done
