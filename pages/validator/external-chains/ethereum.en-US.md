# Ethereum

import Callout from 'nextra-theme-docs/callout'

Set up your Ethereum Ropsten Testnet node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: CPU with 2+ cores, 4GB RAM, 200GB+ free storage space.
- MacOS or Ubuntu 18.04+
- [Official Documentation](https://geth.ethereum.org/docs/getting-started)

## Install Geth

In this guide we will be installing `Geth` with the built-in launchpad PPAs (Personal Package Archives) on Ubuntu. If you are on different OS, please refer to the [official Documentation](https://geth.ethereum.org/docs/getting-started).

##### 1. Enable launchpad repository

```bash
sudo add-apt-repository -y ppa:ethereum/ethereum
```

##### 2. Install the latest version of go-ethereum:

```bash
sudo apt-get update
sudo apt-get install ethereum
```

## Run `geth` through systemd

##### 1. Create systemd service file

After installation of `go-ethereum`, we are now ready to start the `geth` process but in order to ensure it is running in the background and auto-restarts in case of a server failure, we will setup a service file with systemd.

<Callout emoji="📝">
  Note: In the service file below you need to replace `$USER` and path to `geth`, depending on your system configuration.
</Callout>

```bash
sudo tee <<EOF >/dev/null /etc/systemd/system/geth.service
[Unit]
Description=Ethereum Ropsten Node
After=network.target

[Service]
User=$USER
Type=simple
ExecStart=/usr/bin/geth --ropsten --syncmode "snap" --http --http.vhosts "*" --http.addr 0.0.0.0
Restart=on-failure
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF
```

##### 2. Enable and start the `geth` service

```bash
sudo systemctl enable geth
sudo systemctl daemon-reload
sudo systemctl start geth
```

If everything was set-up correctly, your Ethereum node should now be starting the process of synchronization. This will take several hours, depending on your hardware. In order to check the status of the running service or follow logs, you can use:

```bash
sudo systemctl status geth
sudo journalctl -u geth -f
```

## Test your Ethereum RPC connection

Alternatively, you can now also use the Geth JavaScript console and check status of your node by attaching to your newly created `geth.ipc`. Don't forget to replace $USER and path, depending on your server configuration.

```bash
geth attach ipc:/home/$resources/.ethereum/ropsten/geth.ipc
eth.syncing

```

Once your node is fully synced, the output from above will say `false`. To test your Ethereum node, you can send an RPC request using `cURL`

```bash
curl -X POST http://localhost:8545 \
-H "Content-Type: application/json" \
--data '{"jsonrpc":"2.0","method":"eth_syncing","params":[],"id":1}'
```

If you are testing it remotely, please replace `localhost` with the IP or URL of your server.

#### EVM RPC endpoint URL

In order for Axelar Network to connect to your Ethereum node, your `rpc_addr` should be exposed in this format:

```bash
http://IP:PORT
```

Example:
`http://192.168.192.168:8545`
