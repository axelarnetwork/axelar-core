# Fantom

import Callout from 'nextra-theme-docs/callout'

Set up your Fantom Mainnet or Testnet Opera node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 4 vCPU+, 16GB RAM, 1.5TB+ free storage space.
- MacOS or Ubuntu 18.04+
- Build-essential packages
- Golang
- [Official Documentation](https://docs.fantom.foundation/node/run-a-read-only-node)

## Install required dependencies

In order to build the `go-opera`, you first need to install all of the required dependencies.

### 1. Update and install `build-essential`

```bash
sudo apt-get update
sudo apt-get -y upgrade
sudo apt-get install -y build-essential
```

### 2. Install `golang`

Install the [latest version of golang](https://go.dev/doc/install).

## Install Go-Opera

### 1. Checkout and build go-opera

Please make sure you checkout the [latest release tag](https://github.com/Fantom-foundation/go-opera/releases).

```bash
git clone https://github.com/Fantom-foundation/go-opera.git
cd go-opera/
git checkout release/[latest version]
make
```

### 2. Download the genesis file

```bash
cd build/
wget https://download.fantom.network/mainnet-109331-pruned-mpt.g

#Testnet
# wget https://download.fantom.network/testnet-6226-pruned-mpt.g
```

### 3. Create systemd service file

After installation of `go-opera`, we are now ready to start the process but in order to ensure it is running in the background and auto-restarts in case of a server failure, we will setup a service file with systemd.

<Callout emoji="📝">
  Note: In the service file below you need to replace `$USER` and path to `opera`, depending on your system configuration.
</Callout>

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/fantom.service
[Unit]
Description=Fantom Node
After=network.target

[Service]
User=$USER
Type=simple
ExecStart=/root/go-opera/build/opera --genesis /root/go-opera/build/mainnet-109331-pruned-mpt.g --identity <your_name> --cache 8096 --http --http.addr 0.0.0.0 --http.corsdomain '*' --http.vhosts "*" --http.api "eth,net,web3" 
Restart=on-failure
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```
<Callout type="error" emoji="⚠️">
 If you would like to run a node on the Testnet instead, you need to replace `--genesis /root/go-opera/build/testnet-6226-pruned-mpt.g` in the configuration above.
</Callout>

### 4. Enable and start the `fantom` service

```bash
sudo systemctl enable fantom
sudo systemctl daemon-reload
sudo systemctl start fantom
```

If everything was set-up correctly, your Fantom node should now be starting the process of synchronization. This will take several hours, depending on your hardware. To check the status of the running service or to follow the logs, you can use:

```bash
sudo systemctl status fantom
sudo journalctl -u fantom -f
```
Your node will now start to synchronize with the network. It will take several hours before the node is fully synced.

## Test your Fantom RPC connection

Once your node is fully synced, you can run a cURL request to see the status of your node:

```bash
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "eth_syncing", "params":[]}' localhost:18545
```

If the node is successfully synced, the output from above will print `{"jsonrpc":"2.0","id":1,"result":false}`

### EVM RPC endpoint URL

In order for Axelar Network to connect to your Fantom node, your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:18545`
