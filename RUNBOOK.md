# rock-sql Runbook & Deployment Guide

This guide covers everything needed to build, run, develop, and deploy [`rock-sql`](./main.go) across local environments (macOS, Linux, Windows) and production environments (specifically **OpenBSD**).

---

## 1. Quick Start (Local Development)

### Prerequisites
- [Go](https://go.dev/) (version 1.22+)
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/)
- POSIX `make`

### 1-Command Local Dev
To start the local PostgreSQL instance, initialize databases, run seed scripts, and launch the server:
```bash
make dev
```
Open [http://localhost:8080](http://localhost:8080) in your browser.

To stop the background PostgreSQL container:
```bash
make db-down
```

---

## 2. Environment Variables & Configuration

`rock-sql` is configured via environment variables with sensible defaults:

| Variable | Default | Purpose |
| :--- | :--- | :--- |
| `PGHOST` | `127.0.0.1` | PostgreSQL hostname, IP, or Unix socket path |
| `PGPORT` | `5432` | PostgreSQL listening port |
| `PGUSER` | `explorer` | Database username |
| `PGPASSWORD` | `explorer` | Database user password |
| `PGSSLMODE` | `disable` | PostgreSQL SSL connection mode (`disable`, `require`, etc.) |
| `PORT` | `8080` | Web server listening HTTP port |
| `LISTEN_ADDR` | `:<PORT>` | Full listen address (e.g. `127.0.0.1:8080` or `:8080`) |

---

## 3. Makefile Commands

Run `make help` to inspect available build targets:

```text
Usage: make [target]

Targets:
  all              Build binary for current operating system (default)
  help             Show this help message
  build            Build binary for current platform (to bin/rock-sql)
  build-openbsd    Cross-compile static binary for OpenBSD amd64 (to dist/openbsd_amd64/rock-sql)
  run              Build and run locally
  dev              Start local PostgreSQL container and launch application
  db-up            Start PostgreSQL service in background with seeds mounted
  db-down          Stop PostgreSQL container
  db-logs          Follow PostgreSQL container logs
  db-seed          Re-apply SQL seed files to the running PostgreSQL container
  test             Run unit tests
  vet              Run go vet
  fmt              Format Go source files
  clean            Clean build artifacts
```

---

## 4. Deploying to OpenBSD

`rock-sql` runs on OpenBSD with zero external runtime dependencies because it builds as a pure Go static binary without cgo.

### Step 1: Cross-compile for OpenBSD
From your local workstation (macOS, Linux, etc.):
```bash
make build-openbsd
```
This generates `dist/openbsd_amd64/rock-sql`.

### Step 2: Copy Binary & rc.d Script to OpenBSD
```bash
scp dist/openbsd_amd64/rock-sql root@rock.home.arpa:/usr/local/bin/rock-sql
scp deploy/openbsd/rocksql root@rock.home.arpa:/etc/rc.d/rocksql
```

### Step 3: Setup User & Permissions on OpenBSD
On the OpenBSD host (`rock.home.arpa`):
```ksh
# Create dedicated unprivileged daemon user if not already present
useradd -s /sbin/nologin -d /var/empty -c "rock-sql daemon" _rocksql

# Set correct ownership and permissions
chmod 755 /usr/local/bin/rock-sql
chmod 555 /etc/rc.d/rocksql
chown root:bin /usr/local/bin/rock-sql /etc/rc.d/rocksql
```

### Step 4: Configure OpenBSD PostgreSQL
If PostgreSQL is running locally on OpenBSD:
```ksh
# On OpenBSD, postgresql-server runs as _postgresql
# Make sure the explorer user exists and can connect:
su - _postgresql
createuser -P explorer   # password: explorer
createdb -O explorer astronomy
createdb -O explorer math
createdb -O explorer anatomy
exit
```
Apply seeds if needed:
```ksh
psql -U explorer -d astronomy -f astronomy_seed.sql
psql -U explorer -d math -f math_seed.sql
psql -U explorer -d anatomy -f anatomy_seed.sql
```

### Step 5: Enable and Start Service with rcctl(8)
```ksh
rcctl enable rocksql
rcctl start rocksql
```
Check service status or logs:
```ksh
rcctl check rocksql
tail -f /var/log/daemon
```

---

## 5. Running Natively on Other Operating Systems

If you prefer not using Docker and have PostgreSQL installed locally on your system:

### Linux / macOS
```bash
# Set your local Postgres connection credentials
export PGHOST=127.0.0.1
export PGPORT=5432
export PGUSER=explorer
export PGPASSWORD=explorer

# Run tests and start
make test
make run
```
