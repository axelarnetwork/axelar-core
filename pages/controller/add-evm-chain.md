# Add a new EVM chain

import Callout from 'nextra-theme-docs/callout'
import Bleed from 'nextra-theme-docs/bleed'
import Tabs from '../../components/tabs'
import CodeBlock from '../../components/code-block'

Learn how to add a new EVM chain to the Axelar network.

<Callout emoji="💡">
  Example: For clarity, this article targets the Avalanche EVM chain. Substitute your own EVM chain as desired.
</Callout>

## Prerequisites

- Prerequisites for [Controller operations](../controller)
- Axelar validators with enough stake have configured their nodes to support the new EVM chain as per [Support external chains](../validator/external-chains/overview).
- You will deploy smart contracts to the new EVM chain---you need enough native tokens to pay gas fees on that chain. Example: if deploying to Avalanche then you need AVAX tokens, etc.

## Add the new EVM chain

Prepare a JSON file `evm-chain.json` with information about the new EVM chain. Examples of files that have already been used to add EVM chains to Axelar:

- [Mainnet examples](https://github.com/axelarnetwork/axelarate-community/tree/main/resources/mainnet/evm-chain-params)
- [Testnet examples](https://github.com/axelarnetwork/axelarate-community/tree/main/resources/testnet/evm-chain-params)

The params file above specifies the contract bytecode in base64 to be used for the gateway, token, and the burnable contracts. These can be obtained by compiling the appropriate version of the [contracts repo](https://github.com/axelarnetwork/solidity-cgp-gateway).

Make Axelar aware of the new EVM chain.

```bash
axelard tx evm add-chain Avalanche multisig path/to/evm-chain.json --from controller
```

Call a validator vote to confirm new EVM chain.

```bash
axelard tx evm confirm-chain avalanche --from controller
```

Each validator votes `true` or `false` according to whether it is configured to support the new EVM chain as per [Support external chains](../validator/external-chains/overview).

Optional: check your logs for messages of the form `EVM chain Avalanche confirmation result is true`.

Now that the new EVM chain has been added, validators need to register as maintainers for the new chain as per [Support external chains](../validator/external-chains/overview).

If a large enough portion of validator stake has registered as a maintainer of the new EVM chain then that chain can be activated by a special transaction signed by the Axelar governance account. Contact the Axelar team to request such a transaction.

## Deploy the gateway contract

The new EVM chain is active but cannot be used until its gateway contract is deployed.

Generate initial validator keys for the new chain and rotate them into use.

```bash
axelard tx tss start-keygen --id avax-master-genesis --key-role master --from controller
axelard tx tss start-keygen --id avax-secondary-genesis --key-role secondary --from controller
axelard tx tss rotate avalanche master avax-master-genesis --from controller
axelard tx tss rotate avalanche secondary avax-secondary-genesis --from controller
```

Optional: verify that keys were successfully rotated.

```bash
axelard q evm address avalanche --key-role master
axelard q evm address avalanche --key-role secondary
```

Register external governance keys for the new chain.

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
axelard tx tss register-external-keys avalanche --from controller \
 --key avax-external-1:03e4da05dd2c4d1a75567fff2ade93de82ccca8701689ce42da40cebd4cc7a2423 \
 --key avax-external-2:03dccf1720dafc44d6e47635b8f0e2705bd57346acce1f18238580461fd3c900ce \
 --key avax-external-3:0383afc1b42f1dae34649ab70c4c3d67aa86db89fc1842cd697e3c2a574b433ab1 \
 --key avax-external-4:02ad55f4054d47a13cfe2583693bf63a8f299ca33da936f7372a38070cbf5dbc93 \
 --key avax-external-5:02e25f07aca8971908b7489b54d809401c34d1d5a817b521234ced5b75c056f2fd \
 --key avax-external-6:023f39b9bfcead2854bab63f02880487553430a475ee0f3783c52ef98927cb37d7 \
 --key avax-external-7:03e78bbe19444fe98a77b45c340998437fb902572747f8e44ea99b23dc1106e0d2 \
 --key avax-external-8:0203ce85d1564ce9203b45ad6b93511c8daaa0927f31f3e8d53f18b51afc3f7a27
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
axelard tx tss register-external-keys avalanche --from controller \
 --key avax-external-1:041bb2e070cdd8490500f673136f95d80bf4eb9ac6a85fe8fede6070515d75dfd51a7187318aeb17eb53e711a8ccc0939bbda30ac67836969b64422ef6831a6e2a \
 --key avax-external-2:040626a1032fe1f76deb3b4f0cb5c68cc0e29264102111077691478d74eafcdc4c3938bb712b7d8cd15f9c72261432be19217f02e445d4caeb09d7a1abe793642e \
 --key avax-external-3:04d8400ac69c6c919d6963a5da6403ba750f5d7859b0c00f1a6b9a2ce9cd663bf8a20af0351f0e6dbc7bc41040c06156d02b78f25fd6b9b54db03bd53e812b8577 \
 --key avax-external-4:04ee60261b7fd4084b271618cefef4f9a4093338a1dfbeae2f4a18366a53a07dd0657f31fa38c2739885fbc9ee1e83e25b3fcda6581be8f82400c7d0a18eb79070 \
 --key avax-external-5:04aebaae5d1c63bf527331ee9a75dbc727420f51bc557a78dd0d9a0227c42dad2d821c9a3b24ca75b558b3c333cb8e4acaa4fb174a07ebd8268c6a43b83e04ffdc \
 --key avax-external-6:04900ccd93432b25f4758f8d702bfd91e6192b0931345baafa50328185300a4cd0e392398612f5394fd6026f06fa73fb42a4f43ced2fa5ef326e5d658fd90113d3 \
 --key avax-external-7:04761c872fd2c9c501e75ba3ef8fd65ed4d0f1e7ba60901f758b26645ee3621256a65df63f10b009f1f0e458e7cdb69737098cc30e99cf41887d7adc24c9492729 \
 --key avax-external-8:044202188712caa9c047caaa01eb2a97f631b5ffb0ac1e2de40609c89137d7992a07d7b6c48cb69eee0323855377582a5601008a1190ca6c2b37316c2aead28bfd
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
axelard tx tss register-external-keys avalanche --from controller \
 --key avax-external-1:041bb2e070cdd8490500f673136f95d80bf4eb9ac6a85fe8fede6070515d75dfd51a7187318aeb17eb53e711a8ccc0939bbda30ac67836969b64422ef6831a6e2a \
 --key avax-external-2:040626a1032fe1f76deb3b4f0cb5c68cc0e29264102111077691478d74eafcdc4c3938bb712b7d8cd15f9c72261432be19217f02e445d4caeb09d7a1abe793642e \
 --key avax-external-3:04d8400ac69c6c919d6963a5da6403ba750f5d7859b0c00f1a6b9a2ce9cd663bf8a20af0351f0e6dbc7bc41040c06156d02b78f25fd6b9b54db03bd53e812b8577 \
 --key avax-external-4:04ee60261b7fd4084b271618cefef4f9a4093338a1dfbeae2f4a18366a53a07dd0657f31fa38c2739885fbc9ee1e83e25b3fcda6581be8f82400c7d0a18eb79070 \
 --key avax-external-5:04aebaae5d1c63bf527331ee9a75dbc727420f51bc557a78dd0d9a0227c42dad2d821c9a3b24ca75b558b3c333cb8e4acaa4fb174a07ebd8268c6a43b83e04ffdc \
 --key avax-external-6:04900ccd93432b25f4758f8d702bfd91e6192b0931345baafa50328185300a4cd0e392398612f5394fd6026f06fa73fb42a4f43ced2fa5ef326e5d658fd90113d3 \
 --key avax-external-7:04761c872fd2c9c501e75ba3ef8fd65ed4d0f1e7ba60901f758b26645ee3621256a65df63f10b009f1f0e458e7cdb69737098cc30e99cf41887d7adc24c9492729 \
 --key avax-external-8:044202188712caa9c047caaa01eb2a97f631b5ffb0ac1e2de40609c89137d7992a07d7b6c48cb69eee0323855377582a5601008a1190ca6c2b37316c2aead28bfd
</CodeBlock>
}
]} />

The gateway contract can now be deployed on the new EVM chain.

Get the gateway contract bytecode.

```bash
axelard q evm bytecode avalanche gateway-deployment
```

Deploy the above bytecode to a smart contract on the EVM chain.

<Callout emoji="💡">
  Tip: Deploy the gateway contract however you wish. One option is to use [version 5 of MyEtherWallet](https://v5.myetherwallet.com/) with Metamask as suggested in the video [Onboarding Avalanche to the Axelar Network](https://www.youtube.com/watch?v=iZgqneh7s88).
  <Bleed>
    <iframe
      width="80%"
      height="430"
      src="https://www.youtube.com/embed/iZgqneh7s88"
      frameBorder="0"
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
      allowFullScreen
      style={{ margin: "0 0 1.5rem 2rem" }}
    />
  </Bleed>
</Callout>

Note the following from your deployment transaction:

- `{EVM_GATEWAY_TX_HASH}` transaction ID on the EVM chain
- `{EVM_GATEWAY_ADDR}` address of the new gateway smart contract on the EVM chain

Wait until the transaction `{EVM_GATEWAY_TX_HASH}` has received enough block confirmations on the EVM chain. (This number was set in the `confirmation_height` in the file `evm-chain.json` when you executed `add-chain`.)

Call a validator vote to confirm gateway deployment.

```bash
axelard tx evm confirm-gateway-deployment avalanche {EVM_GATEWAY_TX_HASH} {EVM_GATEWAY_ADDR} --from controller
```

Optional: check your logs for messages of the form `Avalanche gateway confirmation result is true`.
