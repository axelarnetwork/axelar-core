#!/bin/bash

axelarCLI tx send "$(axelarCLI keys show treasury -a)" "$(axelarCLI keys show validator -a)" 100000000stake \
  --yes -b block

axelarCLI tx staking create-validator --yes \
  --amount 100000000stake \
  --moniker "${1}" \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1" \
  --pubkey "$(axelarD tendermint show-validator)" \
  --from validator
