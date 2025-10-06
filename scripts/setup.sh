#!/bin/bash
set -e

echo "Setting up Relativistic SDK development environment"

mkdir -p logs
mkdir -p data
mkdir -p configs

cp .env.example .env

echo "Installing dependencies..."
go mod download
go mod verify

echo "Building application..."
go build -o bin/relativisticd ./cmd/relativisticd

echo "Running tests..."
go test ./... -v

echo "Setup completed successfully"