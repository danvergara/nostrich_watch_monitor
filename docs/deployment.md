# Nostrich Watch Monitor - Deployment Guide

This guide walks you through deploying Nostrich Watch Monitor using Podman Quadlet files.

## Prerequisites

- Podman installed and configured
- systemd user services enabled
- Required environment variables set:
  - `NOSTRICH_WATCH_DB_PASSWORD`
  - `NOSTRICH_WATCH_MONITOR_PRIVATE_KEY`

## Deployment Steps

### 1. Clone the Repository

```bash
git clone https://github.com/danvergara/nostrich_watch_monitor.git
cd nostrich_watch_monitor
git checkout v0.3.0
```

### 2. Enable User Linger

Enable linger for your user to allow services to restart without being logged in:

```bash
make enable-linger
```

### 3. Create Secrets

Create the required Podman secrets from your environment variables:

```bash
make create-secrets
```

This will create:
- `nostrich-watch-db-password` - Database password secret
- `nostrich-watch-monitor-private-key` - Private key for Nostr monitoring

### 4. Setup Configuration

Copy the nostr-rs-relay configuration file to the expected location:

```bash
make setup-config
```

This copies `config.toml` to `~/.config/nostrich-watch/config.toml`.

### 5. Setup Services

Copy Quadlet files and reload systemd:

```bash
make setup-services
```

This will:
- Copy all `.container`, `.network`, and `.volume` files to `~/.config/containers/systemd/`
- Reload systemd daemon

### 6. Start Database

Start the PostgreSQL database service:

```bash
make setup-nostrich-watch-db
```

### 7. Run Migrations and Seeds

Execute database migrations and seed data:

```bash
./run-migrations.sh
```

This script will:
- Run database migrations
- Populate initial seed data

### 8. Start All Services

Start the remaining services:

```bash
make start-services
```

This starts:
- `nostrich-watch-cache` (Redis)
- `nostr-relay` (Nostr relay server)
- `job-scheduler` (Background job scheduler)
- `dashboard` (Web dashboard)
- `nostrich-watch-worker` (Background worker)
- `asynqmon` (Queue monitoring UI)

## Service URLs

After deployment, the following services will be available:

- **Dashboard**: http://localhost:8000
- **Asynq Monitor**: http://localhost:8080
- **Nostr Relay**: ws://localhost:7777
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379

## Managing Services

### Check Service Status
```bash
systemctl --user status nostrich-watch-db
systemctl --user status dashboard
# ... etc for other services
```

### View Service Logs
```bash
journalctl --user -u nostrich-watch-db -f
journalctl --user -u dashboard -f
```

### Stop Services
```bash
systemctl --user stop dashboard nostrich-watch-worker job-scheduler
```

### Restart Services
```bash
systemctl --user restart dashboard
```

## Troubleshooting

### Check if secrets exist
```bash
podman secret ls
```

### Verify network exists
```bash
podman network ls | grep monitor
```

### Check container status
```bash
podman ps -a
```

### View container logs
```bash
podman logs nostrich-watch-dashboard
```
