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

isGenesisInitialized() {
  if [ -f "$D_HOME_DIR/config/genesis.json" ]; then
    return 0
  fi

  return 1
}

isValidator() {
  if axelarcli query staking validator "$(axelarcli keys show validator --bech=val -a)"; then
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

startValProc() {
  sleep 10s
  axelard vald-start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"} --validator-addr "$(axelarcli keys show validator -a --bech val)"
}

D_HOME_DIR="$HOME_DIR/.axelar"
}
  done
    fi
      break
      strace -ewrite -p $(pidof vald)
      echo Attaching to vald output
    if [ $PROCESS_1_STATUS = 0 ]; then

    VALD_STATUS=$?
    ps aux |grep vald | grep -q -v grep
    sleep 5
  for i in {1..100}; do
attachValStdout() {

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

if [ "$START_REST" = true ]; then
  # REST endpoint must be bound to 0.0.0.0 for availability on docker host
  axelarcli rest-server \
    --chain-id=axelarcli \
    --laddr=tcp://0.0.0.0:1317 \
    --node tcp://0.0.0.0:26657 \
    --unsafe-cors &
fi

if axelarcli query staking validator "$(axelarcli keys show validator --bech=val -a)"; then
  startValProc &
else
  # assume vald started by makeValidator.sh and attempt to attach output
  attachValStdout &
fi

exec axelard start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"}
