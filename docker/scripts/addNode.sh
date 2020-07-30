#!/bin/bash

if [ -d ../.scavengeD ]
then
  rm -r ../.scavengeD
fi

if [ -d ../.scavengeCLI ]
then
  rm -r ../.scavengeCLI
fi

scavengeD init "${1}" --chain-id scavenge

cp -r ../config ../.scavengeD

peers=$(<peers.txt)
sed "s/persistent_peers = \"\"/persistent_peers = \"$peers\"/g" ../.scavengeD/config/config.toml > tmp && mv tmp ../.scavengeD/config/config.toml

scavengeCLI config keyring-backend test

scavengeCLI keys add treasury --recover < ./treasury_mnemonic.txt
scavengeCLI keys add validator

scavengeCLI config chain-id scavenge
scavengeCLI config output json
scavengeCLI config indent true
scavengeCLI config trust-node true

scavengeD start --rpc.laddr tcp://0.0.0.0:26657
