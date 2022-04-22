# Fantom

import Callout from 'nextra-theme-docs/callout'

Set up your Fantom Opera node.

## Prerequisites

- [Setup your Axelar validator](/validator/setup)
- Minimum hardware requirements: 4 vCPU+, 100GB+ free storage space.
- MacOS or Ubuntu 18.04+
- Build-essential packages
- Golang
- [Official Documentation](https://docs.fantom.foundation/staking/run-a-read-only-node)

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

Please make sure you checkout the [latest release tag](https://github.com/Fantom-foundation/go-opera/tags). In this tutorial we are using `1.1.0-rc.4`.

```bash
git clone https://github.com/Fantom-foundation/go-opera.git
cd go-opera/
git checkout release/1.1.0-rc.4
make
```

### 2. Download the genesis file

```bash
cd build/
wget https://opera.fantom.network/testnet.g
```

### 3. Start the Fantom Opera node

In this guide we are using `tmux` to run the `opera-go` process in the background. In case you don't have tmux installed, then you can do so with:

```bash
apt install tmux
```

Now create a new session called `fantom`:

```bash
tmux new -s fantom
```

Once you're inside the newly created tmux session, start the `opera-go`:

```bash
./opera --genesis testnet.g --http --http.addr=0.0.0.0 --http.vhosts="*" --http.corsdomain="*" --ws --ws.origins="*"
```

Your node will now start to synchronize with the network. It will take several hours before the node is fully synced.

<Callout type="error" emoji="🔥">
  Important: To detach from your current tmux session and keep it running in the background, use `CTRL + B D`. If you want to re-attach to the existing session, then use `tmux attach-session -t fantom`
</Callout>

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
