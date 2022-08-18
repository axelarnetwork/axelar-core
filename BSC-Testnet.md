**Server requierements**

- 2 CPU cores
- 8 GB RAM
- 200GB SSD

**1. Upgrade your server**
```
sudo apt update && sudo apt upgrade -y
```
```
sudo apt install make clang pkg-config libssl-dev libclang-dev build-essential git curl ntp jq llvm tmux htop screen unzip -y
```
Install Go 1.18.3
```
wget https://golang.org/dl/go1.18.3.linux-amd64.tar.gz
```
```
sudo tar -C /usr/local -xzf go1.18.3.linux-amd64.tar.gz
```
```
cat <<EOF >> ~/.profile
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export GO111MODULE=on
export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin
EOF
source ~/.profile
go version
rm -rf go1.18.3.linux-amd64.tar.gz
```
**2. Install BSC tesnet node**
```
git clone https://github.com/bnb-chain/bsc
cd bsc
make geth
```
**3. Download the config files**

You can download the pre-build binaries from [release page](https://github.com/bnb-chain/bsc/releases/latest) or follow the instructions bellow
```
wget https://github.com/bnb-chain/bsc/releases/download/v1.1.11/geth_linux
```
# Testnet
```
wget https://github.com/bnb-chain/bsc/releases/download/v1.1.11/testnet.zip
unzip testnet.zip
```
# Mainnet
```
wget https://github.com/bnb-chain/bsc/releases/download/v1.1.11/mainnet.zip
unzip mainnet.zip
```
```
mv geth_linux /usr/bin/geth
chmod +x /usr/bin/geth
rm -rf testnet.zip
```
**4. Configure config.toml file**
```
nano config.toml
```
Inside the ``config.toml`` found line ``HTTPHost = "127.0.0.1"`` and change IP from ``127.0.0.1`` on your IP address.

Scroll down and delete following lines:
```
[Node.LogConfig]
FileRoot = ""
FilePath = "bsc.log"
MaxBytesSize = 10485760
Level = "info"
```
Save it by command ``CTRL+X,Y``

**5. Write state genesis localy**
```
geth --datadir node init genesis.json
```
**6. Configure systemd service file**
```
tee /etc/systemd/system/bscd.service > /dev/null <<EOF
[Unit]
Description=BSC
After=network-online.target
[Service]
User=root
ExecStart=/usr/bin/geth --config root/bsc/config.toml --datadir root/bsc/node --ws --ws.origins '*'
Restart=always
RestartSec=3
LimitNOFILE=10000
[Install]
WantedBy=multi-user.target
EOF
```
```
systemctl daemon-reload
systemctl start bscd
```
Checkec that node start synching
```
journalctl -u bscd -f -n 100
```
**7. Sync status**

Once your BSC node is fully synced, you can run a cURL request to see the status of your node: Please change ``YOUR_IP_ADDRESS`` on your IP.
```
curl -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "eth_syncing", "params":[]}' YOUR_IP_ADDRESS:8575
```
If the node is successfully synced, the output from above will print ``{"jsonrpc":"2.0","id":1,"result":false}``

**8. Connect your BSC to Axelar**

Axelar Network will be connecting to the EVM compatible ``Binance``, so your rpc_addr should be exposed in this format:

``http://IP:PORT``

# Testnet

Example: ``http://5.168.135.185:8575``


# Mainnet

Example: ``http://5.168.135.185:8545``
