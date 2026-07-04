# Multi-stage build for Alpine Docker
FROM --platform=$BUILDPLATFORM golang:1.21-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Strip 'v' prefix from TARGETVARIANT for GOARM (v6 -> 6, v7 -> 7)
RUN GOARM_VALUE=$(echo ${TARGETVARIANT:-} | sed 's/^v//'); \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} GOARM=${GOARM_VALUE:-} go build -trimpath -ldflags="-s -w" -o dra .

# Final stage: minimal Alpine image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates bzip2 xz

COPY --from=builder /build/dra /usr/local/bin/dra

ENTRYPOINT ["dra"]
