# Moonbeam

import Callout from 'nextra-theme-docs/callout'

Set up your Moonbeam (Moonbase-Alpha) Testnet node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 8+ cores CPU , 16GB+ RAM, 150GB+ free storage space.
- MacOS or Ubuntu 18.04+
- Rust (If you are compiling the binary manually)
- [Official Documentation](https://docs.moonbeam.network/node-operators/networks/run-a-node/)

## Install required dependencies

In order to compile the `moonbeam` binary by yourself, you first need to setup the required dependencies. In this guide we will be using the release binary instead, so you can skip this step.

##### 1. Setup Rust

```bash
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
source $HOME/.cargo/env
cargo build --release
```

## Install Moonbase Alpha

### 1. Download compiled binary

Download the [latest release binary](https://github.com/PureStake/moonbeam/tags) from PureState.

### 2. Create service account and copy the binary

```bash
adduser moonbase_service --system --no-create-home
mkdir /var/lib/alphanet-data
cp ./moonbeam /var/lib/alphanet-data # assumption: ./moonbeam is your downloaded binary
sudo chown -R moonbase_service /var/lib/alphanet-data
```

### 3. Create the systemd service file

After installation of `moonbase-alpha`, we are now ready to start the node but in order to ensure it is running in the background and auto-restarts in case of a server failure, we will setup a service file using systemd.

<Callout emoji="📝">
  Note: In the service file below you need to replace `"YOUR-NODE-NAME"` and path to `moonbeam`, depending on your system configuration.
</Callout>

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/moonbeam.service
[Unit]
Description="Moonbase Alpha systemd service"
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=10
User=moonbase_service
SyslogIdentifier=moonbase
SyslogFacility=local7
KillSignal=SIGHUP
ExecStart=/var/lib/alphanet-data/moonbeam \
     --port 30333 \
     --rpc-port 9933 \
     --ws-port 9944 \
     --unsafe-rpc-external \
     --unsafe-ws-external \
     --rpc-cors all \
     --execution wasm \
     --wasm-execution compiled \
     --pruning=archive \
     --state-cache-size 1 \
     --base-path /var/lib/alphanet-data \
     --chain alphanet \
     --name "YOUR-NODE-NAME" \
     --db-cache 64000 \
     -- \
     --port 30334 \
     --rpc-port 9934 \
     --ws-port 9945 \
     --execution wasm \
     --pruning=archive \
     --name="YOUR-NODE-NAME"

[Install]
WantedBy=multi-user.target
EOF
```

### 4. Enable and start the `moonbeam` service

```bash
sudo systemctl enable moonbeam.service
sudo systemctl start moonbeam.service
```

If everything was set-up correctly, your Moonbeam node should now be starting the process of synchronization. This will take several hours, depending on your hardware. In order to check the status of the running service or follow logs, you can use:

```bash
sudo systemctl status moonbeam.service
sudo journalctl -u moonbeam.service -f
```

## Test your Moonbeam RPC connection

Once your `Moonbeam` node is fully synced, you can run a cURL request to see the status of your node:

```bash
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "eth_syncing", "params":[]}' localhost:9933
```

If the node is successfully synced, the output from above will print `{"jsonrpc":"2.0","result":false,"id":1}`

#### EVM RPC endpoint URL

Axelar Network will be connecting to the EVM compatible `Moonbean`, so your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:9933`
