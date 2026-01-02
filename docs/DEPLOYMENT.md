# Deployment Guide

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ (for building exporters from source)
- Network access to ST 2110 infrastructure
- Root/sudo access (for packet capture capabilities)

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/st2110-monitoring.git
cd st2110-monitoring
```

### 2. Configure Environment

```bash
cp .env.example .env
# Edit .env with your credentials
```

### 3. Configure Streams

```bash
cp config/streams.yaml.example config/streams.yaml
# Edit config/streams.yaml with your multicast addresses
```

### 4. Configure Switches (Optional)

```bash
cp config/switches.yaml.example config/switches.yaml
# Edit config/switches.yaml with your switch configurations
```

### 5. Deploy

```bash
docker-compose up -d
```

### 6. Verify

Access the services:
- Grafana: http://localhost:3000 (admin/admin)
- Prometheus: http://localhost:9090
- Alertmanager: http://localhost:9093

## Production Deployment

### Docker Compose Deployment

See `docker-compose.yml` for the complete stack configuration.

### Kubernetes Deployment

Kubernetes manifests are available in `kubernetes/` directory:

```bash
kubectl apply -f kubernetes/
```

### High Availability

For production deployments, consider:

- Prometheus federation for scale
- Grafana with external database (PostgreSQL)
- Alertmanager clustering
- Multiple exporter instances for redundancy

## Configuration

### Stream Configuration

Edit `config/streams.yaml` to define your ST 2110 streams:

```yaml
streams:
  - name: "Camera 1 - Video"
    stream_id: "cam1_vid"
    multicast: "239.1.1.10:20000"
    interface: "eth0"
    type: "video"
    format: "1080p60"
    expected_bitrate: 2200000000
```

### Switch Configuration

Edit `config/switches.yaml` for gNMI switch monitoring:

```yaml
switches:
  - name: "core-switch-1"
    target: "192.168.1.10:6030"
    username: "prometheus"
    password: "${GNMI_PASSWORD}"
    vendor: "arista"
```

## Security

- Change default passwords in `.env`
- Use TLS/mTLS for production deployments
- Implement network segmentation
- Follow least-privilege access principles
- Enable audit logging

## Troubleshooting

See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions.

