FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0

COPY go.* .
RUN go mod download
COPY *.go .

ARG TARGETARCH
ARG TARGETOS

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/cfcidrwatch

FROM scratch AS bin
LABEL org.opencontainers.image.source https://github.com/parente/homelab
COPY --from=builder /go/bin/cfcidrwatch /go/bin/cfcidrwatch
ENTRYPOINT ["/go/bin/cfcidrwatch"]
