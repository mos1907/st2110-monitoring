#!/bin/bash
# health-check.sh - Verify monitoring stack is healthy

echo "🔍 Checking ST 2110 Monitoring Stack Health..."
echo

# Check Prometheus
if curl -sf http://localhost:9090/-/healthy > /dev/null; then
    echo "✅ Prometheus: Healthy"
else
    echo "❌ Prometheus: DOWN"
fi

# Check Grafana
if curl -sf http://localhost:3000/api/health > /dev/null; then
    echo "✅ Grafana: Healthy"
else
    echo "❌ Grafana: DOWN"
fi

# Check Alertmanager
if curl -sf http://localhost:9093/-/healthy > /dev/null; then
    echo "✅ Alertmanager: Healthy"
else
    echo "❌ Alertmanager: DOWN"
fi

# Check exporters
echo
echo "📊 Checking Exporters..."

if curl -sf http://localhost:9100/metrics | grep -q "st2110_rtp"; then
    echo "✅ RTP Exporter: Running"
else
    echo "❌ RTP Exporter: No metrics"
fi

if curl -sf http://localhost:9200/metrics | grep -q "st2110_ptp"; then
    echo "✅ PTP Exporter: Running"
else
    echo "❌ PTP Exporter: No metrics"
fi

if curl -sf http://localhost:9273/metrics | grep -q "st2110_switch"; then
    echo "✅ gNMI Collector: Running"
else
    echo "❌ gNMI Collector: No metrics"
fi

# Check Prometheus targets
echo
echo "🎯 Checking Prometheus Targets..."
if command -v jq >/dev/null 2>&1; then
    targets=$(curl -s http://localhost:9090/api/v1/targets 2>/dev/null | jq -r '.data.activeTargets[] | select(.health != "up") | .scrapeUrl' 2>/dev/null)
    
    if [ -z "$targets" ]; then
        echo "✅ All targets UP"
    else
        echo "❌ Targets DOWN:"
        echo "$targets"
    fi
else
    echo "⚠️  jq not installed - skipping target check"
fi

# Check for firing alerts
echo
echo "🚨 Checking Alerts..."
if command -v jq >/dev/null 2>&1; then
    alerts=$(curl -s http://localhost:9090/api/v1/alerts 2>/dev/null | jq -r '.data.alerts[] | select(.state == "firing") | .labels.alertname' 2>/dev/null)
    
    if [ -z "$alerts" ]; then
        echo "✅ No firing alerts"
    else
        echo "⚠️  Firing alerts:"
        echo "$alerts"
    fi
else
    echo "⚠️  jq not installed - skipping alert check"
fi

echo
echo "✅ Health check complete!"

