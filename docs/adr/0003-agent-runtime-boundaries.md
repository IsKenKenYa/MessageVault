# ADR 0003: Agent Runtime Boundaries

## Status

Accepted.

## Context

Commory should learn from advanced agent systems without copying them. ClaudeCodeSource, Hermes, and OpenClaw show useful patterns around task state, tool permissions, provider registries, plugin boundaries, diagnostics, and session continuity.

## Decision

Commory's agent work starts with a small Go-owned boundary:

`runtime -> provider adapter -> tool registry -> permission broker -> session store`

The runtime owns orchestration. Providers own model I/O. Tools are described through Commory-owned metadata and invoked only through a registry. Permission decisions are explicit and testable. Sessions are persisted through an interface so the storage layer can evolve independently.

Core packages must not depend on plugin implementations. Plugins can enter through SDK-style interfaces only.

## Consequences

- Agent code starts in `backend/internal/agent` with interfaces and narrow behavior tests.
- Product handlers should call services or runtimes, never reference provider or plugin internals directly.
- Any future plugin API must be additive, documented, and versioned before third-party use.
- Advanced concepts from references must be translated into Commory naming and behavior.

