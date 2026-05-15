# Pressure Testing Command Matrix

## 1. Precheck

Run this from the repo root before any Docker-based pressure test:

```powershell
powershell -ExecutionPolicy Bypass -File scripts/ops/precheck-2c4g-load.ps1 -ComposeEnvFile docker/.env -PsqlPath "C:\Program Files\PostgreSQL\18\bin\psql.exe"
```

Expected checks:

- Docker API and current context
- merged low-resource Compose config
- host Redis port for metrics
- PgBouncer host port
- LAN PostgreSQL reachability through `psql`

## 2. Start Docker

### Standard local stack

```powershell
cd docker
docker compose --env-file .env -f docker-compose.yml -f docker-compose.local.yml up -d --build visionflow worker redis caddy pgbouncer
```

### Temporary 2C4G simulation

```powershell
cd docker
docker compose --env-file .env -f docker-compose.yml -f docker-compose.local.yml -f docker-compose.low-resource.yml up -d --build --scale visionflow=2 --scale worker=2 visionflow worker redis caddy pgbouncer
```

Use the low-resource overlay only for the temporary 2C4G simulation. It is not the main deployment entry.

## 3. Tietiezhi batch harness

### Smoke

```powershell
pnpm ops:tietiezhi-load -- --apply-config --users-list=10 --batch-size-list=30 --concurrency-list=10 --model=tietiezhi --vendor=tietiezhi --skip-matting --stop-on-fail --metrics-interval-ms=5000
```

### Main matrix

```powershell
pnpm ops:tietiezhi-load -- --apply-config --users-list=10,20,40 --batch-size-list=30,50,100 --concurrency-list=10,20,50,100,250,500,1000 --model=tietiezhi --vendor=tietiezhi --skip-matting --stop-on-fail --metrics-interval-ms=5000 --cleanup-oss
```

`--apply-config` writes the 20-minute execution policy and unified capacity profile. The harness also uses 20 minutes for submit/generation/drain tracking by default and raises local test users' batch/concurrency slots to at least `max(1000, capacity-provider-call, max batch)`.

### Safe cleanup pass

Run dry-run first if the last stage was exploratory or if object storage is shared:

```powershell
pnpm ops:tietiezhi-load -- --apply-config --users-list=10 --batch-size-list=30 --concurrency-list=10 --model=tietiezhi --vendor=tietiezhi --skip-matting --cleanup-oss --cleanup-dry-run=true
```

Switch `--cleanup-dry-run=false` only after confirming the run-scoped object list.

## 4. Workflow harness

### Smoke

```powershell
pnpm ops:workflow-load -- --low-resource-profile=2c4g --clone-workflows --concurrency-list=50 --requests=50 --users-limit=10 --user-pool-size=30 --apply-config
```

### Main matrix

```powershell
pnpm ops:workflow-load -- --low-resource-profile=2c4g --clone-workflows --concurrency-list=50,100,250,500,1000 --requests=1000 --users-limit=60 --user-pool-size=180 --apply-config --cleanup-oss --cleanup-dry-run=false
```

Use the default three workflow IDs unless the task explicitly requests a different set:

- `cmobdcmms000dyt64ajrz0x7q`
- `cmobdcjdd000byt64nwmntj5t`
- `cmobdcg710009yt640g3nsyk8`

The workflow harness already forces image-generation nodes to `tietiezhi/tietiezhi` and clones workflows per run when `--clone-workflows` is enabled.

## 5. Required static checks when the load tooling changes

Run these from the repo root when the harness, Docker stack, or pressure-test routing changes:

```powershell
pnpm typecheck
pnpm lint
RUN_DB_TESTS=1 pnpm exec vitest run src/lib/__tests__/balance-concurrency.test.ts --reporter=dot
```

Run `pnpm swagger:sync` only if the change also touched API behavior or Swagger annotations.
