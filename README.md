# ST 2110 Monitoring Stack

Production-ready monitoring solution for SMPTE ST 2110 professional media networks using Prometheus, Grafana, and custom exporters.

## 🎯 Features

- **Real-time RTP Stream Monitoring**: Packet loss, jitter, bitrate tracking
- **PTP Timing Analysis**: Sub-microsecond clock offset monitoring
- **Network Switch Metrics**: gNMI-based telemetry streaming
- **Video Quality Metrics**: TR-03 compliance, buffer analysis
- **SMPTE 2022-7 Monitoring**: Seamless protection switching
- **NMOS Integration**: Auto-discovery and control plane health
- **Production-Ready**: Docker Compose, Kubernetes, CI/CD pipelines
- **Comprehensive Alerting**: Pre-configured rules for broadcast scenarios

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for building exporters)
- Access to ST 2110 network interfaces
- Root/sudo access (for packet capture)

### One-Command Deployment

```bash
# Clone repository
git clone https://github.com/yourusername/st2110-monitoring
cd st2110-monitoring

# Configure your streams
cp config/streams.yaml.example config/streams.yaml
# Edit config/streams.yaml with your multicast addresses

# Start monitoring stack
docker-compose up -d

# Access dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
```

## 📁 Project Structure

```
st2110-monitoring/
├── docker-compose.yml           # Complete monitoring stack
├── .env.example                 # Environment variables template
├── Makefile                     # Build and management commands
│
├── exporters/
│   ├── rtp/                     # RTP stream exporter
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── ptp/                     # PTP timing exporter
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── gnmi/                    # gNMI network collector
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── synthetic/               # Test stream generator
│   │   └── main.go
│   └── vendor/                  # Vendor-specific exporters
│       ├── arista/              # Arista EOS gNMI collector
│       ├── cisco/               # Cisco Nexus gNMI collector
│       ├── grassvalley/         # Grass Valley K-Frame REST API
│       ├── evertz/              # Evertz EQX/VIP SNMP/API
│       └── lawo/                # Lawo VSM REST API
│
├── prometheus/
│   ├── prometheus.yml           # Main configuration
│   └── alerts/
│       ├── st2110.yml          # ST 2110 alert rules
│       ├── tr03.yml            # TR-03 compliance alerts
│       └── multicast.yml       # Multicast alerts
│
├── grafana/
│   ├── provisioning/
│   │   ├── datasources/
│   │   └── dashboards/
│   └── dashboards/
│       └── st2110-dashboard.json  # Main dashboard
│
├── alertmanager/
│   └── alertmanager.yml        # Alert routing
│
├── config/
│   ├── streams.yaml.example    # Stream definitions
│   └── switches.yaml.example   # Network switches
│
├── kubernetes/
│   ├── namespace.yaml
│   ├── prometheus/
│   ├── grafana/
│   └── exporters/
│
├── scripts/
│   ├── health-check.sh
│   ├── backup.sh
│   └── deploy.sh
│
└── docs/
    ├── ARCHITECTURE.md
    ├── DEPLOYMENT.md
    ├── TROUBLESHOOTING.md
    └── API.md
```

## 🔧 Configuration

### 1. Define Your Streams

Edit `config/streams.yaml`:

```yaml
streams:
  - name: "Camera 1 - Video"
    stream_id: "cam1_vid"
    multicast: "239.1.1.10:20000"
    interface: "eth0"
    type: "video"
    format: "1080p60"
    expected_bitrate: 2200000000

  - name: "Camera 1 - Audio"
    stream_id: "cam1_aud"
    multicast: "239.1.1.11:20000"
    interface: "eth0"
    type: "audio"
    channels: 8
    sample_rate: 48000
```

### 2. Configure Network Switches

Edit `config/switches.yaml`:

```yaml
switches:
  - name: "core-switch-1"
    target: "192.168.1.10:6030"
    username: "prometheus"
    password: "${GNMI_PASSWORD}"
    vendor: "arista"  # Options: arista, cisco, juniper
```

**Supported Vendors:**
- **Arista EOS**: Full gNMI support with hardware queue monitoring
- **Cisco Nexus**: gNMI with DME (Data Management Engine) paths
- **Juniper**: gNMI with OpenConfig models

### 3. Vendor-Specific Integrations

#### Network Switches (gNMI)

**Arista EOS Configuration:**
```bash
# Enable gNMI on Arista switch
switch(config)# management api gnmi
switch(config-mgmt-api-gnmi)# transport grpc default
switch(config-mgmt-api-gnmi-transport-default)# ssl profile default
switch(config-mgmt-api-gnmi)# provider eos-native
```

**Cisco Nexus Configuration:**
```bash
# Enable gRPC/gNMI on Cisco Nexus
switch(config)# feature grpc
switch(config)# grpc port 6030
switch(config)# feature nxapi
```

#### Broadcast Equipment (REST APIs)

**Supported Integrations:**
- **Grass Valley K-Frame**: REST API for card status, video inputs, crosspoint monitoring
- **Evertz EQX/VIP**: SNMP and HTTP XML API for module status, IP flows, PTP status
- **Lawo VSM**: REST API for device status, pathway monitoring, alarm aggregation

See the main article for detailed integration examples and code samples.

### 4. Set Environment Variables

```bash
cp .env.example .env
# Edit .env with your credentials
```

## 📊 Dashboards

### Main Production Dashboard

- **Stream Overview**: Real-time status of all streams
- **RTP Metrics**: Packet loss, jitter, bitrate
- **PTP Timing**: Clock offset, path delay
- **Network Health**: Switch ports, QoS, buffer utilization
- **Video Quality**: TR-03 compliance, buffer levels
- **Alerts**: Active incidents and history

### Capacity Planning Dashboard

- Bandwidth growth trends
- Predicted capacity exhaustion
- Stream count projections

## 🔔 Alerting

Pre-configured alert rules for:

- High packet loss (>0.01%)
- Excessive jitter (>1000μs)
- PTP timing issues (>1μs offset)
- Network congestion
- Buffer underruns/overruns
- IGMP membership failures
- SMPTE 2022-7 protection switching

Notifications via:
- Slack
- PagerDuty
- Email
- Webhook

## 🏗️ Building from Source

```bash
# Build all exporters
make build

# Build specific exporter
cd exporters/rtp && go build -o st2110-rtp-exporter

# Run tests
make test

# Generate mocks
make generate
```

## 🐳 Docker Deployment

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Backup Prometheus data
make backup
```

## ☸️ Kubernetes Deployment

```bash
# Deploy to Kubernetes
kubectl apply -f kubernetes/

# Port forward Grafana
kubectl port-forward -n monitoring svc/grafana 3000:3000

# Scale exporters
kubectl scale deployment/st2110-rtp-exporter --replicas=3
```

## 📈 Metrics Reference

### RTP Stream Metrics

```
st2110_rtp_packets_received_total{stream_id, type}
st2110_rtp_packets_lost_total{stream_id, type}
st2110_rtp_jitter_microseconds{stream_id, type}
st2110_rtp_bitrate_bps{stream_id, type}
st2110_rtp_packet_loss_rate{stream_id, type}
```

### PTP Metrics

```
st2110_ptp_offset_nanoseconds{device, interface}
st2110_ptp_mean_path_delay_nanoseconds{device, interface}
st2110_ptp_clock_state{device, interface}
```

### Network Metrics

```
st2110_switch_interface_rx_bytes{switch, interface}
st2110_switch_qos_buffer_utilization{switch, interface, queue}
st2110_switch_qos_dropped_packets{switch, interface, queue}
```

### Vendor-Specific Metrics

**Arista EOS:**
```
arista_hw_queue_drops_total{switch, interface, queue}
arista_ptp_lock_status{switch, domain}
arista_igmp_snooping_groups{switch, vlan}
arista_tcam_utilization_percent{switch, table}
```

**Cisco Nexus:**
```
cisco_nexus_tcam_utilization_percent{switch, table_type}
cisco_nexus_qos_policy_drops_total{switch, policy, class}
cisco_nexus_buffer_drops_total{switch, interface}
```

**Grass Valley K-Frame:**
```
grassvalley_kframe_card_status{chassis, slot, card_type}
grassvalley_kframe_video_input_status{chassis, slot, input}
grassvalley_kframe_crosspoint_count{chassis, router_level}
```

**Evertz EQX/VIP:**
```
evertz_eqx_module_status{chassis, slot, module_type}
evertz_eqx_ip_flow_status{chassis, flow_id, direction}
evertz_eqx_ptp_lock_status{chassis, module}
```

**Lawo VSM:**
```
lawo_vsm_connection_status{device_name, device_type}
lawo_vsm_pathway_status{pathway_name, source, destination}
lawo_vsm_active_alarms{severity}
```

## 🔒 Security

- TLS/mTLS for Prometheus scraping
- RBAC in Grafana
- Network segmentation (VLANs)
- Secrets management with HashiCorp Vault
- Audit logging enabled

## 🛠️ Troubleshooting

### No Metrics Showing

```bash
# Check exporter is running
curl http://localhost:9100/metrics

# Check Prometheus targets
open http://localhost:9090/targets

# Check logs
docker-compose logs st2110-rtp-exporter
```

### High Packet Loss Detected

1. Check network path: `traceroute <multicast_ip>`
2. Verify QoS configuration on switches
3. Check for IGMP membership issues
4. Analyze switch buffer utilization

See [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for more.

## 📚 Documentation

- [Architecture Overview](docs/ARCHITECTURE.md)
- [Deployment Guide](docs/DEPLOYMENT.md)
- [API Reference](docs/API.md)
- [Contributing Guide](CONTRIBUTING.md)

## 🤝 Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- SMPTE ST 2110 standards committee
- Prometheus & Grafana communities
- OpenConfig project (gNMI)
- AMWA NMOS community

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/st2110-monitoring/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/st2110-monitoring/discussions)
- **Email**: murat@muratdemirci.com.tr

## 🎓 Related Resources

- [SMPTE ST 2110 Standards](https://www.smpte.org/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [gNMI Specification](https://github.com/openconfig/gnmi)
- [Arista EOS gNMI Guide](https://www.arista.com/en/um-eos/eos-gnmi)
- [Cisco Nexus gRPC/gNMI Configuration](https://www.cisco.com/c/en/us/td/docs/switches/datacenter/nexus9000/sw/7-x/programmability/guide/b_Cisco_Nexus_9000_Series_NX-OS_Programmability_Guide_7x/b_Cisco_Nexus_9000_Series_NX-OS_Programmability_Guide_7x_chapter_011000.html)

---

**⚠️ Production Note**: This is a monitoring solution. It observes your ST 2110 network but does not control or modify streams. Always test in a non-production environment first.

**🚀 Quick Stats**: Monitor 100+ streams, <5s incident detection, <1μs timing accuracy

