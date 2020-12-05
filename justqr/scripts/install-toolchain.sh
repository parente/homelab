#!/bin/bash
set -e

apt update

if [[ "$TARGETARCH" == "arm/v7" ]]; then
    apt install gcc-arm-linux-gnueabihf
    RUST_TARGET=armv7-unknown-${TARGETOS}-gnueabihf
elif [[ "$TARGETARCH" == "amd64" ]]; then
    RUST_TARGET=x86_64-unknown-${TARGETOS}-gnu
else 
    echo "Unknown arch: $TARGETARCH"
    exit 1
fi

rustup target add $RUST_TARGET