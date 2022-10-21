# Example: Composable USDC

Circle has announced a [plan to support cross-chain transactions in native USDC](https://www.circle.com/en/pressroom/circle-enables-usdc-interoperability-for-developers-with-the-launch-of-cross-chain-transfer-protocol). Currently, it’s available on the Ethereum Goerli and Avalanche Fuji testnets. In this tutorial, we’ll learn how to build a cross-chain USDC dApp using Circle’s Cross-Chain Transfer Protocol (CCTP) and Axelar’s General Message Passing (GMP).

What that means is, users will be able to issue a single transaction with a GMP payload. On the backend, the application takes care of USDC bridging, plus any other action that the user wishes — as indicated in the payload. Axelar services, also working on the backend, can handle conversion and payment for destination-chain gas fees, so the user only has to transact once, using one gas token.

In this example, we will build a cross-chain swap dApp. It converts a native token from one chain to another chain, using native USDC as a routing asset. For example: send ETH to a contract on Ethereum Goerli testnet and receive AVAX on Avalanche Fuji testnet, or vice versa.

There are two parts we have to learn to achieve this:

1. Sending a native USDC token cross-chain.
2. Sending a swap payload cross-chain.

## Part 1: Sending a native USDC token cross-chain

There are three components from Circle that we’ll use in this part:

1. **MessageTransmitter** contract – to mint USDC at the destination chain.
2. **CircleBridge** contract  – to burn USDC at the source chain.
3. **Attestation API** – to retrieve attestation to be used for minting USDC at the destination chain.

Let’s take a look at how to implement this step-by-step:

1. Burn the given amount of **USDC** by calling the function `depositForBurn`at the `CircleBridge` contract. The example Solidity code is below. At this step, the contract does nothing except provide a function to burn the **USDC** in `_depositForBurn` function.

CrosschainNativeSwap.sol

```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.9;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ICircleBridge} from "./ICircleBridge.sol";
import "./StringAddressUtils.sol";

contract CrosschainNativeSwap {
    IERC20 public usdc;
    ICircleBridge public circleBridge;

    // mapping chain name to domain number;
    mapping(string => uint32) public circleDestinationDomains;
    bytes32 constant CHAIN_ETHEREUM = keccak256(abi.encodePacked("ethereum"));
    bytes32 constant CHAIN_AVALANCHE = keccak256(abi.encodePacked("avalanche"));

    constructor(address _usdc, address _circleBridge) {
        usdc = IERC20(_usdc);
        circleBridge = ICircleBridge(_circleBridge);
        circleDestinationDomains["ethereum"] = 0;
        circleDestinationDomains["avalanche"] = 1;
    }

    modifier isValidChain(string memory destinationChain) {
        require(
            keccak256(abi.encodePacked(destinationChain)) == CHAIN_ETHEREUM ||
            keccak256(abi.encodePacked(destinationChain)) == CHAIN_AVALANCHE,
            "Invalid chain"
        );
        _;
    }

    // Step 1: Burn USDC on the source chain with given amount
    function _depositForBurn(
        uint256 amount,
        string memory destinationChain,
        address recipient
    ) private isValidChain(destinationChain) {
        IERC20(address(usdc)).approve(address(circleBridge), amount);

        circleBridge.depositForBurn(
            amount,
            this.circleDestinationDomains(destinationChain),
            bytes32(uint256(uint160(recipient))),
            address(usdc)
        );
    }
}
```

ICircleBridge.sol

```solidity
// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.9;

interface ICircleBridge {
    // this event will be emitted when `depositForBurn` function is called.
    event MessageSent(bytes message);

    /**
    * @param _amount amount of tokens to burn
    * @param _destinationDomain destination domain
    * @param _mintRecipient address of mint recipient on destination domain
    * @param _burnToken address of contract to burn deposited tokens, on local
    domain
    * @return _nonce uint64, unique nonce for each burn
    */
    function depositForBurn(
        uint256 _amount,
        uint32 _destinationDomain,
        bytes32 _mintRecipient,
        address _burnToken
    ) external returns (uint64 _nonce);
}
```

That's it for the contract. We'll continue to add our business logic to it later in Part 2.

2. When the USDC is burned, the `CircleBridge` contract will emit a `MessageSent` event. An interface of the `MessageSent` event looks like this:

```solidity
event MessageSent(bytes message)
```

At this step, we’ll extract `message` from the transaction hash. The code snippet below provides an example of such logic.

constants.ts

```tsx
import { ethers } from "ethers";

export const MESSAGE_TRANSMITTER_ADDRESS = {
  ethereum: "0x40A61D3D2AfcF5A5d31FcDf269e575fB99dd87f7",
  avalanche: "0x52FfFb3EE8Fa7838e9858A2D5e454007b9027c3C",
};
export const PROVIDERS = {
  ethereum: new ethers.providers.WebSocketProvider(
    "wss://goerli.infura.io/ws/v3/INFURA_PROJECT_ID"
  ),
  avalanche: new ethers.providers.WebSocketProvider(
    "wss://api.avax-test.network/ext/bc/C/ws"
  ),
};
```

step2.ts

```tsx
import { ethers } from "ethers";
import { MESSAGE_TRANSMITTER_ADDRESS, PROVIDERS } from "./constant";

// Extract the `message` from the `MessageSent` event
const getMessageFromMessageSentEvent = (
  contract: ethers.Contract,
  txReceipt: ethers.providers.TransactionReceipt
) => {
  const eventLogs = txReceipt.logs;
  const messageSentEventId = ethers.utils.id("MessageSent(bytes)");
  for (const log of eventLogs) {
    if (log.topics[0] === messageSentEventId) {
      return contract.interface.parseLog(log).args.message;
    }
  }
  return null;
};

// Get message based on txHash
export async function getMessageFromTxHash(txHash: string, chain: string) {
  // Initialize MessageTransmitter contract
  const srcContract = new ethers.Contract(
    MESSAGE_TRANSMITTER_ADDRESS[chain],
    ["event MessageSent(bytes message)"],
    PROVIDERS[chain]
  );

  // Retrieves transaction receipt
  const txReceipt = await PROVIDERS[chain].getTransactionReceipt(txHash);

  // Retrives `message` from transaction receipt
  return getMessageFromMessageSentEvent(srcContract, txReceipt);
}
```

3. Call the Circle Attestation API to calculate the signature. Then, send a transaction to call `receiveMessage` function at the `MessageTransmitter` contract on the destination chain.

step3.ts

```tsx
import { ethers } from "ethers";
import { MESSAGE_TRANSMITTER_ADDRESS } from "./constant";
import { getMessageFromTxHash } from "./step2";

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

async function fetchAttestation(messageHash: string, maxAttempt = 10) {
  let attempt = 0;
  while (attempt < maxAttempt) {
    const _response = await fetch(
      `https://iris-api-sandbox.circle.com/attestations/${messageHash}`
    ).then((resp) => resp.json());

    if (_response?.status === "complete") {
      return _response?.attestation;
    }

    await sleep(5000);
    attempt++;
  }
}

async function retrieveUSDC(
  depositTxHash: string,
  depositChain: string,
  withdrawChain: string,
  signer: ethers.Signer
) {
  const messageTransmitterAddress = MESSAGE_TRANSMITTER_ADDRESS[withdrawChain];

  const contract = new ethers.Contract(
    messageTransmitterAddress,
    [
      "function receiveMessage(bytes memory _message, bytes calldata _attestation)",
    ],
    signer
  );

  // Retrieves the message by txHash
  const message = await getMessageFromTxHash(depositTxHash, depositChain);

  // Calculate message hash
  const messageHash = ethers.utils.solidityKeccak256(["bytes"], [message]);

  // Fetch attestation from Circle Attestation Service API.
  const attestation = await fetchAttestation(messageHash);

  // Call `receiveMessage` function to mint USDC to the recipient address
  if (attestation) {
    return contract
      .receiveMessage(message, attestation)
      .then((tx) => tx.wait());
  }
}
```

That's all about sending the USDC cross-chain. Next, let's try to integrate this with Axelar network to complete our cross-chain swap dApp.

## Part 2: Sending a swap payload cross-chain

In this part, we’ll add logic in our contract to send a payload cross-chain with Axelar network.

1. Upgrade our contract to include business logic and integrate with Axelar network.

CrosschainNativeSwap.sol

```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.9;

import {IAxelarForecallable} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarForecallable.sol";
import {IAxelarGasService} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarGasService.sol";
import {IAxelarGateway} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarGateway.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {ICircleBridge} from "./ICircleBridge.sol";
import "./StringAddressUtils.sol";

contract CrosschainNativeSwap is Ownable {
    IERC20 public usdc;
    ICircleBridge public circleBridge;
    IAxelarGasService public gasReceiver;
    IAxelarGateway public gateway;

    // mapping chain name => domain number;
    mapping(string => uint32) public circleDestinationDomains;
    // mapping destination chain name => destination contract address
    mapping(string => address) public siblings;

    bytes32 constant CHAIN_ETHEREUM = keccak256(abi.encodePacked("ethereum"));
    bytes32 constant CHAIN_AVALANCHE = keccak256(abi.encodePacked("avalanche"));

    constructor(
        address _usdc,
        address _gasReceiver,
        address _circleBridge
    ) Ownable() {
        usdc = IERC20(_usdc);
        circleBridge = ICircleBridge(_circleBridge);
        gasReceiver = IAxelarGasService(_gasReceiver);
        circleDestinationDomains["ethereum"] = 0;
        circleDestinationDomains["avalanche"] = 1;
    }

    modifier isValidChain(string memory destinationChain) {
        require(
            keccak256(abi.encodePacked(destinationChain)) == CHAIN_ETHEREUM ||
                keccak256(abi.encodePacked(destinationChain)) ==
                CHAIN_AVALANCHE,
            "Invalid chain"
        );
        _;
    }

    // Set address for this contract that deployed at another chain
    function addSibling(string memory chain_, address address_)
        external
        onlyOwner
    {
        siblings[chain_] = address_;
    }

    /**
     * @dev Swap native token to USDC, burn it, and send swap payload to AxelarGateway contract
     * @param destinationChain Name of the destination chain
     * @param srcTradeData Trade data for the first swap
     * @param destTradeData Trade data for the second swap
     * @param traceId Trace ID of the swap
     * @param fallbackRecipient Recipient address to receive USDC token if the swap fails
     * @param inputPos Position of the input token in destTradeData
     */
    function nativeTradeSendTrade(
        string memory destinationChain,
        bytes memory srcTradeData,
        bytes memory destTradeData,
        bytes32 traceId,
        address fallbackRecipient,
        uint16 inputPos
    ) external payable isValidChain(destinationChain) {
        // Swap native token to USDC
        (uint256 nativeSwapAmount, uint256 usdcAmount) = _trade(srcTradeData);

        _depositForBurn(
            usdcAmount,
            destinationChain,
            this.siblings(destinationChain)
        );

        // encode the payload to send to the sibling contract
        bytes memory payload = abi.encode(
            destTradeData,
            usdcAmount,
            traceId,
            fallbackRecipient,
            inputPos
        );

        // Pay gas to AxelarGasReceiver contract with native token to execute the sibling contract at the destination chain
        gasReceiver.payNativeGasForContractCall{
            value: msg.value - nativeSwapAmount
        }(
            address(this),
            destinationChain,
            AddressToString.toString(this.siblings(destinationChain)),
            payload,
            msg.sender
        );

        // Send all information to AxelarGateway contract.
        gateway.callContract(
            destinationChain,
            AddressToString.toString(this.siblings(destinationChain)),
            payload
        );
    }

    function _depositForBurn(
        uint256 amount,
        string memory destinationChain,
        address recipient
    ) private isValidChain(destinationChain) {
        IERC20(address(usdc)).approve(address(circleBridge), amount);

        circleBridge.depositForBurn(
            amount,
            this.circleDestinationDomains(destinationChain),
            bytes32(uint256(uint160(recipient))),
            address(usdc)
        );
    }

    function _tradeSrc(bytes memory tradeData)
        internal
        returns (bool success, uint256 amount)
    {
        (uint256 amountIn, address router, bytes memory data) = abi.decode(
            tradeData,
            (uint256, address, bytes)
        );
        (success, ) = router.call{value: amountIn}(data);
        return (success, amountIn);
    }

    function _trade(bytes memory tradeData1)
        private
        returns (uint256 amount, uint256 burnAmount)
    {
        // Calculate remaining usdc token in the contract
        uint256 preTradeBalance = tokenBalance(address(usdc));

        // Swap native token to USDC
        (bool success, uint256 _nativeSwapAmount) = _tradeSrc(tradeData1);

        // Revert if trade failed
        require(success, "TRADE_FAILED");

        // Calculate amount of USDC token swapped. This is the amount to be burned at the source chain.
        uint256 _usdcAmount = tokenBalance(address(usdc)) - preTradeBalance;

        // Return amount of native token swapped and amount of USDC token to be burned
        return (_nativeSwapAmount, _usdcAmount);
    }
}
```

There’s a lot of new code added here. Let’s try to understand it, step by step:

**Step 1** — Add the `gasReceiver` variable and initialize it in the constructor. This handles destination-chain gas token conversion and fee payment, so the user need not transact more than once.

**Step 2** — Add the `addSibling` function so the admin can define identical contract addresses at the other chains.

**Step 3** — Add the `nativeTradeSendTrade` function. The client will send a transaction to call this function. This is the most important function in our contract. Here are the implementation details:

- Swap native token to USDC with low-level contract call.
- Burn the USDC with the function that we implemented in Part 1.
- Construct the swap payload to send to the **AxelarGateway** contract. The payload will be relayed by Axelar Relayer service to the destination contract. The destination contract address is defined by `addSibling` function as mentioned in **Step 2**.
- Pay gas to the **AxelarGasService** contract with the native token. The required amount will be calculated off-chain by using [AxelarJS-SDK](https://docs.axelar.dev/dev/axelarjs-sdk/intro) on the client side. See more information about it [here](https://docs.axelar.dev/dev/axelarjs-sdk/axelar-query-api#estimategasfee).
- Send `destinationChain`, `destinationContractAddress`and `payload` to the **AxelarGateway** contract.

2. Upgrade a contract to extend **IAxelarForecallable** interface and override `_execute` function, so the contract can be called by Axelar Executor service at the destination chain. Note that the code snippet below includes newly added code to make things easier to read.

CrosschainNativeSwap.sol

```solidity
// SPDX-License-Identifier: MIT
pragma solidity 0.8.9;

import {IAxelarForecallable} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarForecallable.sol";
import {IAxelarGasService} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarGasService.sol";
import {IAxelarGateway} from "@axelar-network/axelar-cgp-solidity/contracts/interfaces/IAxelarGateway.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {ICircleBridge} from "./ICircleBridge.sol";
import "./StringAddressUtils.sol";

contract CrosschainNativeSwap is IAxelarForecallable, Ownable {
    IERC20 public usdc;

    event SwapSuccess(bytes32 indexed traceId, uint256 amount, bytes tradeData);

    event SwapFailed(
        bytes32 indexed traceId,
        uint256 amount,
        address refundAddress
    );

    constructor(address _usdc, address _gateway)
        IAxelarForecallable(_gateway)
        Ownable()
    {
        usdc = IERC20(_usdc);
    }

    // ** To make things easier to read. Previous implementation details are skipped **

    function _refund(
        bytes32 traceId,
        uint256 amount,
        address recipient
    ) internal {
        SafeERC20.safeTransfer(IERC20(address(usdc)), recipient, amount);
        emit SwapFailed(traceId, amount, recipient);
    }

    // This function will be called by Axelar Executor service.
    function _execute(
        string memory, /*sourceChain*/
        string memory, /*sourceAddress*/
        bytes calldata payload
    ) internal override {
        // Step 1: Decode payload
        (
            bytes memory tradeData,
            uint256 usdcAmount,
            bytes32 traceId,
            address fallbackRecipient,
            uint16 inputPos
        ) = abi.decode(payload, (bytes, uint256, bytes32, address, uint16));

        // Step 2: This hack puts the amount in the correct position.
        assembly {
            mstore(add(tradeData, inputPos), usdcAmount)
        }

        (address srcToken, , address router, bytes memory data) = abi.decode(
            tradeData,
            (address, uint256, address, bytes)
        );

        // Step 3: Approve USDC to the router contract
        IERC20(srcToken).approve(router, usdcAmount);

        // Step 3: Swap USDC to native token
        (bool swapSuccess, ) = router.call(data);

        // Step 3: If swap failed, refund USDC to the user.
        if (!swapSuccess)
            return _refund(traceId, usdcAmount, fallbackRecipient);

        // Step 4: Emit success event so that our application can be notified.
        emit SwapSuccess(traceId, usdcAmount, tradeData);
    }
}
```

This upgrade mainly implements the `_execute` function to perform a final swap at the destination chain before sending it to the recipient wallet. Here are the implementation details:

**Step 1:** The function decodes `payload` to retrieve all information it needs for swap.

**Step 2**: This is a bit hacky way to correct the amount in `tradeData` bytes before the swap.

**Step 3**: Approve USDC to the router contract and call the swap function, and refund if it fails.

**Step 4**: Finally, emit `SwapSuccess` event if the swap is successful.

And we’re done! Here is the [demo](https://www.youtube.com/watch?v=RyQkEcM1nKE) that communicates with the completed contract.
<br />

<iframe width="560" height="315" src="https://www.youtube.com/embed/RyQkEcM1nKE" title="YouTube video player" frameBorder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowFullScreen></iframe>

### Resources

- CrosschainNativeSwap contract: [link](https://github.com/axelarnetwork/crosschain-usdc-demo/blob/live/hardhat/contracts/CrosschainNativeSwap.sol)
- Running Crosschain USDC Example with Hardhat: [link](https://github.com/axelarnetwork/crosschain-usdc-demo/tree/live/hardhat)
- Axelar Cross-chain USDC Demo: [link](https://circle-crosschain-usdc.vercel.app/)

### About Circle

Circle is a global financial technology firm that enables businesses of all sizes to harness the power of digital currencies and public blockchains for payments, commerce and financial applications worldwide. Circle is the issuer of USD Coin (USDC), one of the fastest growing dollar digital currencies powering always-on internet-native commerce and payments. Today, Circle's transactional services, business accounts, and platform APIs are giving rise to a new generation of financial services and commerce applications that hold the promise of raising global economic prosperity for all through the frictionless exchange of financial value. Additionally, Circle operates SeedInvest, a leading startup fundraising platform in the U.S. Learn more at [https://circle.com](https://c212.net/c/link/?t=0&l=en&o=3502185-1&h=961378907&u=https%3A%2F%2Fcircle.com%2F&a=https%3A%2F%2Fcircle.com).

### About Axelar

Axelar delivers secure cross-chain communication. That means dApp users can interact with any asset, any application, on any chain, with one click. You can think of it as Stripe for Web3. Developers interact with a simple API atop a permissionless network that routes messages and ensures network security via proof-of-stake consensus.

Axelar has raised capital from top-tier investors, including Binance, Coinbase, Dragonfly Capital and Polychain Capital. Partners include major proof-of-stake blockchains, such as Avalanche, Cosmos, Ethereum, Polkadot and others. Axelar’s team includes experts in distributed systems/cryptography and MIT/Google/Consensys alumni; the co-founders, Sergey Gorbunov and Georgios Vlachos, were founding team members at Algorand.

More about Axelar: [Website](https://axelar.network) | [GitHub](https://github.com/axelarnetwork/axelar-local-gmp-examples) | [Discord](https://discord.com/invite/aRZ3Ra6f7D) | [Twitter](https://twitter.com/axelarcore).
