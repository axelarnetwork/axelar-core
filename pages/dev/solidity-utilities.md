# Solidity Utilities

To ease cross chain developement we have provided some solidity utilites. These can be found [here](https://github.com/axelarnetwork/axelar-utils-solidity). Each of these as well as their proposed use is detailed below.

## Constant Address Deployer

Creating a cross chain dApp will often require the same contract to be deployed on multiple chains. Furthermore it is useful to know each address of this contract on each chain, either to know where to send remote contract calls, or where to trust remote contract calls from, often both.
If we can guarantee that the contracts in question will be deployed at the same address on each network then the above is trivial.
This can be achieved by deploying each contract from the same address with the same nonce at each network, or by using [`create2`](https://eips.ethereum.org/EIPS/eip-1014). For this purpose we deployed [`ConstAddressDeployer`](https://github.com/axelarnetwork/axelar-utils-solidity/blob/main/src/ConstAddressDeployer.sol) at `0x617179a15fEAa53Fa82ae80b0fc3E85b7359a748` on every EVM testnet and mainnet that is supported by Axelar, and we plan on deploying it on future supported testnets and mainnets too. `ConstAddressDeployer` exposes the following functions

- `deployedAddress(bytes bytecode, address sender, bytes32 salt)`: Calculates the address of contracts that has been/will be deployed with a certain bytecode and salt, by a certain sender.
- `deploy(bytes bytecode, bytes32 salt)`: Deploys a contract with a certain bytecode and salt.
- `deployAndInit(bytes bytecode, bytes32 salt, bytes init)`: Deploys a contract with a certain bytecode and salt and runs `deployedContract.call(init)` afterwards. Use in case you need constructor arguments that are not constant across chains, as different constructor arguments result in different bytecodes.

The above can be used directly, but we also provide some scripts. Simply use `require('@axelar-network/axelar-utils-solidity')` to access:

- `async estimateGasForDeploy(contractJson, args = [])`: Estimates the gas needed to deploy a contract with a certain `contractJson` and `args`
- `async estimateGasForDeployAndInit(contractJson, args = [], initArgs = [])`: Estimates the gas needed to deploy a contract with a certain `contractJson` and `args`, and to have the `init(...initArgs)` called as part of the deployment.
- `async deployContractConstant(deployer, wallet, contractJson, key, args = [])`: Uses `deployer`, an ethers Contract pointing to a `ConstAddressDeployer`, a `wallet` with native currencty, the `contractJson` to deploy, a string `key` which will be hashed to get the `salt`, and the constructor `args` to make a deployment.
- `async deployAndInitContractConstant(deployer, wallet, contractJson, key, args = [], initArgs = [])`: same as above but uses `deployAndInit` (with `initArgs`) to instead of `deploy`.

## String and Address utilities

Axelar uses the string representation of addresses (42 characters) for EVM addresses (20 bytes). It is often useful to convert between the two. [`StringAddressUtils.sol`](https://github.com/axelarnetwork/axelar-utils-solidity/blob/main/src/StringAddressUtils.sol) is a library that can be used to do so. Below see an example of how to use it.

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