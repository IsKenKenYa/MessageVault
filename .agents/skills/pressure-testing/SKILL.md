---
name: pressure-testing
description: Standardize VisionFlow load, pressure, capacity, and soak testing across Docker, PgBouncer, BullMQ, Redis, workflow execution, and tietiezhi batch generation. Use when an agent needs to run or document high-concurrency validation, 2C4G simulations, multi-user submit tests, workflow stress tests, throughput investigations, OSS cleanup for test runs, or performance reporting and tuning guidance.
---

# Pressure Testing

Use this skill to run VisionFlow pressure tests in the repo's supported way instead of inventing an ad hoc flow.

## Pick the right harness

- Use `tests/harness/tietiezhi-concurrency-lab.ts` for `/api/generate/batch`, submit-plane, Redis/BullMQ/DB/PgBouncer, and `tietiezhi` mock upstream validation.
- Use `scripts/ops/workflow-1000-load.ts` for multi-user workflow execution, workflow cloning, and workflow node route validation.
- Use `scripts/ops/precheck-2c4g-load.ps1` before either harness when Docker, Redis, PgBouncer, or LAN PostgreSQL might be the blocker.

Read [`references/commands.md`](./references/commands.md) for exact commands. Read [`references/reporting.md`](./references/reporting.md) before declaring the run successful.

## Follow the standard workflow

1. Run the precheck script from PowerShell and confirm Docker API, Redis host port, PgBouncer host port, and LAN PostgreSQL are all reachable.
2. Start Docker with the repo-supported Compose stack. Use the base Compose plus `docker-compose.local.yml` for normal local runs. Add `docker-compose.low-resource.yml` only when simulating the temporary 2C4G profile.
3. Run a smoke stage first. Do not jump directly to the largest matrix.
4. Ramp through the matrix in ascending order. Keep `--stop-on-fail` enabled so the harness stops once it finds a real platform or routing failure.
5. Save the generated JSON and Markdown reports under `logs/测试服并发性能日志/<runId>/`.
6. Do a run-scoped OSS cleanup only after checking that the report matches the current run.
7. Write the final recommendation as "highest stable stage" and "recommended production ceiling", where the ceiling is usually 60%-70% of the highest stable stage.

## Enforce the repo rules

- Use Docker for pressure tests. Do not use `pnpm dev` as the main test surface.
- Use `--apply-config` for tietiezhi batch runs when validating platform capacity; it writes the 20-minute execution policy and unified `capacity.providerCallConcurrency` profile before publishing hot reload.
- The tietiezhi harness must not keep shorter local wait windows: submit tracker, generation tracker, and drain timeout default to the same 20-minute budget. If a run reports `generation_failed_force_timeout`, first verify the harness was run with the current defaults before treating it as a platform timeout.
- For local tietiezhi runs, `--prepare-local-users` raises test users' `featureBatchLimit` to at least `max(1000, capacity-provider-call, max batch)`, so product membership slots do not masquerade as platform capacity failures.
- Keep image stress tests on `model=tietiezhi` and `vendor=tietiezhi`. Treat Gemini image generation during these runs as a routing failure.
- Keep `--skip-matting` enabled for the tietiezhi harness. Matting tunnel noise is not part of the target benchmark.
- Treat host-side Redis metrics as `127.0.0.1:6379` unless `LOAD_REDIS_URL` or `--queue-redis-url` explicitly overrides it. Do not inherit stale `.env` values like `localhost:16379`.
- Scale with multiple containers and processes, not `worker_threads`. The supported pattern is `--scale visionflow=2 --scale worker=2` and then tuning worker concurrency.
- Keep OSS cleanup scoped to the current run. Never mass-delete historical test assets.
- If you touch docs while adjusting this workflow, update `README*.md`, `DEPLOY*.md`, and `CHANGELOG*.md` together.

## Classify failures correctly

- Treat these as platform failures:
  - queue `503` caused by local platform saturation or admission behavior
  - Prisma pool timeouts
  - balance freeze conflict snowballs
  - workflow orphaning or BullMQ stage failures caused by local logic
  - Gemini image calls, matting calls, or non-`tietiezhi` routing during a `tietiezhi` batch test
- Treat these as infra or upstream blockers unless the repo changed them:
  - Docker API or named pipe access issues
  - LAN PostgreSQL unreachability
  - Redis host port blocked before the test starts
  - `tietiezhi` returning bad images, 1x1 placeholders, or upstream 5xx/timeout noise
- If the harness stops because target IPM was missed but platform errors stayed at zero, record it as a capacity ceiling, not a correctness failure.

## Use the references deliberately

- Read [`references/commands.md`](./references/commands.md) when you need the exact precheck, startup, smoke, matrix, or cleanup commands.
- Read [`references/reporting.md`](./references/reporting.md) when you need acceptance criteria, report contents, failure taxonomy, or recommendation format.
- Read `docker/README.md` only when you need the Compose details behind the command matrix; do not re-derive the stack from scratch.
