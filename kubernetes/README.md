# Kubernetes Deployment

Kubernetes deployment files for ST 2110 Monitoring Stack.

## Prerequisites

- Kubernetes cluster (1.21+)
- kubectl configured
- ConfigMaps and Secrets created (see below)

## Quick Start

```bash
# 1. Create namespace
kubectl apply -f namespace.yaml

# 2. Create ConfigMaps
kubectl create configmap prometheus-config \
  --from-file=../../prometheus/prometheus.yml \
  --from-file=../../prometheus/alerts \
  -n st2110-monitoring

kubectl create configmap alertmanager-config \
  --from-file=../../alertmanager/alertmanager.yml \
  -n st2110-monitoring

kubectl create configmap grafana-dashboards \
  --from-file=../../grafana/dashboards \
  -n st2110-monitoring

kubectl create configmap grafana-provisioning \
  --from-file=../../grafana/provisioning \
  -n st2110-monitoring

kubectl create configmap rtp-exporter-config \
  --from-file=../../config/streams.yaml \
  -n st2110-monitoring

kubectl create configmap gnmi-collector-config \
  --from-file=../../config/switches.yaml \
  -n st2110-monitoring

# 3. Create Secrets
kubectl create secret generic grafana-secrets \
  --from-literal=admin-password=your-secure-password \
  -n st2110-monitoring

kubectl create secret generic gnmi-secrets \
  --from-literal=password=your-gnmi-password \
  -n st2110-monitoring

# 4. Deploy Prometheus
kubectl apply -f prometheus/

# 5. Deploy Grafana
kubectl apply -f grafana/

# 6. Deploy Alertmanager
kubectl apply -f alertmanager/

# 7. Deploy Exporters
kubectl apply -f exporters/

# 8. Verify
kubectl get pods -n st2110-monitoring
kubectl get services -n st2110-monitoring
```

## Port Forwarding

```bash
# Access Grafana
kubectl port-forward -n st2110-monitoring svc/grafana 3000:3000

# Access Prometheus
kubectl port-forward -n st2110-monitoring svc/prometheus 9090:9090

# Access Alertmanager
kubectl port-forward -n st2110-monitoring svc/alertmanager 9093:9093
```

## Notes

- RTP Exporter requires `hostNetwork: true` and privileged mode for packet capture
- Adjust resource limits based on your cluster capacity
- For production, set replicas to 2+ for high availability
- Consider using PersistentVolumes for Prometheus data storage
- Update image tags to use your container registry

