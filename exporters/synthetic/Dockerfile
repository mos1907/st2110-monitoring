# Synthetic Test Stream Generator Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go.mod if exists, otherwise create minimal one
COPY go.mod* ./
RUN go mod init synthetic-generator 2>/dev/null || true
RUN go get github.com/google/gopacket/layers
RUN go get github.com/google/gopacket
RUN go mod tidy 2>/dev/null || true

# Copy source code
COPY main.go .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o synthetic-generator .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /build/synthetic-generator .

ENTRYPOINT ["./synthetic-generator"]

