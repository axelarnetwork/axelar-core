#!/bin/bash

if [ -d ../.scavengeD ]; then
  rm -r ../.scavengeD
fi

if [ -d ../.scavengeCLI ]; then
  rm -r ../.scavengeCLI
fi

scavengeD init "${1}" --chain-id scavenge

scavengeCLI config keyring-backend test

scavengeCLI keys add treasury --recover <./treasury_mnemonic.txt

scavengeCLI config chain-id scavenge
scavengeCLI config output json
scavengeCLI config indent true
scavengeCLI config trust-node true

scavengeCLI keys add validator

scavengeD add-genesis-account "$(scavengeCLI keys show validator -a)" 100000000stake
scavengeD add-genesis-account "$(scavengeCLI keys show treasury -a)" 1000000000000foo,100000000000stake

scavengeD gentx --name validator --keyring-backend test

scavengeD collect-gentxs

scavengeD validate-genesis

echo "$(scavengeD tendermint show-node-id)@$(hostname):26656" >peers.txt

cp ../.scavengeD/config/genesis.json ../config

./modifyConfig.sh

scavengeD start --rpc.laddr tcp://0.0.0.0:26657
