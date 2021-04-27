#!/bin/sh
set -e

HOME_DIR=${HOME_DIR:?home directory not set}

fileCount() {
  find "$1" -maxdepth 1 ! -iname ".*" ! -iname "$(basename "$1")" | wc -l
}

addPeers() {
  sed "s/^seeds =.*/seeds = \"$1\"/g" "$D_HOME_DIR/config/config.toml" >"$D_HOME_DIR/config/config.toml.tmp" &&
  mv "$D_HOME_DIR/config/config.toml.tmp" "$D_HOME_DIR/config/config.toml"
}

#prepareCli() {
#  #### Client
#  echo "Setting up config for CLI"
#  axelarcli config keyring-backend "$KEYRING_BACKEND"
#  axelarcli config chain-id "$CHAIN_ID"
#  axelarcli config output json
#  axelarcli config indent true
#  axelarcli config trust-node true
#}

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
    axelard init "$(hostname)" --chain-id "$CHAIN_ID"
  fi
}

cont(){
  if [ "$1" = true ]; then
    "--continue"
  else
    ""
  fi
}

startValProc() {
  sleep 10s

  if [ "$VALD_CONTINUE" != true ]; then
    unset VALD_CONTINUE
  fi

  dlv --listen=:2346 --headless=true ${VALD_CONTINUE:+--continue} --api-version=2 --accept-multiclient exec \
    /usr/local/bin/axelard -- vald-start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"} --validator-addr "$(axelarcli keys show validator -a --bech val)"
}

CLI_HOME_DIR="$HOME_DIR/.axelar"
D_HOME_DIR="$HOME_DIR/.axelar"

#if ! isCliPrepared; then
#  prepareCli
#fi

if ! isGenesisInitialized; then
  initGenesis
fi

if ! isGenesisInitialized; then
  echo "Missing genesis file"
  exit 1
fi

if [ -n "$CONFIG_PATH" ] && [ -d "$CONFIG_PATH" ]; then
  if [ -f "$CONFIG_PATH/config.toml" ]; then
    cp "$CONFIG_PATH/config.toml" "$D_HOME_DIR/config/config.toml"
  fi
  if [ -f "$CONFIG_PATH/app.toml" ]; then
    cp "$CONFIG_PATH/app.toml" "$D_HOME_DIR/config/app.toml"
  fi
fi

if [ -n "$PEERS_FILE" ]; then
  PEERS=$(cat "$PEERS_FILE")
  addPeers "$PEERS"
fi

if [ "$REST_CONTINUE" != true ]; then
  unset REST_CONTINUE
fi

startValProc &

if [ "$CORE_CONTINUE" != true ]; then
    unset CORE_CONTINUE
fi

dlv --listen=:2345 --headless=true ${CORE_CONTINUE:+--continue} --api-version=2 --accept-multiclient exec \
  /usr/local/bin/axelard -- start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"}
