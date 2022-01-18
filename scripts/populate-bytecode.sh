#!/usr/bin/env bash

artifactsPath="contract-artifacts/gateway"

singlesigGatewayFile="$artifactsPath/AxelarGatewayProxySinglesig.json"
multisigGatewayFile="$artifactsPath/AxelarGatewayProxyMultisig.json"
tokenFile="$artifactsPath/BurnableMintableCappedERC20.json"
burnableFile="$artifactsPath/Burner.json"

if [ ! -f $singlesigGatewayFile ] || [ ! -f $multisigGatewayFile ] || [ ! -f $tokenFile ] || [ ! -f $burnableFile ]; then
    echo "Error: Contract files not found in $artifactsPath"
    exit 1
fi

singlesigGateway="$(cat $singlesigGatewayFile | jq -r '.bytecode')"
multisigGateway="$(cat $multisigGatewayFile | jq -r '.bytecode')"
token="$(cat $tokenFile | jq -r '.bytecode')"
burnable="$(cat $burnableFile | jq -r '.bytecode')"

cp x/evm/types/contracts.go.template x/evm/types/contracts.go

sed -i.bak "s/%AxelarGatewayProxySinglesig_bytecode%/$singlesigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%AxelarGatewayProxyMultisig_bytecode%/$multisigGateway/g" x/evm/types/contracts.go
sed -i.bak "s/%BurnableMintableCappedERC20_bytecode%/$token/g" x/evm/types/contracts.go
sed -i.bak "s/%Burner_bytecode%/$burnable/g" x/evm/types/contracts.go

rm x/evm/types/contracts.go.bak

echo "Generated: x/evm/types/contracts.go"
