# ADR 0001: Reference Governance and Clean-Room Learning

## Status

Accepted.

## Context

Commory benefits from studying strong Go, self-hosted, and agent systems. Some useful references use restrictive or unknown licenses, and some local snapshots may not be suitable as source material. We need a clear line between learning architecture and deriving implementation.

## Decision

All reference projects live under `references/` and are architecture inputs, not dependencies. Commory code must not import, copy, translate, or mechanically port code, strings, schemas, field names, or proprietary structures from reference projects.

Permissive references such as MIT, Apache-2.0, and BSD projects may be studied deeply, but implementation still has to be Commory-owned. AGPL, GPL, LGPL, proprietary, and unknown-license references are read-only design material. Their lessons must be restated as Commory ADRs, interfaces, tests, or design notes before implementation.

`references/ClaudeCodeSource` is clean-room only. It may inform product-level and architecture-level thinking about advanced agents, including task models, tool permissions, session continuity, bridge concepts, context compaction, command organization, and UX. It must not be used as a source for implementation details.

## Consequences

- `references/README.md` is the required catalog for reference purpose, license policy, learning targets, and prohibited borrowing.
- CI runs `scripts/check-reference-governance.sh` to catch uncataloged submodules and missing policy records.
- GitHubDaily is treated as a discovery index; every project found through it still needs separate license validation before being added.
- New architecture inspired by references should land as ADRs and Commory-owned interfaces before product integration.

