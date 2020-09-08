#!/bin/bash

if [ -d ../.axelarD ]; then
  rm -r ../.axelarD
fi

if [ -d ../.axelarCLI ]; then
  rm -r ../.axelarCLI
fi

axelarD init "${1}" --chain-id axelar

axelarCLI config keyring-backend test

axelarCLI keys add treasury --recover <./treasury_mnemonic.txt

axelarCLI config chain-id axelar
axelarCLI config output json
axelarCLI config indent true
axelarCLI config trust-node true

axelarCLI keys add validator

axelarD add-genesis-account "$(axelarCLI keys show validator -a)" 100000000stake
axelarD add-genesis-account "$(axelarCLI keys show treasury -a)" 1000000000000foo,100000000000stake

axelarD gentx --name validator --keyring-backend test

axelarD collect-gentxs

axelarD validate-genesis

echo "$(axelarD tendermint show-node-id)@$(hostname):26656" >peers.txt

cp ../.axelarD/config/genesis.json ../config

./modifyConfig.sh -g

axelarD start --rpc.laddr tcp://0.0.0.0:26657
