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

cp -r ../config ../.scavengeD

peers=$(<peers.txt)
sed "s/persistent_peers = \"\"/persistent_peers = \"$peers\"/g" ../.scavengeD/config/config.toml > tmp && mv tmp ../.scavengeD/config/config.toml

scavengeCLI config keyring-backend test --home ../.scavengeCLI

scavengeCLI keys add treasury --home ../.scavengeCLI --recover < ./treasury_mnemonic.txt
scavengeCLI keys add validator --home ../.scavengeCLI

scavengeCLI config chain-id scavenge --home ../.scavengeCLI
scavengeCLI config output json --home ../.scavengeCLI
scavengeCLI config indent true --home ../.scavengeCLI
scavengeCLI config trust-node true --home ../.scavengeCLI

scavengeD start
