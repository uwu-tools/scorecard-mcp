# AGENTS.md — scorecard-mcp

Guidance for AI agents working in this repository. Humans: this doubles as a
contributor quickstart.

## What this is

An MCP (Model Context Protocol) server for **OpenSSF Scorecard**, written in
**Go**. It exposes Scorecard security-posture data to MCP clients. Incubated
here as a standalone Go module that imports `github.com/ossf/scorecard/v5/pkg`;
the intended endgame is upstreaming as an in-tree `scorecard mcp` cobra
subcommand in `github.com/ossf/scorecard` (analogous to `scorecard serve`).

## Spec-driven development (OpenSpec)

This project uses **OpenSpec** for spec-driven development. Before writing code:

1. Read the active change(s) under `openspec/changes/` and the project context in
   `openspec/config.yaml`.
2. Follow the workflow: **explore → propose → design → specs → tasks →
   implement**. Do not restructure without a spec.
3. Useful commands: `openspec list`, `openspec status --change <name>`,
   `openspec validate <name> --strict`.

## Use the MCP dev Agent Skills

When designing or implementing the MCP server, **use the `mcp-server-dev` Agent
Skills** — they encode MCP design patterns, tool-design rules, capabilities, and
transport choices:

- `build-mcp-server` — entry point (discovery, deployment model, tool patterns)
- `build-mcp-app` — interactive UI widgets (out of scope for this project)
- `build-mcpb` — local bundling (out of scope; we ship a Go binary)

Reference: https://modelcontextprotocol.io/docs/2026-07-28/develop/build-with-agent-skills

**Install:** `/plugin marketplace add anthropics/claude-plugins-official` then
`/plugin install mcp-server-dev`. If the plugin marketplace is blocked by
enterprise policy, fall back to cloning each skill's `SKILL.md` + `references/`
from `anthropics/claude-plugins-official` under
`plugins/mcp-server-dev/skills/*` into your skills directory
(`~/.claude/skills/` for user-level, or `.claude/skills/` for project-level).

## Conventions

- **Language/stack:** Go; official `github.com/modelcontextprotocol/go-sdk`.
  Match the Scorecard toolchain's Go version to avoid CI friction.
- **Architecture:** a `ResultProvider` seam abstracts the backend
  (`CachedRESTProvider` now → `LocalRunProvider` via `pkg/scorecard.Run` later).
- **Tool design:** read-only tools with `readOnlyHint`/`title` annotations,
  contract-style descriptions, structured output (`outputSchema` +
  `structuredContent` + text fallback). Responsible-AI framing and cached-data
  caveats go in the server `instructions` field and result payloads — **never**
  as imperative instructions inside tool descriptions.
- **Commits:** Apache-2.0; DCO sign-off (`git commit -s`); AI co-authorship
  trailer `Co-Authored-By: Claude <noreply@anthropic.com>`. Never commit to
  `main` — use feature branches. Don't create PRs unless asked.
- **Upstream target conventions** (`ossf/scorecard`): minimal dependency policy
  — isolate the MCP SDK under `cmd/`.
