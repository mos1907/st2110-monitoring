# PTP Exporter Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev linux-headers

# Copy go.mod if exists, otherwise create minimal one
COPY go.mod* ./
RUN go mod init ptp-exporter 2>/dev/null || true
RUN go mod tidy 2>/dev/null || true

# Copy source code
COPY main.go .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ptp-exporter .

# Final stage
FROM alpine:latest

# Install ptp4l and pmc (PTP tools)
RUN apk --no-cache add ca-certificates linux-ptp

WORKDIR /app

COPY --from=builder /build/ptp-exporter .

EXPOSE 9200

ENTRYPOINT ["./ptp-exporter"]

