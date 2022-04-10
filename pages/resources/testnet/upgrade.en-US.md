# Testnet network upgrade: 2022-mar-08

Validator instructions for 2022-mar-08 testnet upgrade to axelar-core `v0.15.0`.

1. Validators please vote for the upgrade proposal via

```bash
axelard tx gov vote 4 yes --from validator
```

2. Wait for the proposed upgrade block (1060850). Your node will panic at that block height. Stop your node after chain halt.

```bash
pkill -f 'axelard start'
pkill -f 'axelard vald-start'
pkill -f tofnd
```

3. Backup the state and keys. Example with default path `~/.axelar_testnet`:

```bash
cp -r ~/.axelar_testnet ~/.axelar_testnet-lisbon-3-upgrade-0.15
```

4. Restart your node with the new v0.15.0 build.

Example using join scripts in [axelarate-community git repo](https://github.com/axelarnetwork/axelarate-community):

```bash
# in axelarate-community repo
git checkout main
git pull
KEYRING_PASSWORD="pw-1" ./scripts/node.sh -n testnet
KEYRING_PASSWORD="pw-1" TOFND_PASSWORD="pw-2" ./scripts/validator-tools-host.sh -n testnet
```

The join scripts should automatically pull the new binary from [Testnet resources](../testnet). Or you can add the flag `-a v0.15.0` to force a specific version.