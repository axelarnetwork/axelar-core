#!/bin/sh

# generate docs
go run ./ -docs $1
# ensure docs are canonically formatted
mdformat $1/*