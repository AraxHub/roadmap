#!/bin/bash
# Runs only on first Postgres data dir init (empty volume).
set -euo pipefail

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
	-- Roadmap
	DO \$\$
	BEGIN
	  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'roadmap_app') THEN
	    CREATE ROLE roadmap_app LOGIN PASSWORD '${ROADMAP_DB_PASSWORD}';
	  END IF;
	END
	\$\$;

	SELECT 'CREATE DATABASE roadmap OWNER roadmap_app'
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'roadmap')\gexec

	GRANT ALL PRIVILEGES ON DATABASE roadmap TO roadmap_app;
EOSQL

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname roadmap <<-EOSQL
	GRANT ALL ON SCHEMA public TO roadmap_app;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO roadmap_app;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO roadmap_app;
EOSQL

# Optional Voco DB — only if VOCO_DB_PASSWORD is set at first init.
if [[ -n "${VOCO_DB_PASSWORD:-}" ]]; then
  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres <<-EOSQL
	DO \$\$
	BEGIN
	  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'voco_app') THEN
	    CREATE ROLE voco_app LOGIN PASSWORD '${VOCO_DB_PASSWORD}';
	  END IF;
	END
	\$\$;

	SELECT 'CREATE DATABASE voco OWNER voco_app'
	WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'voco')\gexec

	GRANT ALL PRIVILEGES ON DATABASE voco TO voco_app;
EOSQL

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname voco <<-EOSQL
	GRANT ALL ON SCHEMA public TO voco_app;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO voco_app;
	ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO voco_app;
EOSQL
fi

echo "init databases: done"
