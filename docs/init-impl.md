# Initial Implementation and Scaffolding — Process Record

This document records how `scorecard-mcp` was designed, built, and scaffolded,
and — more importantly — *why* each decision was made. It is a narrative
companion to the specifications under `openspec/` and the review in
[`mcp-design-audit.md`](mcp-design-audit.md), written so future contributors can
reconstruct the reasoning rather than reverse-engineer it from the diff.

## Overview

`scorecard-mcp` is a [Model Context Protocol](https://modelcontextprotocol.io)
(MCP) server for [OpenSSF Scorecard](https://github.com/ossf/scorecard), written
in Go. It was built with three tools working together:

- **Spec-driven development** via [OpenSpec](https://openspec.dev) — every
  behavioral change starts as a proposal, design, and specification before code.
- **MCP development Agent Skills** (`mcp-server-dev`: `build-mcp-server` and
  friends) — used to interrogate the use case, pick the deployment/tool-design
  patterns, and audit the result against MCP and Anthropic Directory criteria.
- **Reference implementations** — the existing Python
  [`steiza/scorecard-mcp`](https://github.com/steiza/scorecard-mcp) server (for
  patterns to inherit and improve on) and [`ossf/allstar`](https://github.com/ossf/allstar)
  (for Go linting and CI conventions).

The work was intentionally **decision-first and iterative**: clarifying
questions and short written rationales preceded each build phase, and each
phase was verified before the next began.

## Key decisions and rationale

### Language: Go

MCP is a protocol boundary, so client/server interoperability is
language-agnostic — the choice comes down to ecosystem and maintenance rather
than interop. For Scorecard that points to Go:

- Scorecard, its Action, and its webapp are all Go; a Go server is one toolchain
  for the same maintainers to own.
- Go lets the server import the Scorecard library directly
  (`github.com/ossf/scorecard/v5`), which unlocks in-process scanning later (see
  scope) and reuse of the check documentation now.
- The official `github.com/modelcontextprotocol/go-sdk` provides first-class
  support.

### Placement: incubate standalone, upstream in-tree

The project is incubated as a standalone Go module that imports the Scorecard
library, with the intended endgame of contributing it as an in-tree
`scorecard mcp` cobra subcommand in `ossf/scorecard` — analogous to the existing
in-tree `scorecard serve` HTTP interface. Standalone incubation allows fast
iteration and lets community consensus form in parallel; because the code
already imports `pkg` and is organized under `cmd/` and `internal/`, upstreaming
is a move plus an `AddCommand` wiring rather than a rewrite.

### Scope: cached now, live later

The server ships reading **pre-computed results from the public Scorecard REST
API** (`api.scorecard.dev`). On-demand, in-process scanning via
`pkg/scorecard.Run` (private repos, the full check set, fresh results) is
designed into the architecture but deferred to a later change. This delivers
value immediately while keeping the larger capability a non-rewrite addition.

### Architecture: a provider seam and a transport-agnostic factory

- A `ResultProvider` interface abstracts where results come from. `main` wires a
  `CachedRESTProvider` today; a `LocalRunProvider` slots in later behind the same
  interface, with every result declaring its `source`.
- The MCP server is built by a single transport-agnostic factory
  (`server.New`). It runs over stdio now; Streamable HTTP can be added as a
  transport binding without touching any tool or resource handler.

### Integration: darnit

The [`darnit`](https://github.com/darnitdevorg/darnit) OpenSSF Baseline tool
(issue #194) consumes Scorecard output as a cached "heavy tool" reused across
many controls. The integration boundary is **JSON, not language**. Accordingly,
every result carries provenance (resolved commit SHA, scan date, Scorecard
version, `source`) so downstream consumers can cache and reproduce it, and the
output mirrors Scorecard's canonical JSON so a consumer's adapter is drop-in.

## Phase 1 — Discovery

Discovery combined the `build-mcp-server` skill's interrogation with direct
investigation of the upstream projects:

- **MCP shape:** local stdio deployment (correct for a CLI-embedded tool);
  tool-design pattern A (one tool per action, a handful of tools); the official
  Go SDK; plain structured JSON (no UI widgets or elicitation).
- **Clarified explicitly** (via targeted questions): conform to Anthropic MCP
  Directory review criteria from the start; GitHub + GitLab platforms; compact
  results by default with drill-down; Prompts deferred.
- **Scorecard internals** (investigated across the library and REST API):
  `pkg/scorecard.Run` is importable out-of-tree; the check catalog is available
  offline via the public `docs/checks` package; the cached REST result is a
  subset of the CLI result (score + checks + metadata; no probe findings or
  maintainer annotations); a score of `-1` means *inconclusive*, not failing;
  authentication for live scans is handled by the library's roundtripper; the
  REST data is licensed CDLA Permissive 2.0.
- **darnit #194:** confirmed the JSON-boundary contract described above.

## Phase 2 — Specification (OpenSpec)

The change `add-scorecard-mcp-server` was authored through OpenSpec's
spec-driven workflow, with review gates between artifacts:

1. **Proposal** — why, what changes, and the capabilities introduced.
2. **Design** — the `ResultProvider` seam, the provenance/determinism contract,
   the tool surface, transport structuring, and the upstreaming path.
3. **Specs** — three capabilities, each with requirements and WHEN/THEN
   scenarios:
   - `mcp-server` — runtime, transport-agnostic factory, reference parsing,
     structured output, error handling, responsible-AI framing, and
     Directory-conformant tool annotations and input constraints.
   - `scorecard-results` — retrieval, comparison, the provider seam, provenance,
     commit pinning, caveats, score semantics, and attribution.
   - `check-catalog` — list/explain tools, documentation resources, and catalog
     version consistency.
4. **Tasks** — an ordered, verifiable implementation checklist.

`openspec validate --strict` was kept green throughout. Project context and the
requirement to use the MCP dev skills were recorded in `openspec/config.yaml`
and `AGENTS.md` so future agent sessions inherit them.

## Phase 3 — Implementation

Implementation proceeded in verified increments:

1. **Core + results** — `internal/model` (result types with provenance,
   caveats, attribution), `internal/scorecardref` (reference parsing and commit
   validation), `internal/provider` (the `ResultProvider` seam and
   `CachedRESTProvider`), `internal/server` (the factory), the three
   `scorecard-results` tools, and the `cmd/scorecard-mcp` stdio entrypoint.
   Verified end to end over stdio against the live API.
2. **Check catalog** — `internal/catalog` (offline via `docs/checks`), the
   `list_checks` and `explain_check` tools, and the `scorecard://checks` and
   `scorecard://checks/{name}` resources.
3. **Tests** — provider unit tests (via `httptest`) and in-memory MCP
   integration tests exercising the tools and resources end to end.

The resulting surface is five read-only tools (`get_repo_score`,
`get_check_result`, `compare_repos`, `list_checks`, `explain_check`) plus two
resources, all returning structured output with a text fallback, Directory-
conformant annotations, and responsible-AI framing carried in the server
`instructions` (never in tool descriptions).

**Toolchain note:** the environment had a `GOROOT` pointing at a different Go
version than the `go` binary on `PATH`, causing standard-library version
mismatches. Go commands were run with the stray `GOROOT` unset
(`env -u GOROOT go ...`) so the toolchain and its standard library matched
(Go 1.25.x, aligning with Scorecard).

## Phase 4 — Quality gates

- **golangci-lint:** the configuration was adapted from `ossf/allstar`, and all
  of the linters that upstream leaves commented out were enabled and satisfied
  (sentinel errors for `err113`, line-length and signature cleanups for `lll`,
  parallel tests, wrapping boundaries for `wrapcheck`, and so on). Reports zero
  issues.
- **zizmor:** a GitHub Actions security-analysis workflow, verified to report no
  findings across all workflows.
- **CI:** a Go build/vet/test workflow plus super-linter, CodeQL, and
  dependency review.

## Phase 5 — Repository scaffolding

Standard open-source repository structure was added, adapted to this project:

- **Community-health files:** Code of Conduct (Contributor Covenant 2.1),
  CONTRIBUTING (with a DCO sign-off requirement), SECURITY (GitHub private
  vulnerability reporting), SUPPORT, MAINTAINERS, and a Keep-a-Changelog
  CHANGELOG.
- **`.github/`:** CODEOWNERS, issue and pull-request templates, dependabot
  (Go modules and Actions), and super-linter configuration.
- **Workflows:** all with pinned action SHAs, `step-security/harden-runner`, and
  `permissions: {}` defaults with least-privilege per-job grants — CI, lint,
  CodeQL, dependency review, Scorecard (via the public `ossf/scorecard-action`),
  stale (via the public `actions/stale`), and zizmor.
- **`.project/`:** CNCF `project.yaml` and a `darnit.yaml` extension.
- **README** badges and a Community section, and a hardened `.gitignore`.

## Conventions

- **Git:** work on feature branches (never commit directly to `main`); every
  commit is DCO signed-off (`git commit -s`); changes are pushed and reviewed,
  and pull requests are opened only on explicit request.
- **Licensing:** Apache-2.0 for code; cached Scorecard data surfaced by the
  server is attributed as CDLA Permissive 2.0.
- **Tooling:** follow the OpenSpec workflow for non-trivial changes and use the
  MCP dev Agent Skills when designing or implementing the server.

## Outcomes and next steps

The initial implementation and scaffolding were reviewed against the MCP rubric
and Directory criteria; the results, findings, and a recommended sequence are in
[`mcp-design-audit.md`](mcp-design-audit.md). The main deferred items are the
live in-process provider (`LocalRunProvider`), a Streamable HTTP transport,
release automation, and eventual upstreaming as an in-tree `scorecard mcp`
subcommand.
