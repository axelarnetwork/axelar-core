# You need server with following requierements

8-Core (16-Thread) / 8GBRAM DDR4 / 500GB SSD

## 1. Upgrade your server
```
sudo apt update && sudo apt upgrade -y
```
```
sudo apt install make clang pkg-config libssl-dev libclang-dev build-essential git curl ntp jq llvm tmux htop screen unzip -y
```
## 2. Install docker and docker-compose

Check the latest version of [docker-compose](https://github.com/docker/compose) and follow the guide
```
sudo apt install docker.io -y
```
```
git clone https://github.com/docker/compose
```
```
cd compose
git checkout v2.6.1
make
```
```
cd
mv compose/bin/docker-compose /usr/bin
docker-compose version
```
## 3. Install and synch Aurora node

Download installation from [official documentation](https://github.com/aurora-is-near/partner-relayer-deploy)
```
git clone https://github.com/aurora-is-near/partner-relayer-deploy
cd partner-relayer-deploy
```
For testnet purpose run 
```
./setup.sh testnet
```
For mainet purpose run
```
./setup.sh
```
### 4. Run Aurora node through docker-compose
To run docker, you must be in the folder `partner-relayer-deploy`
``` 
docker-compose up -d
```
To see logs use command
```
docker-compose logs -f
```
## 5. Create SSH keys on the Aurora Node
Please run the following command on the Aurora server (just press "Enter" for each question):
```
cd
```
```
ssh-keygen -t ed25519
```

### 6. Add public key on the Axelar Node
Once the key pair has been generated you need to create it on the Axelar Node `/root/.ssh/authorized_keys` file and copy content from the `/root/.ssh/id_ed25519.pub` file which was created on the Aurora Node:
```
mkdir .ssh
nano /root/.ssh/authorized_keys
```
After this we need to set the following permissions:
```
chmod 600 /root/.ssh/authorized_keys
```
When you perform these actions you should be able to connect from Aurora Node to Axelar Node without password

### 7. Create SSH tunnel between Aurora Node and Axelar Main Node
On the Aurora server please run the following command to create a tunnel to forwarding ports:
```
ssh -f -N root@X.X.X.X -R 8545:`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}:8545{{end}}' endpoint`
```
Please note `X.X.X.X` this is IP address of your Axelar Main Node

### 8. Check that necessary port is listening on the AxelarNode
On the Axelar Node please run the following command to make sure that port is listening:
```
netstat -atnp | grep 8545
tcp        0      0 0.0.0.0:8545            0.0.0.0:*               LISTEN      2306635/sshd: root
tcp6       0      0 :::8545                 :::*                    LISTEN      2306635/sshd: root
```

### 9. Create a script to run it automatically after reboot
Create directory
```
mkdir /root/work
```

Create tunnel.ssh file:
```
nano /root/work/tunnel.ssh
```
and add the following line into it:
```
ssh -f -N root@X.X.X.X -R 8545:`docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}:8545{{end}}' endpoint`
```
Please note `X.X.X.X` this is IP address of your Axelar Main Node

Then open crontab file and choose option 1:
```
crontab -e
```

and add this line to the end of it:

```
@reboot /root/work/tunnel.sh
```

so when the Aurora server will be rebooted the SSH tunnel will be created automatically

### 10. Connect your Aurora to Axelar
In order for Axelar Network to connect to your Aurora node, your rpc_addr should be exposed in this format:

`"http://127.0.0.1:8545"`

