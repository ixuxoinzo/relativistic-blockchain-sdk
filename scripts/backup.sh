#!/bin/bash
set -e

BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

echo "Starting backup to $BACKUP_DIR"

docker exec relativistic-postgres-1 pg_dump -U postgres relativistic > $BACKUP_DIR/database.sql
docker exec relativistic-redis-1 redis-cli SAVE
docker cp relativistic-redis-1:/data/dump.rdb $BACKUP_DIR/redis.rdb

tar -czf $BACKUP_DIR/backup.tar.gz $BACKUP_DIR/*

echo "Backup completed: $BACKUP_DIR/backup.tar.gz"