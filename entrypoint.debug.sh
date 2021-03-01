#!/bin/sh
set -e

fileCount() {
  find "$1" -maxdepth 1 ! -iname ".*" ! -iname "$(basename "$1")" | wc -l
}

addPeers() {
  sed "s/^seeds =.*/seeds = \"$1\"/g" "$D_HOME_DIR/config/config.toml" >"$D_HOME_DIR/config/config.toml.tmp" &&
  mv $D_HOME_DIR/config/config.toml.tmp $D_HOME_DIR/config/config.toml
}

prepareCli() {
  #### Client
  echo "Setting up config for CLI"
  axelarcli config keyring-backend "$KEYRING_BACKEND"
  axelarcli config chain-id "$CHAIN_ID"
  axelarcli config output json
  axelarcli config indent true
  axelarcli config trust-node true
}

isCliPrepared() {
  if [ -f "$CLI_HOME_DIR/config/config.toml" ]; then
    return 0
  fi

  return 1
}

isGenesisInitialized() {
  if [ -f "$D_HOME_DIR/config/genesis.json" ]; then
    return 0
  fi

  return 1
}

initGenesis() {
  if [ -n "$INIT_SCRIPT" ] && [ -f "$INIT_SCRIPT" ]; then
    echo "Running script at $INIT_SCRIPT to create the genesis file"
    "$INIT_SCRIPT" "$(hostname)" "$CHAIN_ID"
  else
    axelard init $(hostname) --chain-id $CHAIN_ID
  fi
}

CLI_HOME_DIR="$HOME_DIR/.axelarcli"
D_HOME_DIR="$HOME_DIR/.axelard"

if ! isCliPrepared; then
  prepareCli
fi

if ! isGenesisInitialized; then
  initGenesis
fi

if ! isGenesisInitialized; then
  echo "Missing genesis file"
  exit 1
fi

if [ -n "$CONFIG_PATH" ] && [ -f "$CONFIG_PATH" ]; then
  cp "$CONFIG_PATH" "$D_HOME_DIR/config/config.toml"
fi

if [ -n "$PEERS_FILE" ]; then
  PEERS=$(cat "$PEERS_FILE")
  addPeers "$PEERS"
fi

if [ -n "$TOFND_HOST" ]; then
  TOFND_HOST_SWITCH="--tofnd-host $TOFND_HOST"
else
  TOFND_HOST_SWITCH="" # An axelar-core node without tofnd is a non-validator
fi

if [ "$REST_CONTINUE" = true ]; then
  CONTINUE_SWITCH="--continue"
else
  CONTINUE_SWITCH=""
fi

if [ "$START_REST" = true ]; then
    # REST endpoint must be bound to 0.0.0.0 for availability on docker host
    dlv --listen=:2347 --headless=true --api-version=2 $CONTINUE_SWITCH --accept-multiclient exec \
      /usr/local/bin/axelarcli -- rest-server --chain-id=axelarcli --laddr=tcp://0.0.0.0:1317 --node tcp://0.0.0.0:26657 --unsafe-cors &
fi

if [ "$CORE_CONTINUE" = true ]; then
  CONTINUE_SWITCH="--continue"
else
  CONTINUE_SWITCH=""
fi

dlv --listen=:2345 --headless=true $CONTINUE_SWITCH --api-version=2 --accept-multiclient exec /usr/local/bin/axelard -- start $TOFND_HOST_SWITCH
