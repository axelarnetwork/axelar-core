#!/bin/sh
set -e

trap stop_gracefully TERM INT

stop_gracefully(){
  echo "stopping all processes"
  killall "axelard"
  sleep 10
  echo "all processes stopped"
}

HOME_DIR=${HOME_DIR:?home directory not set}

fileCount() {
  find "$1" -maxdepth 1 ! -iname ".*" ! -iname "$(basename "$1")" | wc -l
}

addPeers() {
  sed "s/^seeds =.*/seeds = \"$1\"/g" "$D_HOME_DIR/config/config.toml" >"$D_HOME_DIR/config/config.toml.tmp" &&
  mv "$D_HOME_DIR/config/config.toml.tmp" "$D_HOME_DIR/config/config.toml"
}

isInitialized() {
  if [ "$1" == "startValdProc" ]; then
    if [ -z "$BROADCASTER_ACCOUNT" ] || [ -z "$VALIDATOR_ADDR" ]; then
      return 1
    fi

    ACCOUNTS=$(axelard keys list -n)
    for ACCOUNT in $ACCOUNTS; do
      if [ "$ACCOUNT" == "$BROADCASTER_ACCOUNT" ]; then
        HAS_BROADCASTER=true
      fi
    done

    if [ -z "$HAS_BROADCASTER" ]; then
      return 1
    fi

    return 0
  fi

  if [ -f "$D_HOME_DIR/config/genesis.json" ]; then
    return 0
  fi

  return 1
}

initialize() {
  if [ -n "$INIT_SCRIPT" ] && [ -f "$INIT_SCRIPT" ]; then
    echo "Running script at $INIT_SCRIPT to initialize container"
    source "$INIT_SCRIPT" "$(hostname)" "$AXELARD_CHAIN_ID"
  else
    axelard init "$(hostname)" --chain-id "$AXELARD_CHAIN_ID"
  fi
}

startValdProc() {
  DURATION=${SLEEP_TIME:-"10s"}
  sleep $DURATION
  if [ -n "$RECOVERY_FILE" ] & [ -f "$RECOVERY_FILE" ]; then
    RECOVERY="--tofnd-recovery=$RECOVERY_FILE"
  fi

  axelard vald-start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"} ${VALIDATOR_HOST:+--node "$VALIDATOR_HOST"} \
    --validator-addr "${VALIDATOR_ADDR:-$(axelard keys show validator -a --bech val)}" "$RECOVERY"
}

startNodeProc() {
  axelard start
}

D_HOME_DIR="$HOME_DIR/.axelar"

if ! $(isInitialized $1); then
  initialize
fi

if ! $(isInitialized $1); then
  echo "Container not properly initialized"
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

if [ -z "$1" ]; then

  startValdProc &

  startNodeProc &

  wait

else

  $@ &

  wait

fi
