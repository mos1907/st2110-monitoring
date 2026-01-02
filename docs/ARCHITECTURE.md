# Architecture Overview

## System Architecture

The ST 2110 Monitoring Stack is designed to monitor professional broadcast IP networks using modern observability tools.

```
┌─────────────────────────────────────────────────────────────┐
│                    ST 2110 Environment                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌──────────┐      │
│  │ Cameras │  │Switches │  │Receivers│  │  PTP GM  │      │
│  └─────────┘  └─────────┘  └─────────┘  └──────────┘      │
└────────┬────────────┬────────────┬────────────┬────────────┘
         │            │            │            │
         ▼            ▼            ▼            ▼
┌─────────────────────────────────────────────────────────────┐
│                    Monitoring Stack                         │
│                                                              │
│  ┌────────────┐  ┌────────────┐  ┌──────────────┐         │
│  │ RTP Exporter│ │ PTP Exporter│ │ gNMI Collector│         │
│  │  :9100     │  │  :9200     │  │   :9273      │         │
│  └────────────┘  └────────────┘  └──────────────┘         │
│         │              │                    │               │
│         └──────────────┴────────────────────┘               │
│                           │                                 │
│                           ▼                                 │
│                   ┌──────────────┐                          │
│                   │  Prometheus  │                          │
│                   │   :9090      │                          │
│                   └──────────────┘                          │
│                      │        │                             │
│         ┌────────────┘        └────────────┐                │
│         │                                   │                │
│         ▼                                   ▼                │
│  ┌─────────────┐                   ┌──────────────┐        │
│  │   Grafana   │                   │ Alertmanager │        │
│  │   :3000     │                   │   :9093      │        │
│  └─────────────┘                   └──────────────┘        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Components

### Exporters

- **RTP Exporter** (`exporters/rtp/`): Captures and analyzes RTP packets, extracts metrics for packet loss, jitter, bitrate
- **PTP Exporter** (`exporters/ptp/`): Monitors PTP (IEEE 1588) clock synchronization status
- **gNMI Collector** (`exporters/gnmi/`): Collects network switch metrics via gNMI streaming telemetry

### Core Services

- **Prometheus**: Time-series database that scrapes metrics from exporters
- **Grafana**: Visualization and dashboarding platform
- **Alertmanager**: Alert routing and notification management

### Data Flow

1. Exporters collect metrics from ST 2110 network elements
2. Prometheus scrapes exporters at configured intervals (1-5 seconds)
3. Grafana queries Prometheus for visualization
4. Alertmanager receives alerts from Prometheus and routes notifications

## Network Requirements

- **Prometheus scraping**: Internal network communication
- **gNMI connections**: Access to switch management interfaces (port 6030)
- **RTP capture**: Requires network interface access with CAP_NET_RAW
- **PTP monitoring**: Network interface access for PTP packet inspection

## Resource Requirements

- **CPU**: 2-4 cores for typical deployment
- **Memory**: 4-8 GB RAM
- **Disk**: 50-100 GB for Prometheus time-series data (90-day retention)
- **Network**: Low latency network access to ST 2110 infrastructure

