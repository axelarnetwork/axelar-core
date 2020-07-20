#!/bin/bash

if [ -d ../.scavengeD ]
then
  rm -r ../.scavengeD
fi

if [ -d ../.scavengeCLI ]
then
  rm -r ../.scavengeCLI
fi

scavengeD init "${1}" --chain-id scavenge --home ../.scavengeD

scavengeCLI config keyring-backend test --home ../.scavengeCLI

scavengeCLI keys add treasury --home ../.scavengeCLI --recover < ./treasury_mnemonic.txt

scavengeCLI config chain-id scavenge --home ../.scavengeCLI
scavengeCLI config output json --home ../.scavengeCLI
scavengeCLI config indent true --home ../.scavengeCLI
scavengeCLI config trust-node true --home ../.scavengeCLI

scavengeCLI keys add validator --home ../.scavengeCLI

scavengeD add-genesis-account "$(scavengeCLI keys show validator -a --home ../.scavengeCLI)" 100000000stake --home ../.scavengeD
scavengeD add-genesis-account "$(scavengeCLI keys show treasury -a --home ../.scavengeCLI)" 1000000000000foo,100000000000stake --home ../.scavengeD

scavengeD gentx --name validator --keyring-backend test --home ../.scavengeD

scavengeD collect-gentxs --home ../.scavengeD

scavengeD validate-genesis --home ../.scavengeD

echo "$(scavengeD tendermint show-node-id)@$(hostname):26656" > peers.txt

cp ../.scavengeD/config/genesis.json ../config

scavengeD start --home ../.scavengeD
