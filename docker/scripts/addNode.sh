#!/bin/bash

if [ -d ../.scavengeD ]; then
  rm -r ../.scavengeD
fi

if [ -d ../.scavengeCLI ]; then
  rm -r ../.scavengeCLI
fi

scavengeD init "${1}" --chain-id scavenge

cp -r ../config ../.scavengeD

scavengeCLI config keyring-backend test

scavengeCLI keys add treasury --recover <./treasury_mnemonic.txt
scavengeCLI keys add validator

scavengeCLI config chain-id scavenge
scavengeCLI config output json
scavengeCLI config indent true
scavengeCLI config trust-node true

./modifyConfig.sh

scavengeD start --rpc.laddr tcp://0.0.0.0:26657
