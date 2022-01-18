#!/usr/bin/env bash

singlesigGateway="$(cat contract-artifacts/gateway/AxelarGatewayProxySinglesig.json | jq -r '.bytecode')"
multisigGateway="$(cat contract-artifacts/gateway/AxelarGatewayProxyMultisig.json | jq -r '.bytecode')"
token="$(cat contract-artifacts/gateway/BurnableMintableCappedERC20.json | jq -r '.bytecode')"
burnable="$(cat contract-artifacts/gateway/Burner.json | jq -r '.bytecode')"

cp x/evm/types/contracts.go.template x/evm/types/contracts.go

sed -i.bak "s/%AxelarGatewayProxySinglesig_bytecode%/$singlesigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%AxelarGatewayProxyMultisig_bytecode%/$multisigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%BurnableMintableCappedERC20_bytecode%/$token/g" x/evm/types/contracts.go
sed -i.bak "s/%Burner_bytecode%/$burnable/g" x/evm/types/contracts.go

rm x/evm/types/contracts.go.bak
