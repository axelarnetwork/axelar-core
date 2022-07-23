# Forking mainnet for local testing

For development and debugging purposes, it is often useful to be able to fork preexisting chains. This can easily be done using `ganache`. You can simply install [`axelar-local-dev`](https://github.com/axelarnetwork/axelar-local-dev) by typing `npm i @axelar-network/axelar-local-dev` and writing a simple script:

```js
const { forkAndExport } = '@axelar-network/axelar-local-dev';
forkAndExport();
```

This will clone `mainnet` by default and export local RPC endpoints at `http:/localhost:8500/n` with `n` staring from `0`.

## Aditional Options

`forkAndExport` can take an object of type `CloneLocalOptions` that can specify:
- `chainOutputPath`: a path to save a JSON file with all the information for the chains that are set up. Defaults to `./local.json`.
- `accountsToFund`: a list of addresses to fund.
- `fundAmount`: a string representing the amount of ether to fund accounts with. Defaults to `100 ETH`.
- `env`: a `string` whose value is either `mainnet` or `testnet`, or an `array` of `ChainCloneData`. See the repo for details.
- `chains`: now acts as a filter for which chains to fork. Defaults to all the chains.
- `relayInterval`: amount of time between relay of events in miliseconds. Defaults to `2000`.
- `port`: port to listen to. Defaults to `8500`.
- `afterRelay`: a function `(relayData: RelayData) => void` that will be called after each relay. Mainly to be used for debugging.
- `callback`: a function `(network: Network, info: any) => Promise<null>` that will be called right after setting up each network. Use this to set up additional features, like deploying additional contracts or funding accounts.
- `networkInfo`: the `NetworkInfo` which overwrites the default parameters. See the repo for details.

## Usage

After creating the desired fork, you have access to RPCs (look into the JSON file saved) to them. You can update MetaMask (or any other wallet) to these RPCs and access any dApp online, and everything will happen in the local fork instead of the actual network. All Axelar functionality will happen more quickly (`2s` by default), and you can get debigging capabilites via the `afterRelay` option. Make sure you change MetaMask's RPCs back to the endpoints you trust, to go back to using actual mainnet/testnet.

Forking requires the use of RPCs to the actual networks being forked. For mainnet and testnet some RPCs are provided as part of `axelar-local-dev`, but we cannot guarantee that they will all work in the future. If you want, you can either tweak them in `./node_modules/@axelar-network/axelar-local-dev/info/` or provide an `env` variable that specifies your own RPCs. You can also copy the files in the `info` directory, alter them, load them and pass either of them as the `env` variable.