FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0

COPY ./kube-state-metrics/go.* .
RUN go mod download
COPY ./kube-state-metrics .

ARG TARGETARCH
ARG TARGETOS

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/kube-state-metrics

FROM scratch AS bin
LABEL org.opencontainers.image.source https://github.com/parente/homelab
COPY --from=builder /go/bin/kube-state-metrics /go/bin/kube-state-metrics
ENTRYPOINT ["/go/bin/kube-state-metrics", "--port=8080", "--telemetry-port=8081"]
EXPOSE 8080 8081
