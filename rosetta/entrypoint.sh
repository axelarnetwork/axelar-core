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

if [ -z "$1" ]; then

  startNodeProc &

  wait

else

  $@ &

  wait

fi
