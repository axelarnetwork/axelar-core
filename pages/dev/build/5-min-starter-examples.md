import Button from '../../../components/button'

#  Example cross-chain dApps using GMP

The `axelar-local-gmp-examples` repo has an array of "kick-starter" examples to show the ease of integrating Axelar's GMP protocol into any dApp to bring it cross-chain. The examples range in use cases and complexity but ultimately leverage the two fundamental building blocks for GMP:
- `callContract`
- `callContractWithToken`

Each example is self-contained and generally follows the same steps for deployment and testing:
1. (One-time setup): Install project dependencies and build contracts: [One-time setup instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#one-time-setup).
2. (One-time setup): From a separate terminal window, deploy and run the local network emulator: 
    - `node scripts/createLocal`.
3. Then, deploy and test each example, first locally and then in testnet. A subset of representative examples below.


## "Hello World!"

Say hello to your first dApp on Axelar. The dApp sends a message -- "Hello World" -- from a source to a destination chain using the `callContract` function.

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/call-contract) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#call-contract)

## Airdrop

Send axlUSDC from a source chain to a list of recipients on a destination chain using the `callContractWithToken` function. Each recipient will receive an equal portion of the total tokens sent.

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/call-contract-with-token) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#call-contract-with-token)

## Mint tokens and send cross-chain

Mints some amount of ERC-20 tokens at a source chain and sends it using the `callContract` function to a destination chain. Tokens are burned on the source-chain contract and minted on the destination-chain contract. 

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/cross-chain-token) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#cross-chain-token)

## NFT linker

Send an NFT on a source chain to a recipient on a destination chain using the `callContract` function. If the source chain is where the NFT was originally created, the NFT gets locked in the contract and minted on the destination chain; in the reverse direction, the NFT is burned and transferred to its final recipient on the destination (/original home) chain.

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/nft-linker) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#nft-linker)


## Nonced execution

This is a useful way of implementing ordered execution with nonces for messages that are sent cross-chain. Examples for usage with the `callContract` and the `callContractWithToken` functions are provided. 

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/nonced-execution) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#nonced-execution)

## Two-way example: send back

A two-way example using `callContract` in both directions where a message is sent from a source chain to a destination chain, and an "executed" acknowledgement is sent back to the source chain.

[Source code](https://github.com/axelarnetwork/axelar-local-gmp-examples/tree/main/examples/send-ack) | [Deploy instructions](https://github.com/axelarnetwork/axelar-local-gmp-examples#send-ack)