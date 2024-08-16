#!/bin/sh

# remove old docs
if find "$2" -name "*.md" 2> /dev/null | grep -q .; then
  rm "$2"/*.md
fi

# read build flags from Makefile. The Makefile prints out some additional information that needs to be stripped out first
# Procedure: print the build flags, only grab the line starting with ldflags=,
# strip that 'ldflags=' prefix and assign to "ldflags" variable
ldflags=$(make -f "$1" print-ldflags | grep ^ldflags= | sed s/^ldflags=//)

# generate docs
go run -ldflags "$ldflags" ./ -docs "$2"
# ensure docs are canonically formatted
mdformat "$2"/*
