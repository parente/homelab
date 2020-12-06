# syntax=docker/dockerfile:experimental
FROM --platform=$BUILDPLATFORM rust:1.48 AS builder

ARG TARGETARCH
ARG TARGETVARIANT
ARG TARGETOS

WORKDIR /rust

COPY ./scripts/install-toolchain.sh ./scripts/
RUN ./scripts/install-toolchain.sh

COPY . .

RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/rust/target \
    ./scripts/install-application.sh

FROM gcr.io/distroless/cc-debian10:nonroot-${TARGETARCH} AS bin
LABEL org.opencontainers.image.source https://github.com/parente/homelab
COPY --from=builder /usr/local/cargo/bin/justqr /rust/justqr
COPY --from=builder /rust/static  /rust/static
WORKDIR /rust
ENTRYPOINT ["/rust/justqr"]