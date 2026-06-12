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

LINKWARDEN_PASS=$(kubectl --context=raspi -n linkwarden get secret linkwarden \
  -o jsonpath='{.data.DATABASE_URL}' | base64 -d \
  | python3 -c "import sys,urllib.parse; u=urllib.parse.urlparse(sys.stdin.read().strip()); print(u.password)")

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
  CREATE USER linkwarden WITH PASSWORD '${LINKWARDEN_PASS}';
EXCEPTION WHEN duplicate_object THEN
  ALTER USER linkwarden WITH PASSWORD '${LINKWARDEN_PASS}';
END \$\$;
GRANT ALL PRIVILEGES ON DATABASE linkwarden TO linkwarden;
ALTER DATABASE linkwarden OWNER TO linkwarden;
SQL

echo "==> Granting table privileges..."

$PSQL -d atuin <<SQL
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

$PSQL -d linkwarden <<SQL
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO linkwarden;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO linkwarden;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO linkwarden;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO linkwarden;
SQL

echo "==> Done. Restart affected deployments if needed:"
echo "    kubectl --context=altair -n atuin rollout restart deployment/atuin"
echo "    kubectl --context=altair -n metering rollout restart deployment/metering-api"
echo "    kubectl --context=raspi -n linkwarden rollout restart deployment/linkwarden"
