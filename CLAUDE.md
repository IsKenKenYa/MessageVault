# CLAUDE.md

Claude Code should treat `AGENTS.md` as the canonical project operating spec.

## Skills

- `.agents/skills` is the project source of truth.
- `.claude/skills` is the Claude Code compatibility mirror.
- Do not edit `.claude/skills` directly. Update `.agents/skills`, then run:

```bash
bash scripts/sync-agent-skills.sh
```

## Quick Commands

```bash
bash scripts/check-repo-hygiene.sh
bash scripts/check-android-i18n.sh
bash scripts/sync-agent-skills.sh --check
```

For code work, use the verification commands listed in `AGENTS.md`.
