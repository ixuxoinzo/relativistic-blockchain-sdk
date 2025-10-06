```markdown
# Deployment Guide

## Prerequisites
- Docker and Docker Compose
- Kubernetes cluster (for production)
- PostgreSQL 14+
- Redis 7+
- 2GB RAM minimum, 8GB recommended
- 10GB disk space

## Quick Start

### Local Development
```bash
# Clone repository
git clone https://github.com/ixuxoinzo/relativistic-blockchain-sdk
cd relativistic-blockchain-sdk

# Setup environment
cp .env.example .env
# Edit .env with your configuration

# Run setup script
./scripts/setup.sh

# Start services
docker-compose -f deployments/docker-compose.yml up -d

# Verify deployment
curl http://localhost:8080/api/v1/health
```

Production Deployment

```bash
# Use deploy script
./scripts/deploy.sh production latest

# Or manually with Docker Compose
docker-compose -f deployments/docker-compose.prod.yml up -d
```

Configuration

Environment Variables

Create .env file:

```bash
# Server Configuration
RELATIVISTIC_ENVIRONMENT=production
SERVER_ADDRESS=:8080
LOG_LEVEL=info

# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=relativistic
POSTGRES_USER=postgres
POSTGRES_PASSWORD=your-secret-password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password

# Security
JWT_SECRET=your-32-character-secret-key-here
API_KEY=your-api-key-for-external-access

# Monitoring
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
```

Configuration Files

Update configs/config.prod.yaml for production:

```yaml
server:
  address: ":8080"
  environment: "production"
  log_level: "info"
  cors_allowed_origins:
    - "https://yourdomain.com"

database:
  host: "postgresql.example.com"
  port: 5432
  name: "relativistic"
  username: "relativistic_user"
  ssl_mode: "require"
  max_connections: 100
  max_idle_connections: 10

cache:
  host: "redis.example.com"
  port: 6379
  password: "your-redis-password"
  db: 0
  pool_size: 100

security:
  jwt_secret: "your-32-character-secret"
  jwt_expiry: "24h"
  api_key: "your-api-key"
  rate_limit: 1000

monitoring:
  prometheus:
    enabled: true
    path: "/metrics"
  health_check:
    enabled: true
    interval: "30s"

network:
  node_registration:
    auto_approve: true
    max_nodes: 1000
  propagation:
    max_calculation_time: "5s"
    batch_size: 100
```

Kubernetes Deployment

Prerequisites

· Kubernetes cluster 1.24+
· Helm 3.8+
· NGINX Ingress Controller
· Cert-Manager (for TLS)

Deploy with Helm

```bash
# Add repository
helm repo add relativistic https://charts.relativistic.io
helm repo update

# Install
helm install relativistic-sdk relativistic/relativistic-sdk \
  --namespace relativistic \
  --create-namespace \
  --values deployments/kubernetes/values.yaml
```

Manual Kubernetes Deployment

```bash
# Create namespace
kubectl create namespace relativistic

# Deploy secrets
kubectl apply -f deployments/kubernetes/secrets.yaml

# Deploy configuration
kubectl apply -f deployments/kubernetes/configmap.yaml

# Deploy database
kubectl apply -f deployments/kubernetes/postgresql.yaml

# Deploy Redis
kubectl apply -f deployments/kubernetes/redis.yaml

# Deploy application
kubectl apply -f deployments/kubernetes/deployment.yaml

# Deploy services
kubectl apply -f deployments/kubernetes/service.yaml

# Deploy ingress
kubectl apply -f deployments/kubernetes/ingress.yaml

# Verify deployment
kubectl get all -n relativistic
```

Kubernetes Configuration Files

deployments/kubernetes/secrets.yaml

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: relativistic-secrets
  namespace: relativistic
type: Opaque
data:
  postgres-password: <base64-encoded>
  redis-password: <base64-encoded>
  jwt-secret: <base64-encoded>
  api-key: <base64-encoded>
```

deployments/kubernetes/deployment.yaml

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: relativistic-sdk
  namespace: relativistic
  labels:
    app: relativistic-sdk
spec:
  replicas: 3
  selector:
    matchLabels:
      app: relativistic-sdk
  template:
    metadata:
      labels:
        app: relativistic-sdk
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: relativistic-sdk
        image: relativistic/sdk:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: metrics
        env:
        - name: RELATIVISTIC_ENVIRONMENT
          value: "production"
        - name: POSTGRES_HOST
          value: "relativistic-postgresql"
        - name: REDIS_HOST
          value: "relativistic-redis"
        envFrom:
        - secretRef:
            name: relativistic-secrets
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: http
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /api/v1/health/ready
            port: http
          initialDelaySeconds: 5
          periodSeconds: 5
```

Monitoring Setup

Access Dashboards

· Grafana: http://localhost:3000 (admin/admin)
· Prometheus: http://localhost:9090
· API Documentation: http://localhost:8080/docs

Pre-configured Dashboards

The deployment includes these Grafana dashboards:

· Relativistic SDK Overview
· Network Metrics
· Consensus Metrics
· Node Performance

Custom Metrics

Add custom business metrics:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    propagationCalculations = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "relativistic_propagation_calculations_total",
            Help: "Total number of propagation calculations",
        },
        []string{"source_region", "target_region"},
    )
)
```

Scaling

Horizontal Pod Autoscaler

```yaml
# deployments/kubernetes/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: relativistic-sdk
  namespace: relativistic
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: relativistic-sdk
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: relativistic_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
```

Database Scaling

```bash
# PostgreSQL connection pooling with PgBouncer
kubectl apply -f deployments/kubernetes/pgbouncer.yaml

# Redis cluster for high availability
kubectl apply -f deployments/kubernetes/redis-cluster.yaml
```

Backup and Recovery

Automated Backups

```bash
# Run backup script
./scripts/backup.sh

# Schedule daily backups with cron
0 2 * * * /path/to/relativistic-sdk/scripts/backup.sh
```

Kubernetes Backup with Velero

```bash
# Install Velero
velero install --provider aws --plugins velero/velero-plugin-for-aws:v1.0.0 --bucket relativistic-backups --secret-file ./credentials-velero

# Create backup schedule
velero create schedule relativistic-daily --schedule="0 2 * * *" --include-namespaces relativistic
```

Database Backup

```yaml
# deployments/kubernetes/postgresql-backup.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgresql-backup
  namespace: relativistic
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: postgres:14
            command:
            - /bin/bash
            - -c
            - |
              pg_dump -h relativistic-postgresql -U $POSTGRES_USER $POSTGRES_DB > /backup/backup-$(date +%Y%m%d).sql
            env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: relativistic-secrets
                  key: postgres-password
            volumeMounts:
            - name: backup-volume
              mountPath: /backup
          restartPolicy: OnFailure
          volumes:
          - name: backup-volume
            persistentVolumeClaim:
              claimName: backup-pvc
```

Security

Network Policies

```yaml
# deployments/kubernetes/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: relativistic-sdk-policy
  namespace: relativistic
spec:
  podSelector:
    matchLabels:
      app: relativistic-sdk
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: relativistic-postgresql
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: relativistic-redis
    ports:
    - protocol: TCP
      port: 6379
```

TLS Configuration

```yaml
# deployments/kubernetes/ingress-tls.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: relativistic-sdk
  namespace: relativistic
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - relativistic.example.com
    secretName: relativistic-tls
  rules:
  - host: relativistic.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: relativistic-sdk
            port:
              number: 8080
```

Troubleshooting

Common Issues

1. Database Connection Issues

```bash
# Check PostgreSQL status
kubectl logs -l app=relativistic-postgresql -n relativistic

# Test connection
kubectl exec -it deployment/relativistic-sdk -n relativistic -- \
  nc -zv relativistic-postgresql 5432
```

1. High Memory Usage

```bash
# Check memory usage
kubectl top pods -n relativistic

# Analyze memory profile
curl http://localhost:8080/debug/pprof/heap > heap.pprof
```

1. Node Registration Failures

```bash
# Check node registration logs
kubectl logs -l app=relativistic-sdk -n relativistic --tail=100

# Verify Redis connectivity
kubectl exec -it deployment/relativistic-sdk -n relativistic -- \
  redis-cli -h relativistic-redis ping
```

Logs and Debugging

```bash
# View all logs
kubectl logs -l app=relativistic-sdk -n relativistic

# Stream logs in real-time
kubectl logs -f deployment/relativistic-sdk -n relativistic

# Debug with shell access
kubectl exec -it deployment/relativistic-sdk -n relativistic -- /bin/sh

# Check resource usage
kubectl top pods -n relativistic
```

Performance Tuning

Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX idx_nodes_position ON nodes USING gist (position);
CREATE INDEX idx_nodes_active ON nodes(is_active) WHERE is_active = true;
CREATE INDEX idx_blocks_timestamp ON blocks(timestamp);
```

Redis Optimization

```yaml
# deployments/kubernetes/redis-optimized.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: relativistic-redis
spec:
  template:
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        command:
        - redis-server
        - "--maxmemory 1gb"
        - "--maxmemory-policy allkeys-lru"
        - "--save 900 1"
        - "--save 300 10"
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
```

Migration Guide

Version 1.0.0 to 1.1.0

```bash
# Backup current data
./scripts/backup.sh

# Update configuration
cp configs/config.v1.1.0.yaml configs/config.yaml

# Run migration
kubectl apply -f deployments/kubernetes/migration-job.yaml

# Verify migration
kubectl logs -l job-name=relativistic-migration -n relativistic
```

This deployment guide covers all aspects of deploying the Relativistic SDK in both development and production environments.

```
