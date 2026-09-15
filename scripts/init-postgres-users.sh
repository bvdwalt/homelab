#!/usr/bin/env bash
# Run after a CNPG cluster recreation to restore all app database users.
# Reads passwords from existing k8s secrets so nothing needs to be looked up manually.
set -euo pipefail

PSQL="kubectl --context=altair -n postgresql exec -i postgresql-1 -- psql -U postgres"

echo "==> Fetching passwords from secrets..."

ATUIN_PASS=$(kubectl --context=altair -n atuin get secret atuin \
  -o jsonpath='{.data.ATUIN_DB_URI}' | base64 -d \
  | python3 -c "import sys,urllib.parse; u=urllib.parse.urlparse(sys.stdin.read().strip()); print(u.password)")

METERING_PASS=$(kubectl --context=altair -n metering get secret metering-secrets \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)

JELLYSTAT_PASS=$(kubectl --context=altair -n jellystat get secret jellystat-secrets \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)

INVENTORY_PASS=$(kubectl --context=altair -n order-pipeline get secret order-pipeline-inventory-postgres \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)

SHIPPING_PASS=$(kubectl --context=altair -n order-pipeline get secret order-pipeline-shipping-postgres \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)

NOTIFICATION_PASS=$(kubectl --context=altair -n order-pipeline get secret order-pipeline-notification-postgres \
  -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)

echo "==> Creating databases that don't already exist..."

$PSQL -tc "SELECT 1 FROM pg_database WHERE datname = 'jellystat'" | grep -q 1 \
  || $PSQL -c "CREATE DATABASE jellystat"

$PSQL -tc "SELECT 1 FROM pg_database WHERE datname = 'order_pipeline'" | grep -q 1 \
  || $PSQL -c "CREATE DATABASE order_pipeline"

echo "==> Creating users..."

$PSQL <<SQL
DO \$\$
BEGIN
  CREATE USER atuin WITH PASSWORD '${ATUIN_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER atuin WITH PASSWORD '${ATUIN_PASS}';
END \$\$;
GRANT ALL PRIVILEGES ON DATABASE atuin TO atuin;
ALTER DATABASE atuin OWNER TO atuin;

DO \$\$
BEGIN
  CREATE USER metering WITH PASSWORD '${METERING_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER metering WITH PASSWORD '${METERING_PASS}';
END \$\$;
GRANT ALL PRIVILEGES ON DATABASE data TO metering;
ALTER DATABASE data OWNER TO metering;

DO \$\$
BEGIN
  CREATE USER jellystat WITH PASSWORD '${JELLYSTAT_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER jellystat WITH PASSWORD '${JELLYSTAT_PASS}';
END \$\$;
GRANT ALL PRIVILEGES ON DATABASE jellystat TO jellystat;
ALTER DATABASE jellystat OWNER TO jellystat;

DO \$\$
BEGIN
  CREATE USER inventory_service WITH PASSWORD '${INVENTORY_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER inventory_service WITH PASSWORD '${INVENTORY_PASS}';
END \$\$;
GRANT CONNECT ON DATABASE order_pipeline TO inventory_service;

DO \$\$
BEGIN
  CREATE USER shipping_service WITH PASSWORD '${SHIPPING_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER shipping_service WITH PASSWORD '${SHIPPING_PASS}';
END \$\$;
GRANT CONNECT ON DATABASE order_pipeline TO shipping_service;

DO \$\$
BEGIN
  CREATE USER notification_service WITH PASSWORD '${NOTIFICATION_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER notification_service WITH PASSWORD '${NOTIFICATION_PASS}';
END \$\$;
GRANT CONNECT ON DATABASE order_pipeline TO notification_service;
SQL

echo "==> Pre-creating shared schema_migrations table for order_pipeline..."

# Pre-created as postgres so no single service user ends up owning this shared table.
$PSQL -d order_pipeline <<SQL
CREATE TABLE IF NOT EXISTS schema_migrations (
  name TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
GRANT SELECT, INSERT ON schema_migrations TO inventory_service, shipping_service, notification_service;
-- CREATE TABLE IF NOT EXISTS still requires schema CREATE privilege even when the table already exists.
GRANT CREATE ON SCHEMA public TO inventory_service, shipping_service, notification_service;
SQL

echo "==> Granting table privileges..."

$PSQL -d atuin <<SQL
DO \$\$ DECLARE r RECORD; BEGIN
  FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' LOOP
    EXECUTE 'ALTER TABLE public.' || quote_ident(r.tablename) || ' OWNER TO atuin';
  END LOOP;
  FOR r IN SELECT sequence_name FROM information_schema.sequences WHERE sequence_schema = 'public' LOOP
    EXECUTE 'ALTER SEQUENCE public.' || quote_ident(r.sequence_name) || ' OWNER TO atuin';
  END LOOP;
END \$\$;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO atuin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO atuin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO atuin;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO atuin;
SQL

$PSQL -d data <<SQL
DO \$\$ DECLARE r RECORD; BEGIN
  FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public' LOOP
    EXECUTE 'ALTER TABLE public.' || quote_ident(r.tablename) || ' OWNER TO metering';
  END LOOP;
  FOR r IN SELECT sequence_name FROM information_schema.sequences WHERE sequence_schema = 'public' LOOP
    EXECUTE 'ALTER SEQUENCE public.' || quote_ident(r.sequence_name) || ' OWNER TO metering';
  END LOOP;
END \$\$;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO metering;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO metering;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO metering;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO metering;
SQL

$PSQL -d jellystat <<SQL
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO jellystat;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO jellystat;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO jellystat;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO jellystat;
SQL

echo "==> Done. Restart affected deployments if needed:"
echo "    kubectl --context=altair -n atuin rollout restart deployment/atuin"
echo "    kubectl --context=altair -n metering rollout restart deployment/metering-api"
echo "    kubectl --context=altair -n jellystat rollout restart deployment/jellystat"
echo "    kubectl --context=altair -n order-pipeline rollout restart deployment/inventory-service deployment/shipping-service deployment/notification-service"
