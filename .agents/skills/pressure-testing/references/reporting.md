# Pressure Testing Reporting and Acceptance

## Always capture

- `runId`
- Docker profile used and replica counts for `visionflow` and `worker`
- user matrix, batch-size matrix, and concurrency matrix
- submit latency `p50/p95/p99`
- end-to-end latency `p50/p95/p99`
- planned requests, accepted requests, accepted rate
- planned images, success images, failed images, unresolved images
- `429` count, `503` count, throttle rate
- Redis memory and key queue stats
- BullMQ `waiting/active/failed/delayed`
- DB connection pressure and PgBouncer reachability or stats
- container CPU, memory, restart count
- route validation counts for `model/vendor`
- OSS cleanup results: matched objects, deleted, skipped, errors

## Artifact locations

Write artifacts under:

```text
logs/测试服并发性能日志/<runId>/
```

At minimum keep:

- one JSON report
- one Markdown summary
- any extra reconciliation or cleanup output produced by the harness

## Pass / fail rules

### Platform pass

- no platform-side queue `503` misreport
- no Prisma pool timeout
- no balance-freeze conflict snowball
- no workflow orphaning caused by local logic
- no Gemini image calls during `tietiezhi` batch tests
- no matting calls during `tietiezhi` batch tests
- no non-`tietiezhi` image route mismatch

### Platform fail

- local admission or queue logic fails before upstream becomes the limiter
- BullMQ or workflow stage logic fails under normal load
- run reports show route mismatch or forbidden model/vendor usage
- OSS cleanup deletes outside the current run

### Capacity ceiling, not correctness fail

- target IPM is missed
- higher stages are slower
- the harness stops because `stop-on-fail` hit a throughput threshold
- platform-side correctness signals stay green

Record these as "maximum stable stage reached X" and "recommended production ceiling Y".

### Infra or upstream blocker

- Docker API unavailable
- LAN PostgreSQL unavailable
- Redis host metrics port unreachable before the run starts
- `tietiezhi` upstream returns 5xx, timeouts, or unusable images such as 1x1 placeholders

Do not mix these with platform regressions.

## Recommendation format

State the result in this order:

1. highest stable scenario
2. why the next scenario failed or was stopped
3. recommended production ceiling, usually 60%-70% of the highest stable scenario
4. recommended replica count, worker concurrency, PgBouncer pool settings, and Redis memory budget

Example shape:

```text
最高稳定档：20 users × batch 50 × concurrency 100
停止原因：40 users × batch 50 × concurrency 250 开始出现上游 tietiezhi 5xx，平台侧指标仍稳定
推荐生产档：20 users × batch 50 × concurrency 60-70
建议配置：visionflow=2, worker=2, provider-call=128/worker, PgBouncer default_pool=50, Redis maxmemory=896MB
```
