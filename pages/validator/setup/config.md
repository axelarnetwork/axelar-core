# Configure companion processes

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

Axelar validators need two companion processes called `vald` and `tofnd`.

## Download binaries and initialize configuration

Similar to [Node configuration](../../node/config-node), run `setup-validator.sh` to download and configure `vald` and `tofnd` binaries. In the [`axelerate-community`](https://github.com/axelarnetwork/axelarate-community) repo do:

```bash
./scripts/setup-validator.sh -n [mainnet|testnet|testnet-2]
```

The binary `tofnd` is placed in your `${AXELARD_HOME}/bin` directory. The binary `vald` is actually part of `axelard`.

### Verifying Binaries

By default, the binary signatures are downloaded and the binary is verified using the [axelardev](https://keybase.io/axelardev) PGP key. To verify the binary manually, you can download the PGP signature and verify using the following commands:

```bash
curl https://keybase.io/axelardev/key.asc | gpg --import
gpg --verify [axelard_binary_signature_path] [axelard_binary_path]
```

On github the signatures are attached to the releases. To download the signatures from the axelar-releases AWS Bucket, you can add `.asc` to the end of the binary URL. For example, if the path of the binary is:
```bash
https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/v0.26.0/axelard-darwin-arm64-v0.26.0
```
The path for the signature will be:
```bash
https://axelar-releases.s3.us-east-2.amazonaws.com/axelard/v0.26.0/axelard-darwin-arm64-v0.26.0.asc
```

## Directory structure of a running validator

Later, after you've launched your companion processes and created your validator, your directory structure should look like:

```
.axelar
├── bin
│   ├── axelard -> /Users/gus/.axelar/bin/axelard-vx.y.z
│   ├── axelard-vx.y.z
│   ├── tofnd -> /Users/gus/.axelar/bin/tofnd-va.b.c
│   └── tofnd-va.b.c
├── config
│   ├── app.toml
│   ├── config.toml
│   ├── genesis.json
│   ├── node_key.json
│   ├── priv_validator_key.json
│   ├── priv_validator_state.json
│   └── seeds.toml
├── data
├── logs
├── tofnd
└── vald
    └── state.json
```

Relevant files:

- `priv_validator_key.json`, `node_key.json` : Created when you first launched your node as described in [Basic node management](../../node/basic).
- `priv_validator_state.json` : Last block height signed by the validator. This prevents double signing old blocks. If it’s content is `{}`, it’ll start signing from a synced node’s latest block
- `vald/state.json` : State file specifying the last block processed. If it’s not present, or is too old, vald starts from the latest block instead.
