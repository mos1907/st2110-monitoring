# Troubleshooting Guide

## Common Issues

### No Metrics Showing in Prometheus

**Symptoms**: Prometheus targets show as DOWN or no metrics appear in Grafana

**Solutions**:
1. Check exporter is running: `docker-compose ps`
2. Verify exporter endpoints: `curl http://localhost:9100/metrics`
3. Check Prometheus targets: http://localhost:9090/targets
4. Review exporter logs: `docker-compose logs st2110-rtp-exporter`

### High Packet Loss Detected

**Symptoms**: Alerts firing for packet loss > 0.01%

**Diagnostic Steps**:
1. Check network path: `traceroute <multicast_ip>`
2. Verify QoS configuration on switches
3. Check for IGMP membership issues
4. Analyze switch buffer utilization
5. Verify stream configuration matches network setup

**Common Causes**:
- Network congestion
- Misconfigured QoS policies
- IGMP snooping issues
- Switch buffer exhaustion

### PTP Offset Issues

**Symptoms**: PTP offset > 10μs, sync issues

**Diagnostic Steps**:
1. Verify PTP grandmaster is accessible
2. Check PTP exporter logs
3. Verify network path to grandmaster
4. Check switch PTP configuration

### Grafana Dashboard Not Loading

**Symptoms**: Dashboard shows "No data" or errors

**Solutions**:
1. Verify Prometheus datasource is configured
2. Check datasource connection: http://localhost:9090
3. Verify metrics exist: Query Prometheus directly
4. Check dashboard JSON is valid
5. Review Grafana logs: `docker-compose logs grafana`

### Exporters Not Starting

**Symptoms**: Containers exit immediately

**Solutions**:
1. Check logs: `docker-compose logs <service-name>`
2. Verify configuration files exist and are valid YAML
3. Check network permissions (for RTP exporter)
4. Verify required capabilities are set in docker-compose.yml

### Network Permission Errors

**Symptoms**: RTP exporter cannot capture packets

**Solutions**:
1. Verify `network_mode: host` is set in docker-compose.yml
2. Check `CAP_NET_RAW` and `CAP_NET_ADMIN` are enabled
3. On Linux, ensure user has packet capture permissions
4. Verify network interface exists and is accessible

## Log Locations

- Prometheus logs: `docker-compose logs prometheus`
- Grafana logs: `docker-compose logs grafana`
- Exporter logs: `docker-compose logs st2110-rtp-exporter`
- All logs: `docker-compose logs -f`

## Getting Help

- Check GitHub Issues: https://github.com/yourusername/st2110-monitoring/issues
- Review documentation in `docs/` directory
- Consult the main article for detailed examples

