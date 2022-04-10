# Set up Metamask

1. Connect Metamask to other EVM chains
2. Get testnet tokens for other EVM chains to pay for gas
3. Import Axelar ERC20 tokens on other EVM chains
4. Enable hex data in transactions

## Connect Metamask to other EVM chains

In order to complete exercises for a EVM chain `[chain]` you need to connect your Metamask to `[chain]`.

Open Metamask. In the "Networks" dropdown list choose "Add Network". Enter the data for your desired `[chain]` below and click "Save". Repeat for any chains you like.

| EVM chain | Network Name              | Chain ID | Currency Symbol | RPC URL                                                              | Block Explorer URL                                           |
| --------- | ------------------------- | -------- | --------------- | -------------------------------------------------------------------- | ------------------------------------------------------------ |
| Ethereum  | Ethereum Ropsten Testnet  | 3        | ETH             | [url](https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161) | [url](https://ropsten.etherscan.io)                          |
| Avalanche | Avalanche C-Chain Testnet | 43113    | C-AVAX          | [url](https://api.avax-test.network/ext/bc/C/rpc)                    | [url](https://cchain.explorer.avax-test.network)             |
| Fantom    | Fantom Testnet            | 4002     | FTM             | [url](https://rpc.testnet.fantom.network/)                           | [url](https://testnet.ftmscan.com/)                          |
| Moonbeam  | Moonbase Alpha Testnet    | 1287     | DEV             | [url](https://rpc.testnet.moonbeam.network)                          | [url](https://moonbase-blockscout.testnet.moonbeam.network/) |
| Polygon   | Polygon Mumbain Testnet   | 80001    | MATIC           | [url](https://rpc-mumbai.maticvigil.com/)                            | [url](https://mumbai.polygonscan.com/)                       |

## Get testnet tokens for EVM chains

You need native tokens for each `[chain]` in order to pay transaction fees (gas) on `[chain]`.

You can get native tokens from a faucet. Search the internet for "`[chain]` testnet faucet" or use the links below.

- [Ethereum](https://faucet.dimensions.network/)
- [Avalanche](https://faucet.avax-test.network/)
- [Fantom](https://faucet.fantom.network/)
- [Moonbeam](https://docs.moonbeam.network/builders/get-started/moonbase/#get-tokens) -- No known web faucet; need to join the [Moonbeam discord](https://discord.gg/PfpUATX).
- [Polygon](https://faucet.polygon.technology/)

## Import Axelar ERC20 tokens

Tokens transferred to an EVM chain using Axelar are not visible in Metamask until you import them.

1. Use the "Networks" dropdown list, select your desired `[chain]`.
2. View "Assets" and select "Import Tokens".
3. Paste into "Token Contract Address" the ERC20 address for the token. ("Token symbol" and "token decimal" should be fetched automatically.)

Axelar token contract addresses for each `[chain]` can be found at [Testnet resources](/resources/testnet).

## Enable hex data in transactions

Some advanced exercises require you to send a transaction with hex data from Metamask. The "hex data" field is invisible by default. Edit your settings to make it visible.

Accounts dropdown list -> Settings -> Advanced -> Show Hex Data, switch on.