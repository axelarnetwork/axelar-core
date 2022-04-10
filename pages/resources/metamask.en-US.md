# Set up Metamask

import AddToWeb3 from '../../components/web3'

1. Connect Metamask to other EVM chains
2. Get testnet tokens for other EVM chains to pay for gas
3. Import Axelar ERC20 tokens on other EVM chains
4. Enable hex data in transactions

## Connect Metamask to other EVM chains

In order to complete exercises for a EVM chain `[chain]` you need to connect your Metamask to `[chain]`.

Open Metamask. In the "Networks" dropdown list choose "Add Network". Enter the data for your desired `[chain]` below and click "Save". Repeat for any chains you like.

| EVM chain | Network Name     | Chain ID | Native Token | RPC URL                                                              | Explorer URL                          | Add Chain                                             |
| --------- | ---------------- | -------- | -------------| -------------------------------------------------------------------- | ------------------------------------- | ----------------------------------------------------- |
| Ethereum  | Ethereum Ropsten | 3        | ETH          | [URL](https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161) | [URL](https://ropsten.etherscan.io)   | <AddToWeb3 environment="testnet" chain="ethereum" />  |
| Avalanche | Avalanche Fuji   | 43113    | C-AVAX       | [URL](https://api.avax-test.network/ext/bc/C/rpc)                    | [URL](https://testnet.snowtrace.io)   | <AddToWeb3 environment="testnet" chain="avalanche" /> |
| Polygon   | Polygon Mumbai   | 80001    | MATIC        | [URL](https://rpc-mumbai.maticvigil.com)                             | [URL](https://mumbai.polygonscan.com) | <AddToWeb3 environment="testnet" chain="polygon" />   |
| Fantom    | Fantom Testnet   | 4002     | FTM          | [URL](https://rpc.testnet.fantom.network)                            | [URL](https://testnet.ftmscan.com/)   | <AddToWeb3 environment="testnet" chain="fantom" />    |
| Moonbeam  | Moonbase Alpha   | 1287     | DEV          | [URL](https://rpc.api.moonbase.moonbeam.network)                     | [URL](https://moonbase.moonscan.io)   | <AddToWeb3 environment="testnet" chain="moonbeam" />  |

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