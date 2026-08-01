# scorecard-mcp — MCP Design Audit

- **Date:** 2026-07-31
- **Scope:** the server as merged to `main` (`b00e197`) — tools, resources,
  transport, error handling, and safety.
- **Method:** the `mcp-server-dev` skill's rubric (`tool-design`,
  `server-capabilities`, `resources-and-prompts`), the Anthropic MCP Directory
  pass/fail review criteria, and live probes against the running binary over
  stdio.

## Verdict

**Built correctly.** The server follows the recommended MCP shape end to end:
one-tool-per-action design, Directory-conformant annotations, structured output
with a text fallback, responsible-AI framing in the server `instructions` (not
in tool descriptions), resources exposed as an index plus a URI template, and
tool-level errors that carry recovery hints. The audit found **one genuine
UX/correctness bug (Medium)** and a handful of **low/info polish items**.
Nothing structural needs rework.

## Rubric conformance

| Criterion (source) | Status | Evidence |
| --- | --- | --- |
| Deployment = local stdio, transport-agnostic factory | Pass | `server.New()` builds transport-free; entrypoint binds stdio |
| Tool pattern A (fewer than ~15 actions) | Pass | 5 tools — well under the context-budget ceiling |
| Directory: `readOnlyHint` + `destructiveHint` + `title` on every tool | Pass | verified via `tools/list` (all read-only, `destructiveHint:false`, titled) |
| Directory: tool names <= 64 chars | Pass | longest is `get_check_result` (16) |
| Directory: read/write split | Pass | all tools read-only by construction |
| Directory: no behavioral instructions in descriptions | Pass | descriptions disambiguate siblings, the recommended pattern — not the prohibited imperative kind |
| Directory: reference the upstream API in descriptions | Pass | each references the Scorecard REST API / docs |
| Structured output + text fallback | Pass | typed output auto-populates both; bad input rejected before the handler |
| Errors as tool errors with hints (no transport crash) | Pass (see F1) | error probes returned `isError` + message; server stayed up |
| `instructions` used for cross-cutting framing | Pass | responsible-AI framing lives here, not in tool text |
| Resources: index + RFC 6570 template, offline | Pass | `scorecard://checks` + `scorecard://checks/{name}` via `docs/checks` |
| No stdout pollution | Pass | `forbidigo` bans `fmt.Print*`; stdout is the protocol channel |
| Prompts / sampling / roots / elicitation | Correctly omitted | documented non-goals |
| Progress / cancellation | N/A now | cached reads are fast; `ctx` propagates; planned for the live provider |

## Findings

### F1 — Medium (real): `get_check_result` conflates "unknown check" with "known-but-absent check"

The tool validates the check name only against the *fetched result*, never
against the catalog it already ships. Two very different situations therefore
produce an identical, misleading error. Observed via live probes against
`ossf/scorecard`:

```text
get_check_result "SBOM"          -> isError: check not found in results: "SBOM" ... use list_checks ...
get_check_result "Totally-Bogus" -> isError: check not found in results: "Totally-Bogus" ... use list_checks ...
```

`SBOM` is a valid, *experimental* check that `list_checks` and `explain_check`
describe, so directing the user to `list_checks` is unhelpful — it is already
listed there. The same class of confusion affects `CI-Tests`, `Contributors`,
and `Dependency-Update-Tool` for repositories covered only by the weekly public
scan (they were *not* absent for `ossf/scorecard`, which self-publishes all
checks via the Scorecard Action).

**Fix:** cross-reference the catalog inside `get_check_result` and return three
distinct outcomes:

1. Unknown name — "not a Scorecard check; see `list_checks`".
2. Known but experimental / omitted from the cached scan — "valid check, but not
   present in these cached results; it may be experimental or omitted from the
   weekly scan (live scanning will cover it)".
3. Present — return the check result.

This is also the strongest argument for the live provider (`LocalRunProvider`),
which resolves outcome (2) at the source.

### F2 — Low: input constraints live in code, not in the JSON schema

Commit format and the `compare_repos` cap are enforced in handlers, but the
declared `inputSchema` does not express them — there is no `pattern` on
`commit` and no `maxItems` on `repos`. Expressing these in the schema lets the
host and model validate before a round-trip and self-correct.

### F3 — Low: `compare_repos` fetches sequentially

Up to ten serial HTTP calls means up to 10x latency. Cancellation is already
respected (`ctx` threads through each call); a bounded concurrent fetch would
cut wall-clock time materially.

### F4 — Low: `get_check_result` payload can be large

It returns the full `details[]`, which for checks such as `Pinned-Dependencies`
can be long. `get_repo_score` is correctly compact; consider a `details:false`
option or a truncation note here.

### F5 / F6 / F7 — Info

- **F5:** `Version` is still `dev`; wire it via `-ldflags` at release time.
- **F6:** No Prompts are offered (a documented non-goal). A
  `/scorecard-summary <repo>` prompt would be low-code, high-leverage later.
- **F7:** `goheader` is enabled but a no-op (no template configured), so license
  headers are not actually enforced. Not MCP-specific, but noted.

## What is exemplary

- **Annotation hygiene:** `openWorldHint` is `true` for the three network tools
  and `false` for the two offline catalog tools — a subtlety many servers miss.
- **Framing placement:** responsible-AI guidance is in `instructions` and the
  result `caveats[]`, never as imperative text in descriptions — the
  Directory-safe split.
- **Provenance on every result:** resolved commit SHA, scan date, Scorecard
  version, `source`, plus CDLA attribution — results are cacheable and
  reproducible (the darnit contract).
- **Provider seam:** `ResultProvider` keeps "live later" an addition, not a
  rewrite.
- **Determinism guardrail:** the `forbidigo` ban on `fmt.Print*` protects the
  stdio channel — a correctness property, not just style.

## Recommendation

Ship **F1** as the immediate correctness fix, and let it inform the live
provider (`LocalRunProvider`), which dissolves the underlying data gap. F2–F4
are quick polish. F5–F7 fold into the release and hygiene work.
