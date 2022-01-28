#!/usr/bin/env bash

artifactsPath="contract-artifacts/gateway"

declare -a contracts=("AxelarGatewayProxySinglesig"
                      "AxelarGatewayProxyMultisig"
                      "BurnableMintableCappedERC20"
                      "Burner"
                      "Absorber"
                      )

cp x/evm/types/contracts.go.template x/evm/types/contracts.go

for contract in "${contracts[@]}"
do
  if [ ! -f "$artifactsPath/$contract.json" ]; then
      echo "Error: Contract file $contract.json not found in $artifactsPath"
      exit 1
  fi

  echo "Populating $contract"

  bytecode="$(cat "$artifactsPath/$contract.json" | jq -r '.bytecode')"

  sed -i.bak "s/%${contract}_bytecode%/$bytecode/g" x/evm/types/contracts.go
done

rm x/evm/types/contracts.go.bak

echo "Generated: x/evm/types/contracts.go"
