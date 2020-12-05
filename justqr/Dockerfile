FROM --platform=$BUILDPLATFORM rust:1.48 AS builder

WORKDIR /rust

COPY Cargo.* /rust
ADD ./scripts /rust/scripts

ARG TARGETARCH
ARG TARGETOS

RUN mkdir -p /rust/src \
    && echo "fn main() {}" > /rust/src/main.rs \
    && ./scripts/install-toolchain.sh \
    && ./scripts/build-dependencies.sh

COPY . .

RUN ./scripts/install-application.sh

FROM rust:1.48-slim AS bin
LABEL org.opencontainers.image.source https://github.com/parente/homelab
COPY --from=builder /usr/local/cargo/bin/justqr /rust/
