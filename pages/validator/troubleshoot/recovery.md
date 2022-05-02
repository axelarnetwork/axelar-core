# Recover from secret data

import Callout from 'nextra-theme-docs/callout'

[TODO revise]

## TODO: build tofnd from source?

[TODO: new section on building tofnd?]
```
tofnd binary release is only available for linux-amd64
For other platforms, build your own from the tofnd repo and place it at: /Users/gus/.axelar_testnet/bin/tofnd.
```

# OLD

This document describes the steps necessary to ensure that a validator node can be restored in case its state is lost. In order to achieve this, it is necessary that the following data is safely backed up:

* Tendermint validator key
* Axelar validator mnemonic
* Axelar proxy mnemonic
* Tofnd mnemonic

Besides the data described above, it will also be necessary to retrieve the *recovery data* associated with all the key shares that the validator was responsible for maintaining.

## Recovering an Axelar node

In order to restore the Tendermint key and/or the Axelar validator key used by an Axelard node, you can use the `--tendermint-key` and `--validator-mnemonic` flags with `join/join-testnet.sh` as follows:

```bash
./join/join-testnet.sh --tendermint-key /path/to/tendermint/key/ --validator-mnemonic /path/to/axelar/mnemonic/
```

If you are using the binary, you can add the flags to the binary script, similar to join-testnet.sh, for example:
```bash
./join/join-testnet-with-binaries.sh --tendermint-key /path/to/tendermint/key/ --validator-mnemonic /path/to/axelar/mnemonic/
```

## Recovery data

The recovery data is stored on chain, and enables a validator to recover key shares it created.
To obtain the recovery data for those key shares, you need to find out the corresponding key IDs first.
To query the blockchain for these key IDs - and assuming that the Axelar validator account has already been restored - attach a terminal to the node's container and perform the command:

```bash
axelard q tss key-shares-validator $(axelard keys show validator --bech val -a)
```
```yaml
- key_chain: Bitcoin
  key_id: btc-master
  key_role: KEY_ROLE_MASTER_KEY
  num_total_shares: "5"
  num_validator_shares: "1"
  snapshot_block_number: "23"
  validator_address: axelarvaloper1mx627hm02xa8m57s0xutgjchp3fjhrjwp2dw42
- key_chain: Bitcoin
  key_id: btc-secondary
  key_role: KEY_ROLE_SECONDARY_KEY
  num_total_shares: "5"
  num_validator_shares: "1"
  snapshot_block_number: "56"
  validator_address: axelarvaloper1mx627hm02xa8m57s0xutgjchp3fjhrjwp2dw4
```

In this example, the validator participated in generating the keys with ID `btc-master` and `btc-secondary`.
With the help of the key IDs, you can now retrieve the recovery data for the keys:

```bash
axelard q tss recover $(axelard keys show validator --bech val -a) btc-master btc-secondary --output json > recovery.json
```

The command above will fetch the recovery info for the aforementioned keys and store it to a `recovery.json` file.
This file will contain the data necessary to perform share recovery.

## Recovering the vald process

In order to restore the Axelar proxy key used by the Vald process, you can use the `--validator-mnemonic` flag with `join/launch-validator-tools.sh` as follows:

```bash
./join/join-testnet.sh --proxy-mnemonic /path/to/axelar/mnemonic/
```

## Recovering Tofnd state

If you want to reset your tofnd (e.g. on a new machine, after unexpected data loss, etc), you will have to recover your tofnd state. Tofnd's state consists of the following:
1. **your private key**: Internal tofnd key used to encrypt your recovery data. This private key is derived from a mnemonic that is generated automatically when tofnd is executed for the first time on your machine. You should have stored this mneminic safely, since it is the only passphrase that can be used to recover your key shares.
2. **your key shares**: Data that is generated when you participate into a keygen and is used to perform sign.

Each time you participated in a keygen, your key shares were encrypted and stored on the blockchain. This means that you can easily fetch your shares, but you must have your private key (i.e. launch tofnd with your mnemonic) to successfully decrypt them.

### Running tofnd in a containerized environment

In order to restore tofnd's private key and your key shares, you can use `join/launch-validator-tools.sh` with the `--tofnd-mnemonic` and `--recovery-info` flags with as follows:

```bash
./join/join-testnet.sh --tofnd-mnemonic <mnemonic file> ---recovery-info <recover json file>
```

1. `<mnemonic file>`: A file that contains your mnemonic passphrase
2. `<recover json file>`: The recovery information in json format you receive by executing
    ```
    axelard q tss recover $(axelard keys show validator --bech val -a) btc-master btc-secondary --output json > recovery.json
    ```
    after attaching to your validator container (see section [Recover Data](#Recovery_Data)).

### Running tofnd as binary

If you are running a tofnd binary, follow the steps below:
1. Create your recovery json file from your vald process (see section [Recovery Data](#Recovery_Data))
2. Copy the json recovery file to `~/.axelar_testnet/.vald/recovery.json`
3. Navigate to the directory of your tofnd binary.
4. Create a folder under the name `.tofnd/`.
5. Create a file `.tofnd/import` that contains your mnemonic passphrase.
6. Execute tofnd in *import* mode:
    ```bash
    ./tofnd -m import
    ```
    The output should be similar to the following:
    ```
    tofnd listen addr 0.0.0.0:50051, use ctrl+c to shutdown
    Importing mnemonic
    kv_manager cannot open existing db [.tofnd/kvstore/mnemonic]. creating new db
    kv_manager cannot open existing db [.tofnd/kvstore/shares]. creating new db
    Mnemonic successfully added in kv store
    ```
7. Restart vald
8. You should now see the following in your tofnd logs:
    ```
    Recovering keypair for party X ...
    Finished recovering keypair for party X
    Recovery completed successfully!
    ```

Tofnd has now re-created the private key that is derived from your mnemonic into tofnd's internal database, fetched your shares from the blockchain, decrypted them using your private key, and finally stored them at its internal tofnd database. Once recoverred, can remove your mnemonic file you used, as it is no longer needed.

<Callout type="error" emoji="☠️">
  Danger: Remember to still keep your mnemonic stored at an offline, secure place for future recoveries.
</Callout>