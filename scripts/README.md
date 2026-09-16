# scripts/

Runnable scripts for cluster maintenance tasks.

---

## init-postgres-users.sh

Re-creates all application database users on the Altair CNPG PostgreSQL cluster.

**When to run:** after the CNPG cluster is recreated (e.g. fresh deploy, disaster recovery). CNPG provisions the cluster and the `postgres` superuser, but does not create app-specific users — this script fills that gap.

**Prerequisites:**

- `kubectl` context `altair` must be configured and reachable.
- The CNPG cluster must be healthy (`postgresql-1` pod running).
- The app secrets that hold passwords must already exist:
  - `altair/atuin` → `ATUIN_DB_URI`
  - `altair/metering` → `POSTGRES_PASSWORD` (from `metering-secrets`)
- The target databases (`atuin`, `data`) must already exist (CNPG creates them from the cluster bootstrap config).

**Usage:**

```bash
bash scripts/init-postgres-users.sh
```

The script is idempotent: it uses `DO $$ BEGIN … EXCEPTION WHEN duplicate_object` blocks so re-running it on an existing cluster only updates passwords without erroring.

**After running**, restart affected workloads if the passwords changed:

```bash
kubectl --context=altair -n atuin rollout restart deployment/atuin
kubectl --context=altair -n metering rollout restart deployment/metering-api
```

---

## loadtest-order-pipeline.js

k6 burst load test against order-pipeline's `order-service` `/orders` endpoint. Used to generate traffic for observing the Grafana dashboard (throughput, latency, consumer lag, retries, dead-letters) or to exercise chaos injection.

**Prerequisites:**

- [k6](https://k6.io/): `brew install k6`
- A port-forward to `order-service`: `kubectl --context=altair -n order-pipeline port-forward svc/order-service 8090:80`

**Usage:**

```bash
k6 run scripts/loadtest-order-pipeline.js
```

Defaults to 50 VUs for 60s against `http://localhost:8090`. Override with `-e VUS=<n>`, `-e DURATION=<duration>`, `-e BASE_URL=<url>`.
