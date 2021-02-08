#!/bin/sh

# remove old docs
rm "$1"/*.md
# generate docs
go run ./ -docs "$1"
# ensure docs are canonically formatted
mdformat "$1"/*