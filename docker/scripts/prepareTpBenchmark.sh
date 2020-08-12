#!/bin/bash

if [ -z "$1" ]; then
  echo "define the number of test accounts to create"
  exit 1
fi

for ((i = 0; i < $1; i++)); do
  scavengeCLI keys add "test$i"
  scavengeCLI tx send "$(scavengeCLI keys show treasury -a)" "$(scavengeCLI keys show "test$i" -a)" 10000foo -y
done
