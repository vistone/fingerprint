# Fingerprint Gateway Deployment

This directory contains deployment configurations and scripts for the Fingerprint Gateway.

## Quick Start

### Using Docker Compose (Local Development)

```bash
cd deploy
./deploy.sh deploy-docker
```

### Using Kubernetes (Production)

```bash
cd deploy
./deploy.sh deploy-k8s --namespace fingerprint
```

## Directory Structure

```
deploy/
├── deploy.sh              # Unified deployment script
├── docker/                # Docker Compose configurations
│   ├── docker-compose.yml
│   ├── Dockerfile
│   ├── prometheus.yml
│   └── grafana/
├── kubernetes/            # Kubernetes manifests
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── hpa.yaml
│   ├── networkpolicy.yaml
│   ├── rbac.yaml
│   ├── servicemonitor.yaml
│   └── kustomization.yaml
└── README.md
```

## Deployment Script Usage

```bash
./deploy.sh [COMMAND] [OPTIONS]

Commands:
    build           Build Docker image
    push            Push Docker image to registry
    deploy-docker   Deploy using Docker Compose
    deploy-k8s      Deploy to Kubernetes
    undeploy        Remove deployment
    status          Check deployment status
    logs            Show logs
    scale           Scale deployment (k8s only)
    test            Run smoke tests
    help            Show help message

Options:
    -v, --version   Set version tag (default: v3.0.0)
    -r, --registry  Set image registry (default: docker.io)
    -n, --namespace Set Kubernetes namespace (default: fingerprint)
```

## Docker Compose Deployment

### Requirements

- Docker 20.10+
- Docker Compose 2.0+

### Deploy

```bash
./deploy.sh deploy-docker
```

### Access Services

- API: http://localhost:8080
- gRPC: localhost:9090
- Metrics: http://localhost:8080/metrics

### ONNX Canary Defaults

The deployment manifests now include a safe default rollout profile:

- `FP_ML_INFERENCE_BACKEND=native`
- `FP_ML_CANARY_ENABLED=true`
- `FP_ML_CANARY_RATE=0.05`
- `FP_ML_CANARY_BACKEND=onnx`
- `FP_ML_SHADOW_COMPARE_ENABLED=true`

### View Logs

```bash
./deploy.sh logs -f
```

### Undeploy

```bash
./deploy.sh undeploy
```

## Kubernetes Deployment

### Requirements

- Kubernetes 1.25+
- kubectl configured
- kustomize (optional, kubectl includes it)
- Metrics server (for HPA)
- Prometheus Operator (for ServiceMonitor)

### Deploy

```bash
# Deploy to default namespace
./deploy.sh deploy-k8s

# Deploy to specific namespace
./deploy.sh deploy-k8s --namespace production

# Deploy specific version
./deploy.sh deploy-k8s --version v3.0.1
```

### Configuration

Edit `kubernetes/configmap.yaml` to customize configuration:

```yaml
data:
  config.yaml: |
    server:
      http_port: 8080
      grpc_port: 9090
    
    rate_limiter:
      enabled: true
      requests_per_second: 1000
    
    cache:
      max_size: 10000
      ttl: 5m
```

### Scaling

```bash
# Scale to 10 replicas
./deploy.sh scale --replicas 10

# Or use kubectl
kubectl -n fingerprint scale deployment fingerprint-gateway --replicas=10
```

### Monitoring

The deployment includes Prometheus ServiceMonitor for metrics collection.

Access metrics:
```bash
kubectl -n fingerprint port-forward svc/fingerprint-gateway 8080:8080
curl http://localhost:8080/metrics
```

### Troubleshooting

Check pod status:
```bash
./deploy.sh status
```

View logs:
```bash
# All pods
./deploy.sh logs

# Follow logs
./deploy.sh logs -f

# Specific pod
kubectl -n fingerprint logs -f fingerprint-gateway-xxxx
```

Run smoke tests:
```bash
./deploy.sh test
```

## Production Checklist

- [ ] Update image registry in deploy.sh
- [ ] Configure resource limits
- [ ] Set up TLS certificates
- [ ] Configure monitoring alerts
- [ ] Review NetworkPolicy rules
- [ ] Test disaster recovery
- [ ] Set up log aggregation

## Security Considerations

### NetworkPolicy

Default NetworkPolicy restricts:
- Ingress: Only from ingress-nginx and monitoring namespaces
- Egress: DNS and HTTPS only

Review and customize `kubernetes/networkpolicy.yaml` for your environment.

### RBAC

Minimal RBAC permissions granted:
- Read pods and configmaps in the same namespace

### Pod Security

- Non-root container user
- Read-only root filesystem
- No privileged containers
- Resource limits enforced

## Performance Tuning

### Horizontal Pod Autoscaler

Default HPA configuration:
- Min replicas: 3
- Max replicas: 20
- Target CPU: 70%
- Target Memory: 80%
- Target RPS: 1000

Edit `kubernetes/hpa.yaml` to adjust thresholds.

### Resource Limits

Default resources:
- Requests: 100m CPU, 128Mi memory
- Limits: 500m CPU, 512Mi memory

Adjust in `kubernetes/deployment.yaml` based on your workload.

## Advanced Configuration

### Custom Configuration

Create a custom ConfigMap:

```bash
kubectl -n fingerprint create configmap fingerprint-config-custom \
  --from-file=config.yaml=custom-config.yaml
```

Update deployment to use custom ConfigMap:

```bash
kubectl -n fingerprint set env deployment/fingerprint-gateway \
  FP_CONFIG_PATH=/config/custom-config.yaml
```

### Multi-Environment Deployment

Use kustomize overlays:

```bash
# staging/
kubernetes/overlays/staging/
├── kustomization.yaml
├── configmap-patch.yaml
└── deployment-patch.yaml

# production/
kubernetes/overlays/production/
├── kustomization.yaml
├── configmap-patch.yaml
└── deployment-patch.yaml
```

Deploy:

```bash
kustomize build kubernetes/overlays/production | kubectl apply -f -
```

## Troubleshooting

### Pods not starting

```bash
kubectl -n fingerprint describe pods -l app=fingerprint-gateway
kubectl -n fingerprint get events --field-selector type=Warning
```

### Service not accessible

```bash
kubectl -n fingerprint get svc
kubectl -n fingerprint port-forward svc/fingerprint-gateway 8080:8080
curl http://localhost:8080/health
```

### HPA not working

```bash
kubectl -n fingerprint get hpa
kubectl -n fingerprint top pods
```

Ensure metrics-server is installed:
```bash
kubectl get pods -n kube-system | grep metrics-server
```
