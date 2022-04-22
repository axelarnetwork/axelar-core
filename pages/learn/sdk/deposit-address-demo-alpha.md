# Deposit address demo - alpha (v0.5.x) sdk

Git: [sdk-demo-v5-deposit-address](https://github.com/axelarnetwork/sdk-demo-v5-deposit-address)

This simple frontend demo app uses [axelarjs-sdk](https://github.com/axelarnetwork/axelarjs-sdk) to enable a user to transfer AXL tokens from Axelar to Avalanche.

This demo performs one task: call `getDepositAddress` to request a one-time deposit address `A` from the Axelar network and present `A` to the user.

From here, the user may send AXL tokens to `A` on the Axelar blockchain. Any user who does this will soon see wrapped Axelar tokens appear in her Avalanche wallet.

# Developer notes

Refer to [axelarjs-sdk](https://github.com/axelarnetwork/axelarjs-sdk) for code snippets on SDK setup, instantiation, and invocation.

# What the user sees

## Run the demo

Clone this repo, install axelarjs-sdk, and run the server

```bash
git clone git@github.com:axelarnetwork/sdk-demo-v5-deposit-address.git
cd sdk-demo-v5-deposit-address
npm install
npm start
```
Your browser should open automatically to URL `http://localhost:3000/` and display the following

![deposit-address-demo welcome screen](/images/deposit-address-demo-welcome-alpha.png)

Click the "Generate" button in the demo. After a few seconds you should see the following

![deposit-address-demo example address](/images/deposit-address-demo-address-alpha.png)

The demo is now complete.

## Optional: transfer AXL tokens to Avalanche

Send AXL testnet tokens to the one-time deposit address `axelar1...`. One way to do this is visit [Axelar Testnet Faucet](https://faucet.testnet.axelar.dev/).

Check the balance of your one-time deposit address at [Axelarscan testnet explorer](https://testnet.axelarscan.io/)

The Axelar network microservices will automatically transfer AXL tokens from your one-time address to your Metamask Avalanche testnet account.

Wait a few minutes. Then check the "Axelar (AXL)" ERC20 token balance of your Metamask Avalanche testnet account at [SnowTrace: Avalanche Testnet C-Chain Blockchain Explorer](https://testnet.snowtrace.io/)
