#!/bin/bash

# Run database migrations
echo "Running database migrations..."
podman run --rm --network systemd-monitor \
  -e NOSTRICH_WATCH_DB_HOST=systemd-nostrich-watch-db \
  -e NOSTRICH_WATCH_DB_PORT=5432 \
  -e NOSTRICH_WATCH_DB_USER=monitor \
  -e NOSTRICH_WATCH_DB_PASSWORD="$NOSTRICH_WATCH_DB_PASSWORD" \
  -e NOSTRICH_WATCH_DB_NAME=monitor \
  ghcr.io/danvergara/nostrich-watch-monitor:0.2.0 /app/monitor migrate up

# Run seeder
echo "Running database seeder..."
podman run --rm --network systemd-monitor \
  -e NOSTRICH_WATCH_DB_HOST=systemd-nostrich-watch-db \
  -e NOSTRICH_WATCH_DB_PORT=5432 \
  -e NOSTRICH_WATCH_DB_USER=monitor \
  -e NOSTRICH_WATCH_DB_PASSWORD="$NOSTRICH_WATCH_DB_PASSWORD" \
  -e NOSTRICH_WATCH_DB_NAME=monitor \
  ghcr.io/danvergara/nostrich-watch-monitor:0.2.0 /app/monitor seeds

echo "Migrations and seeding completed!"
