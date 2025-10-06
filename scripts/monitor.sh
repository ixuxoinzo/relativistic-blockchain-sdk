#!/bin/bash
set -e

echo "Starting monitoring..."

while true; do
    clear
    echo "=== Relativistic SDK Monitor ==="
    echo "Time: $(date)"
    echo ""
    
    curl -s http://localhost:8080/api/v1/health | jq .
    echo ""
    
    curl -s http://localhost:8080/api/v1/metrics | jq '.metrics | {total_nodes: .total_nodes, active_nodes: .active_nodes, network_coverage: .network_coverage}'
    
    sleep 30
done