#!/bin/bash
set -e

echo "Starting benchmark tests..."

CONCURRENT_USERS=100
REQUESTS_PER_USER=1000
BASE_URL="http://localhost:8080"

run_benchmark() {
    local endpoint=$1
    local name=$2
    echo "Benchmarking $name..."
    
    wrk -t12 -c$CONCURRENT_USERS -d30s --latency \
        -s scripts/benchmark_$endpoint.lua $BASE_URL/api/v1/$endpoint
}

run_benchmark "nodes" "Nodes API"
run_benchmark "calculations/propagation" "Propagation Calculations"
run_benchmark "validation/timestamp" "Timestamp Validation"

echo "Benchmark completed"