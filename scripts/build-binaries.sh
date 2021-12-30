#!/usr/bin/env bash

platforms="$(go tool dist list | grep 'darwin\|linux' | grep 'amd64\|arm')"
version="${1}"
build_tags="${2}"
ldflags="${3}"
echo "Building version $version"
for platform in $platforms
do
    arch="$(echo "$platform" | awk -F/ '{print $NF}')"
    os="$(echo "$platform" | awk -F/ '{print $(NF-1)}')"
    echo "Building binary for OS $os Architecture $arch"
    echo "ld_flags: ${ldflags}"
    echo "build_tags: ${build_tags}"
    echo "version: ${version}"
    GOOS=$os GOARCH=$arch go build -o ./bin/axelard-"$os"-"$arch"-"$version" -mod=readonly -tags "$build_tags" -ldflags "$ldflags" ./cmd/axelard
done

cd bin || exit 1
sha256sum * > SHA256SUMS
