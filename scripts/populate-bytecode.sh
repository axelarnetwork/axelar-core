#!/usr/bin/env bash

singlesigGateway="$(cat artifacts/AxelarGatewayProxySinglesig.json | jq -r '.bytecode')"
multisigGateway="$(cat artifacts/AxelarGatewayProxyMultisig.json | jq -r '.bytecode')"
token="$(cat artifacts/BurnableMintableCappedERC20.json | jq -r '.bytecode')"
burnable="$(cat artifacts/Burner.json | jq -r '.bytecode')"

cp x/evm/types/contracts.go.template x/evm/types/contracts.go

sed -i.bak "s/%AxelarGatewayProxySinglesig_bytecode%/$singlesigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%AxelarGatewayProxyMultisig_bytecode%/$multisigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%BurnableMintableCappedERC20_bytecode%/$token/g" x/evm/types/contracts.go
sed -i.bak "s/%Burner_bytecode%/$burnable/g" x/evm/types/contracts.go

cat x/evm/types/contracts.go
