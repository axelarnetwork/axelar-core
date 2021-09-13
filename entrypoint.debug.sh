#!/bin/bash
set -e

trap stop_gracefully SIGTERM SIGINT

stop_gracefully(){
  echo "stopping all processes"
  pkill "axelard"
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

cont(){
  if [ "$1" = true ]; then
    "--continue"
  else
    ""
  fi
}

startValdProc() {
  DURATION=${SLEEP_TIME:-"10s"}
  sleep $DURATION
  
  if [ "$VALD_CONTINUE" != true ]; then
    unset VALD_CONTINUE
  fi

  dlv --listen=:2346 --headless=true ${VALD_CONTINUE:+--continue} --api-version=2 --accept-multiclient exec \
    /usr/local/bin/axelard -- vald-start ${TOFND_HOST:+--tofnd-host "$TOFND_HOST"} ${VALIDATOR_HOST:+--node "$VALIDATOR_HOST"} --validator-addr "${VALIDATOR_ADDR:-$(axelard keys show validator -a --bech val)}"
}

startNodeProc() {
  if [ "$CORE_CONTINUE" != true ]; then
    unset CORE_CONTINUE
  fi

  dlv --listen=:2345 --headless=true ${CORE_CONTINUE:+--continue} --api-version=2 --accept-multiclient exec \
    /usr/local/bin/axelard -- start
}

D_HOME_DIR="$HOME_DIR/.axelar"

  if [ -n "$PRESTART_SCRIPT" ] && [ -f "$PRESTART_SCRIPT" ]; then
    echo "Running pre-start script at $PRESTART_SCRIPT"
    source "$PRESTART_SCRIPT"
  fi

if [ -n "$CONFIG_PATH" ] && [ -d "$CONFIG_PATH" ]; then
  if [ -f "$CONFIG_PATH/config.toml" ]; then
    cp "$CONFIG_PATH/config.toml" "$D_HOME_DIR/config/config.toml"
  fi
  if [ -f "$CONFIG_PATH/app.toml" ]; then
    cp "$CONFIG_PATH/app.toml" "$D_HOME_DIR/config/app.toml"
  fi
  if [ -f "$CONFIG_PATH/vald.toml" ]; then
    cp "$CONFIG_PATH/vald.toml" "$D_HOME_DIR/config/vald.toml"
  fi
fi

if [ -n "$PEERS_FILE" ]; then
  PEERS=$(cat "$PEERS_FILE")
  addPeers "$PEERS"
fi

if [ "$REST_CONTINUE" != true ]; then
  unset REST_CONTINUE
fi

if [ -z "$1" ]; then

  startValdProc &

  startNodeProc &

  wait

else

  $@ &

  wait

fi
