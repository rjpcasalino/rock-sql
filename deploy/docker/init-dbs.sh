#!/bin/bash
set -e

echo "=== Initializing PostgreSQL for rock-sql ==="

# Create explorer user with password
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'explorer') THEN
            CREATE ROLE explorer WITH LOGIN PASSWORD 'explorer' CREATEDB;
        END IF;
    END
    \$\$;
EOSQL

# Databases defined in rock-sql
DATABASES=(
    "astronomy"
    "math"
    "anatomy"
    "elements"
    "library"
    "pagila"
    "chinook"
    "employees"
    "lego"
    "netflix"
    "titanic"
    "happiness_index"
    "misc"
)

for db in "${DATABASES[@]}"; do
    echo "Checking database: $db"
    psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
        SELECT 'CREATE DATABASE $db OWNER explorer'
        WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec
        GRANT ALL PRIVILEGES ON DATABASE "$db" TO explorer;
EOSQL
done

# Apply available seeds
for seed in /seeds/*_seed.sql; do
    if [ -f "$seed" ]; then
        filename=$(basename "$seed")
        dbname="${filename%_seed.sql}"
        echo "Applying seed $seed to database '$dbname'..."
        psql -v ON_ERROR_STOP=1 --username explorer --dbname "$dbname" -f "$seed" || echo "Warning: failed seeding $seed"
    fi
done

echo "=== Database initialization complete ==="
