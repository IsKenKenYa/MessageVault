# ADR 0003：Agent Runtime 边界

## 状态

已接受。

## 背景

Commory 需要向高级 Agent 系统学习，但不能复制它们。ClaudeCodeSource、Hermes 和 OpenClaw 展示了 task state、tool permissions、provider registries、plugin boundaries、diagnostics 和 session continuity 等有价值的模式。

## 决策

Commory 的 Agent 工作从一个很小的 Go 自有边界开始：

`runtime -> provider adapter -> tool registry -> permission broker -> session store`

runtime 负责 orchestration。provider 负责 model I/O。tool 通过 Commory 自有 metadata 描述，并且只能经由 registry 调用。permission decision 必须显式且可测试。session 通过 interface 持久化，让 storage layer 可以独立演进。

核心 packages 不得依赖 plugin implementation。plugin 只能通过 SDK-style interface 进入系统。

## 后果

- Agent 代码从 `backend/internal/agent` 开始，包含 interface 和窄行为测试。
- 产品 handler 应调用 service 或 runtime，不直接引用 provider 或 plugin internals。
- 未来 plugin API 必须是 additive、documented、versioned，才能开放给第三方。
- 来自参考资料的高级概念必须翻译成 Commory 自有命名和行为。
