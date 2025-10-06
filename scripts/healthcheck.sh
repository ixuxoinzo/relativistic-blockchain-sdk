#!/bin/bash
set -e

echo "Running health checks..."

check_endpoint() {
    local url=$1
    local name=$2
    if curl -f -s "$url" > /dev/null; then
        echo "✓ $name is healthy"
        return 0
    else
        echo "✗ $name is unhealthy"
        return 1
    fi
}

check_endpoint "http://localhost:8080/api/v1/health" "API Server"
check_endpoint "http://localhost:8080/api/v1/health/ready" "Readiness"
check_endpoint "http://localhost:9090/metrics" "Metrics"
check_endpoint "http://localhost:5432" "PostgreSQL"
check_endpoint "http://localhost:6379" "Redis"

echo "Health checks completed"