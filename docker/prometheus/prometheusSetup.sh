#!/bin/bash

if ! grep -qi prometheus "/etc/passwd"; then
  echo "Creating user \"prometheus\""
  sudo useradd -rs /bin/false prometheus
fi
sed "s/UID=.*/UID=$(id -u prometheus)/g" .env >.env.tmp && mv .env.tmp .env
sed "s/GID=.*/GID=$(id -g prometheus)/g" .env >.env.tmp && mv .env.tmp .env
echo "Modifying permissions for prometheus directories"
mkdir "$(dirname "$0")"/data/prometheus -p
sudo chown prometheus:prometheus "$(dirname "$0")"/data/prometheus
