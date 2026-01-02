# API Reference

## Prometheus Metrics API

### RTP Stream Metrics

All metrics are exposed at `/metrics` endpoint of each exporter.

#### `st2110_rtp_packets_received_total`
- **Type**: Counter
- **Description**: Total number of RTP packets received
- **Labels**: `stream_id`, `stream_name`, `multicast`, `type`

#### `st2110_rtp_packets_lost_total`
- **Type**: Counter
- **Description**: Total number of RTP packets lost
- **Labels**: `stream_id`, `stream_name`, `multicast`, `type`

#### `st2110_rtp_jitter_microseconds`
- **Type**: Gauge
- **Description**: Current interarrival jitter in microseconds
- **Labels**: `stream_id`, `stream_name`, `multicast`, `type`

#### `st2110_rtp_bitrate_bps`
- **Type**: Gauge
- **Description**: Current RTP stream bitrate in bits per second
- **Labels**: `stream_id`, `stream_name`, `multicast`, `type`

#### `st2110_rtp_packet_loss_rate`
- **Type**: Gauge
- **Description**: Packet loss rate as percentage (0-100)
- **Labels**: `stream_id`, `stream_name`, `multicast`, `type`

### PTP Metrics

#### `st2110_ptp_offset_nanoseconds`
- **Type**: Gauge
- **Description**: PTP offset from master in nanoseconds
- **Labels**: `device`, `interface`, `master`

#### `st2110_ptp_mean_path_delay_nanoseconds`
- **Type**: Gauge
- **Description**: Mean path delay in nanoseconds
- **Labels**: `device`, `interface`

#### `st2110_ptp_clock_state`
- **Type**: Gauge
- **Description**: PTP clock state (0=FREERUN, 1=LOCKED, 2=HOLDOVER)
- **Labels**: `device`, `interface`

### Network Switch Metrics

#### `st2110_switch_interface_rx_bytes`
- **Type**: Counter
- **Description**: Received bytes on switch interface
- **Labels**: `switch`, `interface`, `vlan`

#### `st2110_switch_qos_buffer_utilization`
- **Type**: Gauge
- **Description**: QoS buffer utilization percentage
- **Labels**: `switch`, `interface`, `queue`

#### `st2110_switch_qos_dropped_packets`
- **Type**: Counter
- **Description**: Dropped packets by QoS policy
- **Labels**: `switch`, `interface`, `queue`

## Exporter HTTP Endpoints

### RTP Exporter (:9100)
- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check endpoint

### PTP Exporter (:9200)
- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check endpoint

### gNMI Collector (:9273)
- `GET /metrics` - Prometheus metrics
- `GET /health` - Health check endpoint

## Query Examples

### Packet Loss Rate
```promql
rate(st2110_rtp_packets_lost_total[5m]) / 
rate(st2110_rtp_packets_received_total[5m]) * 100
```

### Average Jitter
```promql
avg_over_time(st2110_rtp_jitter_microseconds[5m])
```

### Total Bandwidth
```promql
sum(st2110_rtp_bitrate_bps)
```

