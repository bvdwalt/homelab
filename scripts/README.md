# scripts/

Runnable scripts for cluster maintenance tasks.

---

## init-postgres-users.sh

Re-creates all application database users on the Altair CNPG PostgreSQL cluster.

**When to run:** after the CNPG cluster is recreated (e.g. fresh deploy, disaster recovery). CNPG provisions the cluster and the `postgres` superuser, but does not create app-specific users — this script fills that gap.

**Prerequisites:**

- `kubectl` contexts `altair` and `raspi` must be configured and reachable.
- The CNPG cluster must be healthy (`postgresql-1` pod running).
- The app secrets that hold passwords must already exist:
  - `altair/atuin` → `ATUIN_DB_URI`
  - `altair/metering` → `POSTGRES_PASSWORD` (from `metering-secrets`)
  - `raspi/linkwarden` → `DATABASE_URL`
- The target databases (`atuin`, `data`, `linkwarden`) must already exist (CNPG creates them from the cluster bootstrap config).

**Usage:**

```bash
bash scripts/init-postgres-users.sh
```

The script is idempotent: it uses `DO $$ BEGIN … EXCEPTION WHEN duplicate_object` blocks so re-running it on an existing cluster only updates passwords without erroring.

**After running**, restart affected workloads if the passwords changed:

```bash
kubectl --context=altair -n atuin rollout restart deployment/atuin
kubectl --context=altair -n metering rollout restart deployment/metering-api
kubectl --context=raspi -n linkwarden rollout restart deployment/linkwarden
```
