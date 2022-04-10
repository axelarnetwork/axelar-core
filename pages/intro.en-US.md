# What is Axelar?

Axelar is a decentralized overlay network that connects all blockchains, assets and applications through a universal set of protocols and APIs. It is designed from the ground up to deliver secure interoperability and transport layer for web3.

Deploy your dApp on any blockchain, choosing the one most suitable for your use case. Through a single integration with the Axelar Gateway API, unlock access to users, assets, liquidity and data on any other connected blockchain.

## Learn for your role:

### [Developer](/dev)

Use Axelar gateway contracts to call any EVM contract on any chain:

```solidity
interface IAxelarGateway {

  function callContractWithToken(
    string memory destinationChain,
    string memory contractAddress,
    bytes memory payload,
    string memory symbol,
    uint256 amount
  ) external;

}
```

### [Satellite user](/resources/satellite)

_Satellite_ is a web app built on top of the Axelar network. Use it to transfer assets from one chain to another.

### [Node operator](/node/join)

Learn how to run a node on the Axelar network.

### [Validator](/validator/setup/overview)

Axelar validators facilitate cross-chain connections by participating in the following activities:

- Maintaining the Axelar blockchain with its decentralized and permisionless validator set
- Reaching consensus on cross-chain transactions and events that occur on other blockchains
- Running multiparty cryptography protocols for passing general messages between blockchains

## Links

- [Axelar discord](https://discord.gg/aRZ3Ra6f7D)
- [White paper](https://axelar.network/wp-content/uploads/2021/07/axelar_whitepaper.pdf)