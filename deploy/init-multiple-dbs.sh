#!/bin/sh
set -eu

if [ -z "${POSTGRES_MULTIPLE_DATABASES:-}" ]; then
  echo "[init-db] POSTGRES_MULTIPLE_DATABASES not set, skipping"
  exit 0
fi

for db in $(echo "${POSTGRES_MULTIPLE_DATABASES}" | tr ',' ' '); do
  [ -z "$db" ] && continue

  echo "[init-db] -> $db"
  psql -v ON_ERROR_STOP=1 --username "${POSTGRES_USER}" --dbname postgres <<-EOSQL
    SELECT 'CREATE DATABASE "${db}"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '${db}')\gexec
    GRANT ALL PRIVILEGES ON DATABASE "${db}" TO "${POSTGRES_USER}";
EOSQL
done

echo "[init-db] done"

