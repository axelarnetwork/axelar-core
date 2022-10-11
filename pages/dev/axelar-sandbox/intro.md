# Introduction

Cross-chain dapp testing is significantly more time-consuming than single-chain dapp testing. Developers would typically deploy smart contracts to supported networks on testnet, submit a cross-chain transaction to the deployed contract, wait for our network to relay the transaction to the destination chain, and then check the result at your destination contract.

The complete process normally takes about 2-3 minutes on the Testnet. It is too slow when the destination chain's smart contract contains unknown problems that needs be debugged by developers. Axelar Sandbox offers simulated settings that minimize block time and automatically relay transactions without the need for validators. As a result, developers may reduce the time it takes from start to finish from 2-3 minutes to 10 seconds.

## What is Axelar Sandbox?

Axelar Sandbox is a rapid prototyping sandbox for cross-chain dapps. It includes a Solidity and Javascript editor that runs in your browser, as well as evm-blockchain nodes, Axelar Gateway Contract, and Axelar Relayer running on our server. The cross-chain transactions will be picked up by our simulated relayer periodically, and if everything is correct, they will be relayed to the destination chain immediately. The outcome of the cross-chain transaction can then be confirmed by developers using javascript's console, transaction logs, and transaction events.

The sandbox app can be found at https://xchainbox.axelar.dev

To learn more, please continue reading [how-to-use](/dev/axelar-sandbox/how-to-use) doc.
