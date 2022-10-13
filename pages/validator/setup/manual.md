# Manual setup

import Callout from 'nextra-theme-docs/callout'
import Markdown from 'markdown-to-jsx'
import Tabs from '../../../components/tabs'
import CodeBlock from '../../../components/code-block'

## Prerequisites

- Ubuntu (tested on 18.04 and 20.04)
- `sudo apt-get install wget liblz4-tool aria2 jq -y`

## Get Binaries

Check the appropriate version for the network accordingly:

- [Mainnet](/resources/mainnet)
- [Testnet](/resources/testnet)
- [Testnet-2](/resources/testnet-2)

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`AXELARD_RELEASE=v0.26.0-patch
TOFND_RELEASE=v0.10.1`}
</CodeBlock>
},
{
title: "Testnet",
content: <CodeBlock language="bash">
{`AXELARD_RELEASE=v0.26.0-patch
TOFND_RELEASE=v0.10.1`}
</CodeBlock>
},
{
title: "Testnet-2",
content: <CodeBlock language="bash">
{`AXELARD_RELEASE=v0.17.3
TOFND_RELEASE=v0.10.1`}
</CodeBlock>
}
]} />

```bash
# verify correct versions
echo $AXELARD_RELEASE $TOFND_RELEASE

# create a temp dir for binaries
cd $HOME
mkdir binaries && cd binaries


# get axelard, tofnd binaries and rename
wget https://github.com/axelarnetwork/axelar-core/releases/download/$AXELARD_RELEASE/axelard-linux-amd64-$AXELARD_RELEASE
wget https://github.com/axelarnetwork/tofnd/releases/download/$TOFND_RELEASE/tofnd-linux-amd64-$TOFND_RELEASE
mv axelard-linux-amd64-$AXELARD_RELEASE axelard
mv tofnd-linux-amd64-$TOFND_RELEASE tofnd

# make binaries executable
chmod +x *

# move to usr bin
sudo mv * /usr/bin/

# get out of binaries directory
cd $HOME

# check versions
axelard version
tofnd --help
```

## Generate keys

To create new keys

```bash
axelard keys add broadcaster
axelard keys add validator
tofnd -m create
```

To recover exsiting keys

```bash
axelard keys add broadcaster --recover
axelard keys add validator --recover
tofnd -m import
# type your desired keyring password and enter mnemonics when prompted
```

Your `tofnd` secret mnemonic is in a file do ` cat $HOME/.tofnd/export` to check.
Save this mnemonic somewhere safe and delete the file by ` rm $HOME/.tofnd/export`.

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

```bash
# it's recommended to manually edit the file and add it
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
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/resources/mainnet/seeds.toml -O $HOME/.axelar/config/seeds.toml

# set external ip to your config.json file
sed -i.bak 's/external_address = \"\"/external_address = \"'"$(curl -4 ifconfig.co)"':26656\"/g' $HOME/.axelar/config/config.toml`}
</CodeBlock>
},
{ title: "Testnet", content: <CodeBlock language="bash">
{`axelard init $MONIKER --chain-id $CHAIN_ID
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/config.toml -O $HOME/.axelar/config/config.toml
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/app.toml -O $HOME/.axelar/config/app.toml
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/resources/testnet/genesis.json -O $HOME/.axelar/config/genesis.json
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/resources/testnet/seeds.toml -O $HOME/.axelar/config/seeds.toml

# set external ip to your config.json file
sed -i.bak 's/external_address = \"\"/external_address = \"'"$(curl -4 ifconfig.co)"':26656\"/g' $HOME/.axelar/config/config.toml`}
</CodeBlock>
},
{ title: "Testnet-2", content: <CodeBlock language="bash">
{`axelard init $MONIKER --chain-id $CHAIN_ID
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/config.toml -O $HOME/.axelar/config/config.toml
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/configuration/app.toml -O $HOME/.axelar/config/app.toml
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/resources/testnet-2/genesis.json -O $HOME/.axelar/config/genesis.json
wget https://raw.githubusercontent.com/axelarnetwork/axelarate-community/main/resources/testnet-2/seeds.toml -O $HOME/.axelar/config/seeds.toml

# set external ip to your config.json file
sed -i.bak 's/external_address = ""/external_address = "'"$(curl -4 ifconfig.co)"':26656"/g' $HOME/.axelar/config/config.toml`}
</CodeBlock>
}
]} />

## Sync From Snapshot

<Tabs tabs={[
{
title: "Mainnet",
content: <CodeBlock language="bash">
{`axelard tendermint unsafe-reset-all
URL=\`curl -L https://quicksync.io/axelar.json | jq -r '.[] |select(.file=="axelar-dojo-1-pruned")|.url'\`
echo $URL
cd $HOME/.axelar/
wget -O - $URL | lz4 -d | tar -xvf -
cd $HOME`}
</CodeBlock>
},
{ title: "Testnet", content: <CodeBlock language="bash">
{`axelard tendermint unsafe-reset-all
URL=\`curl -L https://quicksync.io/axelar.json | jq -r '.[] |select(.file=="axelartestnet-lisbon-3-pruned")|.url'\`
echo $URL
cd $HOME/.axelar/
wget -O - $URL | lz4 -d | tar -xvf -
cd $HOME`}
</CodeBlock>
},
{ title: "Testnet-2", content: <CodeBlock language="bash">
{`axelard tendermint unsafe-reset-all
URL="https://snapshots.bitszn.com/snapshots/axelar/axelar.tar"
echo $URL
cd $HOME/.axelar/data
wget -O - $URL | tar -xvf -
cd $HOME`} </CodeBlock>
}
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
# change log settings to persistent if not already
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
IDENTITY="YOUR_KEYBASE_IDENTITY" # optional
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

## Upgrade Process

```bash
cd $HOME
cd binaries
# put the tag/version that we are upgrading to
# set versions: the below is an example and the numbers should be replaced
# AXELARD_RELEASE=v0.17.1
# TOFND_RELEASE=v0.10.1

AXELARD_RELEASE=<GIVE_VERSION>
TOFND_RELEASE=<GIVE_VERSION>  # if we are upgrading tofnd too

echo $AXELARD_RELEASE $TOFND_RELEASE

# get axelard, tofnd binaries and rename
wget https://github.com/axelarnetwork/axelar-core/releases/download/$AXELARD_RELEASE/axelard-linux-amd64-$AXELARD_RELEASE
# if we are upgrading tofnd too
wget https://github.com/axelarnetwork/tofnd/releases/download/$TOFND_RELEASE/tofnd-linux-amd64-$TOFND_RELEASE


mv axelard-linux-amd64-$AXELARD_RELEASE axelard # if we are upgrading tofnd too
mv tofnd-linux-amd64-$TOFND_RELEASE tofnd

# make binaries executable
chmod +x *

# move to usr bin
sudo mv * /usr/bin/


# check versions
axelard version
echo $AXELARD_RELEASE
# axelard version and echo $RELEASE should have same tag/version


# check versions
tofnd --help
echo $TOFND_RELEASE
# tofnd version and echo $TOFND_RELEASE should have same tag/version

# restart services
sudo systemctl restart axelard
sudo systemctl restart tofnd # if we are upgrading tofnd too
sudo systemctl restart vald

# check logs
journalctl -u axelard.service -f -n 100
journalctl -u tofnd.service -f -n 100
journalctl -u vald.service -f -n 1000
```
