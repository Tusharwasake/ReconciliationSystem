# Deployment Guide

This guide provides instructions for deploying the Payment Reconciliation System with batch processing optimizations in production environments.

## Production Environment Setup

### System Requirements

#### Minimum Requirements

- **CPU**: 4 cores
- **RAM**: 8GB
- **Storage**: 100GB SSD
- **Network**: 1 Gbps
- **OS**: Linux (Ubuntu 20.04+ recommended)

#### Recommended Requirements

- **CPU**: 8+ cores
- **RAM**: 32GB+
- **Storage**: 500GB+ NVMe SSD
- **Network**: 10 Gbps
- **OS**: Linux (Ubuntu 22.04 LTS)

### Database Setup

#### PostgreSQL Configuration

1. **Install PostgreSQL 15+**:

```bash
sudo apt update
sudo apt install postgresql-15 postgresql-client-15 postgresql-contrib-15
```

2. **Optimize PostgreSQL for batch processing**:

```sql
-- /etc/postgresql/15/main/postgresql.conf
shared_buffers = '2GB'
work_mem = '256MB'
maintenance_work_mem = '1GB'
effective_cache_size = '8GB'
random_page_cost = 1.1
checkpoint_completion_target = 0.9
wal_buffers = '16MB'
max_connections = 200
```

3. **Create production database**:

```sql
CREATE DATABASE portdb_prod;
CREATE USER portuser WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE portdb_prod TO portuser;
```

### Application Deployment

#### Docker Deployment

1. **Create Dockerfile**:

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/schema.sql .

CMD ["./main"]
```

2. **Create docker-compose.yml**:

```yaml
version: "3.8"
services:
  app:
    build: .
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_USER=portuser
      - DB_PASSWORD=secure_password
      - DB_NAME=portdb_prod
      - BATCH_SIZE=5000
      - WORKER_COUNT=10
      - DB_MAX_OPEN_CONNS=25
      - DB_MAX_IDLE_CONNS=10
    volumes:
      - ./data:/app/data
      - ./output:/app/output
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=portdb_prod
      - POSTGRES_USER=portuser
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./schema.sql:/docker-entrypoint-initdb.d/schema.sql
    ports:
      - "5432:5432"
    restart: unless-stopped

volumes:
  postgres_data:
```

3. **Deploy with Docker Compose**:

```bash
docker-compose up -d
```

#### Kubernetes Deployment

1. **Create ConfigMap**:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: reconciliation-config
data:
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_USER: "portuser"
  DB_NAME: "portdb_prod"
  BATCH_SIZE: "5000"
  WORKER_COUNT: "10"
  DB_MAX_OPEN_CONNS: "25"
  DB_MAX_IDLE_CONNS: "10"
```

2. **Create Secret**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: reconciliation-secret
type: Opaque
stringData:
  DB_PASSWORD: "secure_password"
```

3. **Create Deployment**:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: reconciliation-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: reconciliation-app
  template:
    metadata:
      labels:
        app: reconciliation-app
    spec:
      containers:
        - name: reconciliation-app
          image: your-registry/reconciliation:latest
          envFrom:
            - configMapRef:
                name: reconciliation-config
            - secretRef:
                name: reconciliation-secret
          resources:
            requests:
              memory: "1Gi"
              cpu: "500m"
            limits:
              memory: "4Gi"
              cpu: "2000m"
          volumeMounts:
            - name: data-volume
              mountPath: /app/data
            - name: output-volume
              mountPath: /app/output
      volumes:
        - name: data-volume
          persistentVolumeClaim:
            claimName: data-pvc
        - name: output-volume
          persistentVolumeClaim:
            claimName: output-pvc
```

### Environment Configuration

#### Production Environment Variables

```bash
# Database Configuration
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=portuser
DB_PASSWORD=secure_password
DB_NAME=portdb_prod

# Performance Configuration
BATCH_SIZE=5000
WORKER_COUNT=10
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
CHUNK_SIZE=10000
MEMORY_LIMIT=1GB

# Monitoring Configuration
LOG_LEVEL=INFO
METRICS_ENABLED=true
METRICS_PORT=8080

# Security Configuration
TLS_ENABLED=true
TLS_CERT_FILE=/etc/ssl/certs/app.crt
TLS_KEY_FILE=/etc/ssl/private/app.key
```

#### Configuration Management

Use a configuration management tool like Ansible:

```yaml
- name: Deploy Reconciliation System
  hosts: production
  vars:
    app_version: "1.0.0"
    db_password: "{{ vault_db_password }}"
  tasks:
    - name: Update system packages
      apt:
        update_cache: yes
        upgrade: yes

    - name: Install Docker
      apt:
        name: docker.io
        state: present

    - name: Deploy application
      docker_compose:
        project_src: /opt/reconciliation
        state: present
```

## Security Configuration

### Database Security

1. **Enable SSL/TLS**:

```sql
-- postgresql.conf
ssl = on
ssl_cert_file = '/etc/ssl/certs/server.crt'
ssl_key_file = '/etc/ssl/private/server.key'
```

2. **Configure authentication**:

```
# pg_hba.conf
hostssl all all 0.0.0.0/0 md5
```

3. **Set up firewall rules**:

```bash
# Allow PostgreSQL only from application servers
sudo ufw allow from 10.0.1.0/24 to any port 5432
```

### Application Security

1. **Enable TLS for API endpoints**:

```go
// Add to main.go
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS12,
}

server := &http.Server{
    Addr:      ":8443",
    TLSConfig: tlsConfig,
}
```

2. **Implement rate limiting**:

```go
// Add rate limiting middleware
limiter := rate.NewLimiter(10, 100) // 10 requests per second, burst of 100
```

3. **Add input validation**:

```go
// Validate file inputs
if !isValidFile(filename) {
    return fmt.Errorf("invalid file format")
}
```

## Monitoring and Logging

### Application Monitoring

1. **Add Prometheus metrics**:

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    recordsProcessed = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "records_processed_total",
            Help: "Total number of records processed",
        },
        []string{"type"},
    )

    processingDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "processing_duration_seconds",
            Help: "Time spent processing records",
        },
        []string{"operation"},
    )
)
```

2. **Set up health checks**:

```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    // Check database connection
    if err := db.Ping(); err != nil {
        http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

### Log Management

1. **Structured logging**:

```go
import "github.com/sirupsen/logrus"

log := logrus.WithFields(logrus.Fields{
    "batch_id": batchID,
    "records":  recordCount,
    "duration": duration,
})
log.Info("Batch processed successfully")
```

2. **Log aggregation with ELK stack**:

```yaml
# docker-compose.yml
elasticsearch:
  image: docker.elastic.co/elasticsearch/elasticsearch:8.5.0
  environment:
    - discovery.type=single-node
    - "ES_JAVA_OPTS=-Xms1g -Xmx1g"

logstash:
  image: docker.elastic.co/logstash/logstash:8.5.0
  volumes:
    - ./logstash.conf:/usr/share/logstash/pipeline/logstash.conf

kibana:
  image: docker.elastic.co/kibana/kibana:8.5.0
  ports:
    - "5601:5601"
```

## Backup and Recovery

### Database Backup

1. **Automated backups**:

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="/var/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump -h localhost -U portuser -d portdb_prod > "$BACKUP_DIR/backup_$DATE.sql"
```

2. **Backup retention**:

```bash
# Keep daily backups for 30 days
find $BACKUP_DIR -name "backup_*.sql" -type f -mtime +30 -delete
```

### Application Data Backup

1. **File system backups**:

```bash
# Backup data and output directories
tar -czf /backups/app_data_$(date +%Y%m%d).tar.gz /app/data /app/output
```

2. **Recovery procedures**:

```bash
# Restore database
psql -h localhost -U portuser -d portdb_prod < backup_20240115_120000.sql

# Restore application data
tar -xzf app_data_20240115.tar.gz -C /
```

## Performance Monitoring

### Key Metrics to Monitor

1. **Application Metrics**:

   - Records processed per second
   - Memory usage
   - CPU utilization
   - Error rates

2. **Database Metrics**:

   - Connection pool usage
   - Query execution time
   - Lock contention
   - Cache hit ratio

3. **System Metrics**:
   - Disk I/O
   - Network throughput
   - Available memory
   - CPU load average

### Alerting Rules

```yaml
# Prometheus alerting rules
groups:
  - name: reconciliation.rules
    rules:
      - alert: HighMemoryUsage
        expr: (process_resident_memory_bytes / 1024 / 1024) > 1000
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage detected"

      - alert: SlowDatabaseQueries
        expr: postgres_stat_database_tup_fetched_rate5m > 10000
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Database queries are running slowly"
```

## Scaling Considerations

### Horizontal Scaling

1. **Load balancing**:

```nginx
upstream reconciliation_backend {
    server app1:8080;
    server app2:8080;
    server app3:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://reconciliation_backend;
    }
}
```

2. **Database read replicas**:

```yaml
# Add read replica configuration
read_replica:
  image: postgres:15
  environment:
    - POSTGRES_MASTER_SERVICE=db
    - POSTGRES_REPLICATION_USER=replicator
    - POSTGRES_REPLICATION_PASSWORD=repl_password
```

### Vertical Scaling

1. **Resource optimization**:

```yaml
# Kubernetes resource limits
resources:
  requests:
    memory: "4Gi"
    cpu: "2000m"
  limits:
    memory: "8Gi"
    cpu: "4000m"
```

2. **Database scaling**:

```sql
-- Increase database resources
ALTER SYSTEM SET shared_buffers = '4GB';
ALTER SYSTEM SET work_mem = '512MB';
ALTER SYSTEM SET maintenance_work_mem = '2GB';
```

## Maintenance Procedures

### Regular Maintenance

1. **Database maintenance**:

```sql
-- Weekly maintenance
VACUUM ANALYZE;
REINDEX DATABASE portdb_prod;
```

2. **Log rotation**:

```bash
# /etc/logrotate.d/reconciliation
/var/log/reconciliation/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 644 app app
}
```

3. **System updates**:

```bash
# Monthly system updates
sudo apt update && sudo apt upgrade -y
sudo docker system prune -f
```

### Troubleshooting Guide

1. **High memory usage**:

   - Check for memory leaks
   - Adjust batch sizes
   - Monitor garbage collection

2. **Database connection issues**:

   - Check connection pool settings
   - Monitor active connections
   - Verify network connectivity

3. **Performance degradation**:
   - Check system resources
   - Analyze slow queries
   - Review application logs

This deployment guide provides a comprehensive approach to deploying the batch processing optimized reconciliation system in production environments with proper security, monitoring, and scaling considerations.
