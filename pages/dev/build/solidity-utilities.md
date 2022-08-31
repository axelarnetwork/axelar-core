# Solidity utilities

To facilitate cross-chain development, we have provided some Solidity utilites. These can be found [here](https://github.com/axelarnetwork/axelar-utils-solidity). Each of these is described below, along with their proposed use.

## Constant Address Deployer

Creating a cross-chain dApp will often require the same contract to be deployed on multiple chains.
Furthermore, it is useful to know each address of this contract on each chain, either to know where to send remote contract calls, or where to trust remote contract calls from -- often both.
If we can guarantee that the contracts in question will be deployed at the same address on each network, then the above is trivial.
This can be achieved by deploying each contract from the same address with the same nonce at each network, or by using [`create2`](https://eips.ethereum.org/EIPS/eip-1014).
For this purpose, we deployed [`ConstAddressDeployer`](https://github.com/axelarnetwork/axelar-utils-solidity/blob/main/contracts/ConstAddressDeployer.sol) at
`0x98b2920d53612483f91f12ed7754e51b4a77919e` on every EVM testnet and mainnet that is supported by Axelar.
We plan on deploying it on future supported testnets and mainnets, too.

`ConstAddressDeployer` exposes the following functions:
- `deployedAddress(bytes bytecode, address sender, bytes32 salt)`: calculates the address of contracts that has been/will be deployed with a certain bytecode and salt, by a certain sender.
- `deploy(bytes bytecode, bytes32 salt)`: deploys a contract with a certain bytecode and salt.
- `deployAndInit(bytes bytecode, bytes32 salt, bytes init)`: deploys a contract with a certain bytecode and salt, and runs `deployedContract.call(init)` afterwards. Use in case you need constructor arguments that are not constant across chains, as different constructor arguments result in different bytecodes.

The above can be used directly, but we also provide some scripts. Simply use `require('@axelar-network/axelar-utils-solidity')` to access:

- `async estimateGasForDeploy(contractJson, args = [])`: estimates the gas needed to deploy a contract with a certain `contractJson` and `args`
- `async estimateGasForDeployAndInit(contractJson, args = [], initArgs = [])`: estimates the gas needed to deploy a contract with a certain `contractJson` and `args`, and to have the `init(...initArgs)` called as part of the deployment.
- `async deployContractConstant(deployer, wallet, contractJson, key, args = [])`: uses `deployer`, an Ethers.js contract pointing to:
    - `ConstAddressDeployer`, a `wallet` with native currency. 
    - The `contractJson` to deploy. 
    - A string `key`, which will be hashed to get the `salt`. 
    - The constructor `args` to make a deployment.
- `async deployAndInitContractConstant(deployer, wallet, contractJson, key, args = [], initArgs = [])`: same as above, but uses `deployAndInit` (with `initArgs`), instead of `deploy`.

## String and address utilities

Axelar uses the string representation of addresses (42 characters) for EVM addresses (20 bytes). It is often useful to convert between the two. [`StringAddressUtils.sol`](https://github.com/axelarnetwork/axelar-utils-solidity/blob/main/contracts/StringAddressUtils.sol) is a library that can be used to do so. Below, see an example showing how to use it.

```solidity
import { StringToAddress, AddressToString } from '../StringAddressUtils.sol';

contract Test {
    using AddressToString for address;
    using StringToAddress for string;

    function addressToString(address address_) external pure returns (string memory) {
        return address_.toString();
    }

    function stringToAddress(string calldata string_) external pure returns (address) {
        return string_.toAddress();
    }
}
```