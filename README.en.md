# Commory

> Turn communication into memory.

Commory is a communication memory system that transforms SMS, call logs, contacts, and future communication sources into structured, queryable, AI-ready data.

This repository is in the first phase of a public-facing rename and positioning cleanup. The product brand is now `Commory`, and `MsgLayer` is the name of the underlying communication data layer.

## Positioning

Commory is not a backup-first tool.

- Traditional tools produce backup files
- Commory aims to produce structured data assets that can be searched, analyzed, and reused

In this repository today:

- `android/` is the ingestion layer
- `previewer/` is the viewer layer
- existing SDK modules are the foundation for the future `MsgLayer`

## Architecture

```text
Commory
├── Ingestion Layer
│   └── Android app
├── MsgLayer
│   ├── schema
│   ├── export models
│   └── future SDK / CLI interfaces
├── Viewer Layer
│   └── previewer
└── Future Extensions
    ├── CLI
    ├── plugins
    ├── local analysis
    └── agent integrations
```

## References

`references/` contains read-only external source references.

Current entry:

- `references/art-design-pro/`
  - source: `https://github.com/Daymychen/art-design-pro.git`
  - purpose: UI, interaction, and project-structure reference
  - status: read-only in this repository

## Naming Status

- Public brand: `Commory`
- Data layer name: `MsgLayer`
- Historical names such as `MessageVault` and `SMS Previewer` still exist in repository history and some component-level docs
- This phase does not rename package names, Gradle modules, or runtime code identifiers

## GitHub Description

```text
🧠 Commory · 通信记忆系统｜SMS/Call → Structured Data & AI｜Self-hosted · Privacy-first · Powered by MsgLayer
```

## License

The root repository remains under the [GNU General Public License v3.0](LICENSE).

External submodules under `references/` keep their own upstream licenses and ownership.
