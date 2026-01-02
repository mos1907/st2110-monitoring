#!/bin/bash
# backup.sh - Backup Prometheus and Grafana data

set -e

BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

echo "📦 Creating backup at $(date)..."
echo

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup Prometheus data
if docker volume ls | grep -q "st2110-monitoring_prometheus_data"; then
    echo "Backing up Prometheus data..."
    docker run --rm \
        -v st2110-monitoring_prometheus_data:/data \
        -v "$(pwd)/$BACKUP_DIR:/backup" \
        alpine tar czf "/backup/prometheus-backup-$TIMESTAMP.tar.gz" -C /data .
    echo "✅ Prometheus backup created: prometheus-backup-$TIMESTAMP.tar.gz"
else
    echo "⚠️  Prometheus volume not found, skipping..."
fi

# Backup Grafana data
if docker volume ls | grep -q "st2110-monitoring_grafana_data"; then
    echo "Backing up Grafana data..."
    docker run --rm \
        -v st2110-monitoring_grafana_data:/data \
        -v "$(pwd)/$BACKUP_DIR:/backup" \
        alpine tar czf "/backup/grafana-backup-$TIMESTAMP.tar.gz" -C /data .
    echo "✅ Grafana backup created: grafana-backup-$TIMESTAMP.tar.gz"
else
    echo "⚠️  Grafana volume not found, skipping..."
fi

# Backup configurations
echo "Backing up configurations..."
tar czf "$BACKUP_DIR/config-backup-$TIMESTAMP.tar.gz" \
    prometheus/ \
    grafana/ \
    alertmanager/ \
    config/ \
    2>/dev/null || true
echo "✅ Configuration backup created: config-backup-$TIMESTAMP.tar.gz"

echo
echo "✅ Backup complete! Files saved to $BACKUP_DIR/"

