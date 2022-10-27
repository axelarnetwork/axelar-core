# Learn about Axelar

Dive into all the components that come together to form the Axelar network, for a high-level understanding of the vision and design; go deeper and gain technical knowledge of specialized functions, including our CLI and SDK. 


## Gateway smart contracts

Gateway smart contracts allow Axelar to communicate messages across all connected chains. For each EVM chain connected to Axelar network, a Gateway contract is deployed to that chain. This Gateway contract is used to pass messages from the Axelar network to the connected chain, and the Gateway contract is controlled by a key, which is held jointly by all the Axelar validators. This is accomplished through a multi-party cryptography scheme, where the key is divided into many pieces, called key shares. Each validator holds many key shares, and the amount of shares is dictated by the amount of staked AXL tokens the validator has. The Gateway can only execute actions on the external chain if the number of validators holding key shares who authorize the action reaches a set threshold.

The Gateway contract is used by Axelar network to execute cross-chain transactions. Transactions must be confirmed by a vote of validators before they are authorized by the Gateway. This allows Axelar to securely mint and burn tokens or approve General Message Passing transactions across all connected chains.


## Validators

Axelar validators play two main roles in the network.

First, the validators participate in consensus on the Axelar network, producing blocks and validating transactions as with other proof-of-stake chains. Specifically, they perform the generic duties expected from validators of all Cosmos SDK-based chains. For more information about what Cosmos SDK validators do, read the Cosmos SDK [validator docs](https://hub.cosmos.network/main/validators/validator-faq.html).

Beyond this, Axelar validators also have additional duties as they are responsible for verifying all cross-chain activity being processed by the network. This requires validators to run nodes for Axelar-supported chains, and observe those external chains for activity. For example, in an asset transfer flow moving tokens from chain A to chain B, the user requesting the transfer must deposit tokens to a deposit address on chain A, and wait for Axelar network to confirm this deposit. This confirmation is done by the validators. A vote is started on the Axelar network, asking each validator to observe their chain A nodes for the deposit transaction made by the user. The validators then cast votes on whether or not the deposit transaction was observed on their chain A node. The votes are tallied and if the number of confirmation votes surpasses a set threshold, the deposit transaction is considered confirmed by the Axelar network.

At this point, Axelar's multi-party cryptography scheme kicks in. If the destination chain B has an Axelar Gateway deployed, the tokens transferred must be minted by the Gateway smart contract and transferred to the user's chain B deposit address. Each Gateway contract is controlled by a key that is able to issue commands to the Gateway and approve transactions by signing. Each Axelar validator holds a piece of this key, called a key share. Validators agree through their confirmation votes to confirm a deposit and sign a transaction transferring the tokens to the user's address on chain B, which completes the asset transfer. Once enough key shares have agreed, the transaction can proceed.

This process can be applied to passing general messages through the network as well. As you can see, Axelar validators handle the important task of authenticating cross-chain activity, and authorizing the transfer of messages and funds.

As Axelar grows and connects more and more chains, it becomes increasingly restrictive to require every Axelar validator to run a node for every chain supported by the network. Instead, Axelar validators are incentivized to run nodes for as many supported chains as possible, through increased staking rewards, based on the number of chains they support.


## Relayer services

Relayer services are a type of optional convenience services provided by Axelar. These are tasks that can be performed by anyone, and no form of trust is required to authorize or complete the task. These tasks are still important however, as they must be done by someone to enable successful cross-chain communication. Because relayer services do not require any element of trust, and can be implemented by anyone, app developers within the Axelar community can choose to build their own version of existing relayer services for their app to use, instead of using the existing Axelar relayer services.

One example of a relayer service is starting deposit confirmation votes. During a cross-chain asset transfer moving tokens from chain A to chain B, the user first generates a linked deposit address on chain A, and sends the number of tokens they want to transfer cross-chain to the deposit address. Next, Axelar network validators monitor the deposit address for the deposit transaction, and vote to confirm the deposit. Axelar relayer services are responsible for starting off the vote for the deposit confirmation. After the linked deposit address is generated on chain A, the relayer services take the deposit address and submit a request on the Axelar network for validators to monitor the chain A deposit address for a deposit transaction. Without relayer services, the validators would not know which addresses they should be monitoring.


## Gas receiver

The gas receiver is an example of Axelar relayer services.

During General Message Passing, a cross-chain smart contract call is approved by Axelar network validators, and the approval is stored in the Axelar Gateway on the destination chain. This transaction is approved and ready to be executed, but it will not be executed immediately. In order to complete the General Message Passing workflow, the executable destination contract needs to be called with the exact parameters approved by the Axelar Gateway, however, this contract call incurs a gas fee that needs to be paid.

Application developers using Axelar for General Message Passing have two options. They can build their own relayer services on the destination chain, which will cover the gas fees required for the final executable smart contract call; or they can pay gas fees with the Axelar Gas Receiver on the source chain.

The Gas Receiver is a smart contract that accepts tokens as payment to cover costs of contract execution for general message passing transactions. First, send funds to the Gas Receiver on the source chain, and specify the general message passing transaction that should be covered, as well as the payment token and amount. Axelar relayer services will confirm the gas payment on the source chain, then automatically execute the smart contract call on the destination chain when it gets approved.


## Tech stack diagram & network overview

Read Axelar's tech stack [blog post](https://axelar.network/an-introduction-to-the-axelar-network).

## Tech stack walkthrough

Watch an in-depth tech stack [walkthrough video](https://www.youtube.com/watch?v=0-Q1mP2vmGE).

## Tutorial: Building cross-chain NFT

Follow a step by step tutorial for building a [cross-chain app](https://www.youtube.com/watch?v=pAxuQ7PIl8g).

## Tutorial: Onboarding a new EVM chain

Learn how to add a new [EVM chain](https://www.youtube.com/watch?v=iZgqneh7s88&t=13s) to the network.

## AxelarJS SDK

Learn how to build apps with Axelar using the [AxelarJS SDK](./dev/axelarjs-sdk/intro) 

## CLI

Learn about how to interact with an Axelar node through a [command line interface (CLI)](./learn/cli).
