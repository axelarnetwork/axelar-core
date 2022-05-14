# Moonbeam

import Callout from 'nextra-theme-docs/callout'


Set up your Moonbeam Mainnet or Testnet (Moonbase-Alpha) node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 8+ cores CPU , 16GB+ RAM, 500GB+ free storage space.
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://docs.moonbeam.network/node-operators/networks/run-a-node/)


## Install Moonbeam / Moonbase Alpha

### 1. Download compiled binary

Download the [latest release binary](https://github.com/PureStake/moonbeam/tags) from PureState. In this tutorial, we are using `v0.23.0`

```bash
wget https://github.com/PureStake/moonbeam/releases/download/v0.23.0/moonbeam

```
### 2. Create a service account and copy the binary

```bash
#For Mainnet (Moonbeam) use:
adduser moonbeam_service --system --no-create-home
mkdir /var/lib/moonbeam-data
mv ./moonbeam /var/lib/moonbeam-data # assumption: ./moonbeam is your downloaded binary
sudo chown -R moonbeam_service /var/lib/moonbeam-data

#For Testnet (Moonbase Alpha) use:
adduser moonbeam_service --system --no-create-home
mkdir /var/lib/alphanet-data
mv ./moonbeam /var/lib/alphanet-data # assumption: ./moonbeam is your downloaded binary
sudo chown -R moonbeam_service /var/lib/alphanet-data
```

### 3. Create the systemd service file

After the installation of `moonbeam`, we are now ready to start the node, but to ensure it is running in the background and auto-restarts in case of a server failure, we will set up a service file using systemd.

<Callout emoji="📝">
  Note: In the service file below you need to replace `"YOUR-NODE-NAME"` and replace `50% RAM in MB` for 50% of the actual RAM your server has (Example: `--db-cache 16000` if your server has 32GB RAM).
  
  If you are connecting to Testnet instead (Moonbase Alpha), you will also need to change the path to `/var/lib/alphanet-data/` and add `--chain alphanet`.
</Callout>

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/moonbeam.service
[Unit]
Description="Moonbeam systemd service"
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=on-failure
RestartSec=10
User=moonbeam_service
SyslogIdentifier=moonbeam
SyslogFacility=local7
KillSignal=SIGHUP
ExecStart=/var/lib/moonbeam-data/moonbeam \
     --port 30333 \
     --rpc-port 9933 \
     --ws-port 9944 \
     --unsafe-rpc-external \
     --rpc-cors all \
     --execution wasm \
     --wasm-execution compiled \
     --pruning=archive \
     --state-cache-size 1 \
     --db-cache <50% RAM in MB> \
     --base-path /var/lib/moonbeam-data \
     --chain moonbeam \
     --name "YOUR-NODE-NAME" \
     -- \
     --port 30334 \
     --rpc-port 9934 \
     --ws-port 9945 \
     --execution wasm \
     --pruning=1000 \
     --name="YOUR-NODE-NAME (Embedded Relay)"

[Install]
WantedBy=multi-user.target
EOF
```

### 4. Enable and start the `moonbeam` service

```bash
sudo systemctl enable moonbeam.service
sudo systemctl start moonbeam.service
```

If everything was set-up correctly, your Moonbeam node should now be starting the process of synchronization. This will take several hours, depending on your hardware. To check the status of the running service or to follow the logs, use:

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

Axelar Network will be connecting to the EVM compatible `Moonbeam`, so your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:9933`
