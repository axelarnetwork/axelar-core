#!/bin/bash
pwd

if [ -d ../.scavengeD ]
then
  rm -r ../.scavengeD
fi

if [ -d ../.scavengeCLI ]
then
  rm -r ../.scavengeCLI
fi

scavengeD init mynode --chain-id scavenge --home ../.scavengeD

scavengeCLI config keyring-backend test --home ../.scavengeCLI

scavengeCLI keys add treasury --home ../.scavengeCLI --recover < ./treasury_mnemonic.txt

scavengeCLI config chain-id scavenge --home ../.scavengeCLI
scavengeCLI config output json --home ../.scavengeCLI
scavengeCLI config indent true --home ../.scavengeCLI
scavengeCLI config trust-node true --home ../.scavengeCLI

cp -r ../config ../.scavengeD

if [[ -n ${PERSISTENT_PEERS} ]]
then
  sed "s/persistent_peers = \"\"/persistent_peers = \"$PERSISTENT_PEERS\"/g" ../.scavengeD/config/config.toml > tmp && mv tmp ../.scavengeD/config/config.toml
fi

scavengeD start --home ../.scavengeD
#tail -f /dev/null