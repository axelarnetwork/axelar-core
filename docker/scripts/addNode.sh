#!/bin/bash

if [ -d ../.axelarD ]; then
  rm -r ../.axelarD
fi

if [ -d ../.axelarCLI ]; then
  rm -r ../.axelarCLI
fi

axelarD init "${1}" --chain-id axelar

cp -r ../config ../.axelarD

axelarCLI config keyring-backend test

axelarCLI keys add treasury --recover <./treasury_mnemonic.txt
axelarCLI keys add validator

axelarCLI config chain-id axelar
axelarCLI config output json
axelarCLI config indent true
axelarCLI config trust-node true

./modifyConfig.sh

axelarD start --rpc.laddr tcp://0.0.0.0:26657
