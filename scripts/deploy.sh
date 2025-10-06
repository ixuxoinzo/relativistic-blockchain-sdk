#!/bin/bash
set -e

DEPLOY_ENV=${1:-production}
VERSION=${2:-latest}

echo "Deploying Relativistic SDK version $VERSION to $DEPLOY_ENV"

export RELATIVISTIC_ENVIRONMENT=$DEPLOY_ENV

docker-compose -f deployments/docker-compose.yml pull relativistic-sdk
docker-compose -f deployments/docker-compose.yml up -d relativistic-sdk

sleep 30

echo "Running health checks..."
curl -f http://localhost:8080/api/v1/health || exit 1
curl -f http://localhost:8080/api/v1/health/ready || exit 1

echo "Deployment completed successfully"