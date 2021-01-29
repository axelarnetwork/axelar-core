#!/bin/sh
set -e

fileCount() {
  find "$1" -maxdepth 1 ! -iname ".*" ! -iname "$(basename "$1")" | wc -l
}

modify() {
  sed "s/^$1 =.*/$1 = $2/g" /root/.axelard/config/config.toml >/root/.axelard/config/config.toml.tmp &&
  mv /root/.axelard/config/config.toml.tmp /root/.axelard/config/config.toml
}

prepareCli() {
  #### Client
  echo "Modifying config"
  axelarcli config keyring-backend test
  axelarcli config chain-id axelar
  axelarcli config output json
  axelarcli config indent true
  axelarcli config trust-node true

  echo "Adding global treasury account as a token well"
  axelarcli keys add treasury --recover <$TREASURY_MNEMONIC_PATH

  echo "Adding local validator account"
  axelarcli keys add validator

  echo "Adding local broadcaster account"
  axelarcli keys add broadcaster
}

prepareD(){
  echo "Initializing application"
  axelard init "${1}" --chain-id axelar

  if $GENESIS; then
    ### Daemon
    echo "Setting up genesis file"
    axelard add-genesis-account "$(axelarcli keys show validator -a)" 100000000stake
    axelard add-genesis-account "$(axelarcli keys show broadcaster -a)" 100000000000stake
    axelard add-genesis-account "$(axelarcli keys show treasury -a)" 100000000000stake

    axelard gentx --name validator --keyring-backend test

    axelard collect-gentxs

    axelard validate-genesis

    if ! $LOCAL; then
      ipAddress=$(wget -qO - https://api.ipify.org)
    else
      ipAddress=$(hostname -i)
    fi

    echo "$(axelard tendermint show-node-id)@$ipAddress:26656" > "/root/shared/peers.txt"
    cp "/root/.axelard/config/genesis.json" "/root/shared/genesis.json"
  fi
  echo "Preparations done"
}

if [ ! -d "/root/shared" ]; then
  mkdir "/root/shared"
fi

if [ ! -d "/root/.axelard" ]; then
  mkdir "/root/.axelard"
fi

if [ ! -d "/root/.axelarcli" ]; then
  mkdir "/root/.axelarcli"
fi

if [ "$(fileCount /root/.axelarcli)" -eq 0 ]; then
  prepareCli "$(hostname)"
fi

if [ "$(fileCount /root/.axelard)" -eq 0 ]; then
  prepareD "$(hostname)"
fi

cp $CONFIG_PATH "/root/.axelard/config/config.toml"
if ! $GENESIS; then
  until [ -f "/root/shared/genesis.json" ] && [ -f "/root/shared/peers.txt" ] ; do
    echo "Waiting for genesis.json and peers.txt to be accessible in /root/shared/"
    sleep 1
  done

  modify "persistent_peers" "\"$(head -n 1 "/root/shared/peers.txt")\""
  cp "/root/shared/genesis.json" "/root/.axelard/config"
fi

if [ -n "$TSSD_HOST" ]; then
  TSSD_HOST_SWITCH="--tssd-host $TSSD_HOST"
else
  TSSD_HOST_SWITCH=""
fi

if [ "$DLV_AXELARCLI_CONTINUE" = true ]; then
  CLI_CONTINUE="--continue"
fi

if [ "$DLV_AXELARD_CONTINUE" = true ]; then
  D_CONTINUE="--continue"
fi

if [ "$START_REST" = true ]; then
  # $DEBUG not set or set to false
  if [ -z "$DEBUG" ] || ! $DEBUG; then
    # REST endpoint must be bound to 0.0.0.0 for availability on docker host
    axelarcli rest-server --chain-id=axelarcli --laddr=tcp://0.0.0.0:1317 --node tcp://0.0.0.0:26657 --unsafe-cors &
  else
    dlv --listen=:2347 --headless=true --api-version=2 $CLI_CONTINUE --accept-multiclient exec /root/axelarcli -- rest-server --chain-id=axelarcli --laddr=tcp://0.0.0.0:1317 --node tcp://0.0.0.0:26657 --unsafe-cors &
  fi
fi

# $DEBUG not set or set to false
if [ -z "$DEBUG" ] || ! $DEBUG; then
  axelard start $TSSD_HOST_SWITCH
else
  dlv --listen=:2345 --headless=true $D_CONTINUE --api-version=2 --accept-multiclient exec /root/axelard -- start $TSSD_HOST_SWITCH
fi
