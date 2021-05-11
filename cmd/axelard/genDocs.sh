#!/bin/sh

# remove old docs
if find "$1" -name "*.md" 2> /dev/null | grep -q .; then
  rm "$1"/*.md
fi

# generate docs
go run ./ -docs "$1"
# ensure docs are canonically formatted
mdformat "$1"/*