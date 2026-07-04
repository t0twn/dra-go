# Multi-stage build for Alpine Docker
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o dra .

# Final stage: minimal Alpine image
FROM alpine:3.19

RUN apk add --no-cache ca-certificates bzip2 xz

COPY --from=builder /build/dra /usr/local/bin/dra

ENTRYPOINT ["dra"]
