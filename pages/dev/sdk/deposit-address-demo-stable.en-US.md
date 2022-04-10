# Deposit address demo - stable (v0.4.x) sdk

Git: [deposit-address-demo](https://github.com/axelarnetwork/deposit-address-demo)

This simple frontend demo app uses [axelarjs-sdk](https://github.com/axelarnetwork/axelarjs-sdk) to enable a user to transfer AXL tokens from Axelar to Avalanche.

This demo performs one task: call `axelarJsSDK.getDepositAddress` to request a one-time deposit address `A` from the Axelar network and present `A` to the user.

From here, the user may send AXL tokens to `A` on the Axelar blockchain. Any user who does this will soon see wrapped Axelar tokens appear in her Avalanche wallet.

## Prerequisites

Configure your Metamask as per [Set up Metamask for EVM chains | Axelar Docs](/resources/metamask):

- Add Avalanche testnet to your "networks"
- Import the Axlear ERC20 token to your "assets" for Avalanche

## Run the demo

Clone this repo, install axelarjs-sdk, and run the server

```bash
git clone git@github.com:axelarnetwork/deposit-address-demo.git
cd deposit-address-demo
npm i --save @axelar-network/axelarjs-sdk
npm start
```

Your browser should open automatically to URL `http://localhost:3000/` and display the following

![deposit-address-demo welcome screen](/images/deposit-address-demo-welcome-stable.png)

Metamask should automatically appear. You'll be asked to connect Metamask to `http://localhost:3000/`. (If Metamask is locked then you first need to unlock it using your Metamask password.)

<table>
<tr>
    <td> <img src="/images/metamask-connect-1.png" alt="metamask connect screen 1"/> </td>
    <td> <img src="/images/metamask-connect-2.png" alt="metamask connect screen 2"/> </td>
</tr>
</table>

Select the account you wish to connect with Metamask.

**This Metamask account is the Avalanche testnet account to which you will send AXL tokens.**

Click "next" then "connect" in Metamask.

Click the text "Click here to generate a link address..." in the demo. Metamask will appear again, asking you to sign a one-time code.

![metamask sign code screen](/images/metamask-sign-code.png)

Click "sign" in Metamask.

After a few seconds you should see the following

![deposit-address-demo example address](/images/deposit-address-demo-address-stable.png)

The demo is now complete.

## Optional: transfer AXL tokens to Avalanche

Send AXL testnet tokens to the one-time deposit address `axelar1...`. One way to do this is visit [Axelar Testnet Faucet](https://faucet.testnet.axelar.dev/). (The minimum transfer amount is 10 AXL. You may need to use the faucet multiple times to accumulate enough AXL in your one-time deposit address.)

Check the balance of your one-time deposit address at [Axelarscan testnet explorer](https://testnet.axelarscan.io/)

The Axelar network microservices will automatically transfer AXL tokens from your one-time address to your Metamask Avalanche testnet account.

Wait a few minutes. Then check the "Axelar (AXL)" ERC20 token balance of your Metamask Avalanche testnet account at [SnowTrace: Avalanche Testnet C-Chain Blockchain Explorer](https://testnet.snowtrace.io/)
