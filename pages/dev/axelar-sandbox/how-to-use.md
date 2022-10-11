# Getting Familiar with Sandbox UI (Alpha version)

The Axelar Sandbox UI is divided into four panels.

**Solidity editor (top left) —** Write your Solidity smart contract code here. Currently, only two smart contracts can be compiled. We'll allow developer to add more smart contracts soon.

**JavaScript editor (top right)** — Write the JavaScript code that deploys and interacts with the smart contracts.

**JavaScript console output (bottom right) —** This is where you'll see the output of the JavaScript execution as well as any errors that occurred.

**Transactions overview (bottom left)** — a list of transactions that occurred as a result of your Solidity and JavaScript code, including with transaction data such as event logs, function calls, and function arguments

## How to use it?

Let's begin by executing our examples. So, here's how the website looks like:

![Guide-1](/images/sandbox-guide-1.png)

1. We have some built-in examples at the bottom left that the developer can play with without modifying anything.

![Guide-2](/images/sandbox-guide-2.png)

2. `hello-world` is the default example. It shows how to update destination contract states from the source chain contract. The example can be run against testnet or simulated environment.

![Guide-3](/images/sandbox-guide-3.png)

3. Clicks at the **Execute** button in the upper right corner to run an example.

![Guide-4](/images/sandbox-guide-4.png)

4. Next, the solidity contracts will be compiled; if successful, the javascript file will be executed; otherwise, the error message will be displayed. The "Output" panel displays all Javascript log messages that have already been executed.

![Guide-5](/images/sandbox-guide-5.png)

5. The "**Transactions**" panel displays all transaction info sent from the "**sender**" and the "**relayer**" wallet. The relayer wallet is the wallet that interacts with your destination contract while the sender does that with the source contract. The transaction information includes:

- `Transaction hash`
- `Transaction status`
- `Source address`
- `Destination address`
- `Function name`
- `Function arguments`
- `Emitted event data`

The image below shows how it looks like.

![Guide-6](/images/sandbox-guide-6.png)

## Using Javascript to interact with smart contracts

You can use [ethers.js](https://github.com/ethers-io/ethers.js/) to deploy and interact with your smart contracts using Javascript. `ethers` variable is automatically injected into the js code editor.

### Available JavaScript global variables

You can use the global variables listed below anywhere in your JavaScript code.

#### **Chain**

An enum representation of string value of chain name. The snippet below expands all available chains.

```ts
Chain.ETHEREUM; // ethereum
Chain.AVALANCHE; // avalanche
Chain.FANTOM; // fantom
Chain.MOONBEAM; // moonbeam
Chain.POLYGON; // polygon
```

#### **$contracts**

A map of contracts representing compiled contracts including `abi` and `bytecode`. The map is indexed by the smart contract name defined in the Solidity file.

```ts
// access to MessageSender contract's abi
$contracts["MessagerSender"].abi;
// access to MessageReceiver contract's bytecode
$contracts["MessageReceiver"].bytecode;
```

#### **$abis**

A map of abis for commonly used contracts.

```ts
const { erc20, gateway, gasReceiver } = $abis;
```

#### **$getSigner**

A signer account for given chain.

```ts
const ethereumSigner = await $getSigner(Chain.ETHEREUM);
const avalancheSigner = await $getSigner(Chain.AVALANCHE);
```

#### **$chains**

A map of chains info which uses `Chain` as a key. The snippet below expands all available chain values.

```ts
const { rpcUrl, gateway, gasReceiver, tokens } = $chains[Chain.ETHEREUM];
const { aUSDC } = tokens;

// instantiate a provider
const provider = new ethers.providers.JsonRpcProvider(rpcUrl);

// instantiate a gateway contract
const gatewayContract = new ethers.Contract(gateway, $abis.gateway, provider);

// instantiate a gas receiver contract
const gasReceiverContract = new ethers.Contract(
  gasReceiver,
  $abis.gasReceiver,
  provider
);

// instantiate a wrapped usdc contract.
const aUsdcContract = new ethers.Contract(aUSDC, $abis.erc20, provider);
```

## Deploy a Contract

The following code is the example of how to deploy a smart contract that we wrote in “**MessageSender.sol**” file.

```ts
const signer = await $getSigner(Chain.MOONBEAM);
const contractFactory = new ethers.ContractFactory(
  $contracts["MessageSender"].abi,
  $contracts["MessageSender"].bytecode,
  srcSigner
);
const contract = await srcContractFactory.deploy(
  $chains.moonbeam.gateway,
  $chains.moonbeam.gasReceiver
);
```

That's all we have for now. The Axelar Sandbox is still in its early stages of development. We intend to improve it further in the future. Keep an eye out. 👀
