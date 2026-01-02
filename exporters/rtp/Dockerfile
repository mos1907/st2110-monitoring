FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git libpcap-dev gcc musl-dev

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o st2110-rtp-exporter .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates libpcap

WORKDIR /app

COPY --from=builder /build/st2110-rtp-exporter .

EXPOSE 9100

ENTRYPOINT ["./st2110-rtp-exporter"]

