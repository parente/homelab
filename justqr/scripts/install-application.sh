#!/bin/bash
set -e

if [[ "$TARGETARCH" == "arm/v7" ]]; then
    RUST_TARGET=armv7-unknown-${TARGETOS}-gnueabihf
elif [[ "$TARGETARCH" == "amd64" ]]; then
    RUST_TARGET=x86_64-unknown-${TARGETOS}-gnu
else 
    echo "Unknown arch: $TARGETARCH"
    exit 1
fi

cargo install --target $RUST_TARGET --path .