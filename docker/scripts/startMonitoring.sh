#!/bin/bash

/prometheus/prometheus --config.file=/prometheus/prometheus.yml --storage.tsdb.path="/data/prometheus" &
/run.sh
