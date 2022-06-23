# General Message Passing

Axelar's General Message Passing (GMP) enables a developer building on one chain to call any function on any other connected chain. (We use the word “function” to encompass both smart contracts at the application layer and functions built at the protocol layer, as in Cosmos, for example.) That means complete composability across Web3.

With GMP you can:

- Call a contract on chain B from chain A.
- Call a contract on chain B from chain A and attach some tokens.

### Prerequisites
- For GMP to work, both chain A and chain B must be EVM chains with a deployed Axelar Gateway contract.
- The application’s executable contract must be deployed on the destination contract.
- The application must be on one of Axelar's supported EVM chains. See [Chain names](../chain-names) for a list of EVM chains that have an Axelar Gateway deployed. The list is updated as new chains are added.

### Flow Architecture (in steps)

![gmp-diagram.png](/images/gmp-diagram.png)

### Steps

#### At the Source Chain

1. User (dApp) calls a `callContract` (or `callContractWithToken`) function on the Axelar Gateway contract to initiate a call. Once the call is initiated, the user can see its status at https://axelarscan.io/gmp/[txHash] or programmatically track it via the [[AxelarJS SDK](../axelarjs-sdk/token-transfer-dep-addr)].
2. User prepays the relayer gas fee on the source chain to Axelar’s Gas Services contract.
3. The call enters the Axelar Gateway from the source chain.

#### At the Axelar network
4. Axelar network confirms the call and converts the paid gas from the source chain’s native tokens to the destination chain’s native token.  

#### At the destination chain
5. The call is approved and emerges from the Axelar Gateway on the destination chain.
6. The executor service relays and executes the approved call to the application’s Axelar Executable interface.

Suppose the paid gas (step 2) is insufficient to relay the transfer to the application interface (step 6). In that case, the user can resubmit a new gas amount via a new transaction to the Gas Services contract from the source chain or manually execute it (either on UI via the [Axelarscan block explorer](https://axelarscan.io) or our SDK) by paying gas from the destination chain.

