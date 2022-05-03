# Manual setup

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

## Prerequisites

- Ubuntu (tested on 18.04)
- `sudo apt-get install wget liblz4-tool aria2 jq -y`

## Get Binaries

```bash
# create a temp dir for binaries
mkdir binaries && cd binaries

# get axelard, tofnd binaries and rename
wget https://github.com/axelarnetwork/axelar-core/releases/download/v0.17.1/axelard-linux-amd64-v0.17.1
wget https://github.com/axelarnetwork/tofnd/releases/download/v0.10.1/tofnd-linux-amd64-v0.10.1
mv axelard-linux-amd64-v0.17.1 axelard
mv tofnd-linux-amd64-v0.10.1 tofnd

# make binaries executable
chmod +x *

# move to usr bin
sudo mv * /usr/bin/

# clean up temp dir
cd .. && rmdir binaries

# check versions
axelard version
tofnd --help
```

## Generate keys

```bash
axelard keys add broadcaster
axelard keys add validator
tofnd -m create
```

Your `tofnd` secret mnemonic is in a file `.tofnd/export`. Save this mnemonic somewhere safe and delete the file `.tofnd/export`.

## Set environment variables

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{"echo export CHAIN_ID=axelar-dojo-1 >> $HOME/.profile"}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{"echo export CHAIN_ID=axelar-testnet-lisbon-3 >> $HOME/.profile"}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{"echo export CHAIN_ID=axelar-testnet-casablanca-1 >> $HOME/.profile"}
</CodeBlock>
}
]} />

```bash
echo export MONIKER=PUT_YOUR_MONIKER_HERE >> $HOME/.profile
VALIDATOR_OPERATOR_ADDRESS=`axelard keys show validator --bech val --output json | jq -r .address`
BROADCASTER_ADDRESS=`axelard keys show broadcaster --output json | jq -r .address`
echo export VALIDATOR_OPERATOR_ADDRESS=$VALIDATOR_OPERATOR_ADDRESS >> $HOME/.profile
echo export BROADCASTER_ADDRESS=$BROADCASTER_ADDRESS >> $HOME/.profile
```

<Callout type="warning" emoji="⚠️">
  Protect your keyring password: The following instructions instruct you to store your keyring plaintext password in a file on disk. This instruction is safe only if you can prevent unauthorized access to the file. Use your discretion---substitute your own preferred method for securing your keyring password.
</Callout>

Choose a secret `{KEYRING_PASSWORD}` and add the following line to `$HOME/.profile`:

```
echo export KEYRING_PASSWORD=PUT_YOUR_KEYRING_PASSWORD_HERE >> $HOME/.profile
```

Apply your changes

```bash
source $HOME/.profile
```

## Configuration setup

Initialize your Axelar node, fetch configuration, genesis, seeds.

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`axelard init $MONIKER --chain-id $CHAIN_ID
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/config.toml -O $HOME/.axelar/config/config.toml
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/app.toml -O $HOME/.axelar/config/app.toml
wget https://axelar-mainnet.s3.us-east-2.amazonaws.com/genesis.json -O $HOME/.axelar/config/genesis.json
wget https://axelar-mainnet.s3.us-east-2.amazonaws.com/seeds.txt -O $HOME/.axelar/config/seeds.txt

# enter seeds to your config.json file

sed -i.bak 's/seeds = \"\"/seeds = \"'$(cat $HOME/.axelar/config/seeds.txt)'\"/g' $HOME/.axelar/config/config.toml

# set external ip to your config.json file

sed -i.bak 's/external_address = \"\"/external_address = \"'"$(curl -4 ifconfig.co)"':26656\"/g' $HOME/.axelar/config/config.toml`} </CodeBlock> }
]} />

## Sync From Snapshot

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`axelard unsafe-reset-all URL=\`curl https://quicksync.io/axelar.json | jq -r '.[] |select(.file=="axelar-dojo-1-pruned")|.url'\`
echo $URL
cd $HOME/.axelar/
wget -O - $URL | lz4 -d | tar -xvf -
cd $HOME`} </CodeBlock> }
]} />

## Create services

Use `systemctl` to set up services for `axelard`, `tofnd`, `vald`.

### axelard

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/axelard.service
[Unit]
Description=Axelard Cosmos daemon
After=network-online.target

[Service]
User=$USER
ExecStart=/usr/bin/axelard start
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

cat /etc/systemd/system/axelard.service
sudo systemctl enable axelard
```

### tofnd

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/tofnd.service
[Unit]
Description=Tofnd daemon
After=network-online.target

[Service]
User=$USER
ExecStart=/usr/bin/sh -c 'echo $KEYRING_PASSWORD | tofnd -m existing -d $HOME/.tofnd'
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

cat /etc/systemd/system/tofnd.service
sudo systemctl enable tofnd
```

### vald

```bash
# TODO is --chain-id necessary?
sudo tee <<EOF >/dev/null /etc/systemd/system/vald.service
[Unit]
Description=Vald daemon
After=network-online.target
[Service]
User=$USER
ExecStart=/usr/bin/sh -c 'echo $KEYRING_PASSWORD | /usr/bin/axelard vald-start --validator-addr $VALIDATOR_OPERATOR_ADDRESS --log_level debug --chain-id $CHAIN_ID --from broadcaster'
Restart=on-failure
RestartSec=3
LimitNOFILE=4096

[Install]
WantedBy=multi-user.target
EOF

cat /etc/systemd/system/vald.service
sudo systemctl enable vald
```

## Start all services

Order of operations:

1. `axelard`: ensure it's fully synced before proceeding
2. `tofnd`: required for `vald`
3. `vald`

```bash
sudo systemctl daemon-reload
sudo systemctl restart axelard
sudo systemctl restart tofnd
sudo systemctl restart vald
```

## Check logs

```bash
# change log settings to persistent
sed -i 's/#Storage=auto/Storage=persistent/g' /etc/systemd/journald.conf
sudo systemctl restart systemd-journald

journalctl -u axelard.service -f -n 100
journalctl -u tofnd.service -f -n 100
journalctl -u vald.service -f -n 100
```

## Register broadcaster proxy

<Callout emoji="📝">
  Note: Fund your `validator` and `broadcaster` accounts before proceeding.
</Callout>

```bash
axelard tx snapshot register-proxy $BROADCASTER_ADDRESS --from validator --chain-id $CHAIN_ID
```

## Create validator

```bash
# set temporary variables for create-validator command
IDENTITY="YOUR_KEYBASE_IDENTITY"
AMOUNT=PUT_AMOUNT_OF_TOKEN_YOU_WANT_TO_DELEGATE
DENOM=uaxl

axelard tx staking create-validator --yes \
 --amount $AMOUNT$DENOM \
 --moniker $MONIKER \
 --commission-rate="0.10" \
 --commission-max-rate="0.20" \
 --commission-max-change-rate="0.01" \
 --min-self-delegation="1" \
 --pubkey="$(axelard tendermint show-validator)" \
 --from validator \
 -b block \
 --identity=$IDENTITY \
 --chain-id $CHAIN_ID
```

## Register external chains

See [Support external chains](../external-chains).
