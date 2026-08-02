# OpenSSF Baseline — Process Record

This document records how `scorecard-mcp` was assessed against the
[OpenSSF Baseline](https://baseline.openssf.org) (OSPS Baseline), what the
assessment found, and — as importantly — _why_ the remaining gaps are handled the
way they are. It is a companion to [`init-impl.md`](init-impl.md) (how the server
was built) and [`mcp-design-audit.md`](mcp-design-audit.md) (the MCP/Directory
review), written so a future contributor — including whoever prepares the
upstreaming PR — can reconstruct the reasoning rather than re-derive it.

## Overview

The [OpenSSF Baseline](https://baseline.openssf.org) is a tiered catalog of
security-hygiene controls for open source projects, organized into three maturity
levels (L1 → L3). Controls carry IDs of the form `OSPS-<domain>-<n>.<m>` across
domains such as Access Control (`AC`), Build & Release (`BR`), Documentation
(`DO`), Governance (`GV`), Legal (`LE`), Quality Assurance (`QA`), Security
Assessment (`SA`), and Vulnerability Management (`VM`).

The assessment was run with [`darnit`](https://github.com/darnitdevorg/darnit),
the OpenSSF Baseline tool, driven over its MCP server. `darnit` is also the
downstream consumer of Scorecard output described in
[`init-impl.md`](init-impl.md#integration-darnit), so exercising it here doubles
as dogfooding the integration this project is meant to serve.

## Method

The assessment followed `darnit`'s intended flow:

1. **Audit** at Level 3 (the strictest, so every applicable control is surfaced).
2. **Confirm project context** — answer the questions that decide which controls
   apply and how, so results are not skewed by wrong defaults. The confirmed
   values are persisted to [`.project/darnit.yaml`](../.project/darnit.yaml).
3. **Re-audit** with context applied, producing the numbers below.
4. **Decide remediation** deliberately per control group, rather than
   bulk-applying generated files.

Repository settings were inspected read-only via the GitHub API to determine
what is actually achievable (see [Blocked](#whats-blocked-repository-settings-5)).

## Confirmed project context

Recorded in [`.project/darnit.yaml`](../.project/darnit.yaml):

| Key                   | Value             | Notes                              |
| --------------------- | ----------------- | ---------------------------------- |
| `governance_model`    | `meritocracy`     | Judgment call — see below          |
| `has_subprojects`     | `false`           | Standalone single repo             |
| `is_library`          | `false`           | Judgment call — see below          |
| `has_releases`        | `true`            | Releases planned during incubation |
| `has_compiled_assets` | `true`            | Ships a Go binary                  |
| `ci_provider`         | `github`          | GitHub Actions                     |
| `primary_language`    | `go`              |                                    |
| `security_contact`    | `See SECURITY.md` |                                    |

Two values were judgment calls worth recording:

- **`governance_model: meritocracy`.** Today the project is effectively
  single-maintainer, which reads as `bdfl`. But the stated endgame
  ([`init-impl.md`](init-impl.md#placement-incubate-standalone-upstream-in-tree))
  is upstreaming into `ossf/scorecard` as an in-tree `scorecard mcp` subcommand.
  OpenSSF/Scorecard governance is meritocratic — authority earned through
  contribution — so recording `meritocracy` matches the intended home and
  trajectory rather than the transient solo state.
- **`is_library: false`.** This is an application, not a library. The only
  `package main` is `cmd/scorecard-mcp` (it builds the MCP server binary), all
  logic lives under `internal/` (which Go forbids other modules from importing),
  and the endgame is a cobra subcommand — a command, not an importable API. It
  _consumes_ the Scorecard library; it is not consumed as one.

Beyond the audit context, the `.project/` metadata fields were completed this
session: `description` and `repositories` in
[`.project/project.yaml`](../.project/project.yaml), and `slug`, `project_lead`,
and `package_managers` (`gomod: go.mod`) in
[`.project/darnit.yaml`](../.project/darnit.yaml). The remaining empty fields
(`maturity_log`, `social`, `mailing_lists`, `audits`, `cncf_slack_channel`,
`controls`) are CNCF-lifecycle or per-control-override fields that do not apply
and are intentionally left blank.

## Results at a glance

Level 3 audit, with context applied:

| Status                       | Count |
| ---------------------------- | ----: |
| ✅ Pass                      |    42 |
| ❌ Fail                      |    16 |
| ⚠️ Needs manual verification |     5 |
| ➖ Not applicable            |     2 |
| **Total**                    |    65 |

Level compliance (all currently non-compliant):

| Level | Failed | Unverified |
| ----- | -----: | ---------: |
| L1    |      2 |          2 |
| L2    |      4 |          1 |
| L3    |     10 |          2 |

The foundational hygiene that _travels with the code_ is already in place; the
gaps are concentrated in repository settings (blocked by plan/visibility) and in
documentation that would be redundant post-upstream.

## What passes today (42)

Satisfied controls span, by theme:

- **Legal & licensing:** `LE-01.01`, `LE-02.01`, `LE-02.02`, `LE-03.01`,
  `LE-03.02` — Apache-2.0 `LICENSE`, license metadata.
- **Governance & community health:** `GV-01.01`, `GV-01.02`, `GV-02.01`,
  `GV-03.01`, `GV-03.02`, `GV-04.01` — `MAINTAINERS.md`, `CONTRIBUTING.md`,
  Code of Conduct, `.github/CODEOWNERS`.
- **Documentation:** `DO-01.01`, `DO-02.01`, `DO-03.01`, `DO-04.01`, `DO-06.01`,
  `DO-07.01` — `README.md`, `SUPPORT.md`, usage/build docs.
- **Access control (repo hygiene checkable without settings):** `AC-01.01`,
  `AC-02.01`, `AC-04.01`, `AC-04.02`.
- **Quality assurance:** `QA-01.02`, `QA-02.01`, `QA-05.01`, `QA-05.02`,
  `QA-06.01`, `QA-06.03` — CI present, tests, public build, dependency pinning.
- **Vulnerability management:** `VM-01.01`, `VM-02.01`, `VM-04.01`, `VM-05.02`,
  `VM-05.03`, `VM-06.02` — `SECURITY.md`, Dependabot, CodeQL/SAST wiring.
- **Build & release:** `BR-01.01`, `BR-01.02`, `BR-03.01`, `BR-03.02`,
  `BR-04.01`, `BR-05.01`, `BR-06.01`, `BR-07.01` — pinned actions, hardened
  workflows, `CHANGELOG.md`, `go.mod`, `.gitignore`.
- **Security assessment:** `SA-02.01`.

## What's blocked: repository settings (5)

These five controls require GitHub repository-settings changes that are
**unavailable on this repository** and cannot be satisfied by any file we add:

| Control         | Level | Requires                                    |
| --------------- | ----- | ------------------------------------------- |
| `OSPS-AC-03.01` | L1    | Branch protection: require PRs to `main`    |
| `OSPS-AC-03.02` | L1    | Branch protection: block deletion of `main` |
| `OSPS-QA-03.01` | L2    | Branch protection: required status checks   |
| `OSPS-QA-07.01` | L3    | Branch protection: required review approval |
| `OSPS-VM-03.01` | L2    | Private vulnerability reporting (PVR)       |

**Why blocked.** `uwu-tools/scorecard-mcp` is a **private** repository owned by
the `uwu-tools` organization on a **free** plan. Both the modern
**rulesets** API and **classic branch protection** return
`HTTP 403 — "Upgrade to GitHub Pro or make this repository public to enable this
feature."` So there is no fallback mechanism for the four branch-protection
controls while the repo is private on the free plan.

**Two independent constraints keep these unfixable for now:**

- The repository must remain **private** (organizational policy). Making it
  public — which would enable rulesets, branch protection, _and_ PVR for free —
  is not an option at this time.
- The `uwu-tools` org **will not be upgraded** to a paid plan (GitHub Team) at
  this time. (Even if it were, `VM-03.01` would remain out of reach: **Private
  vulnerability reporting is a public-repository-only feature**, so no paid tier
  enables it on a private repo.)

**Disposition:** deferred to upstream. When the code lands in `ossf/scorecard` —
a public OpenSSF org repository — it inherits that org's branch protection,
required reviews, and private vulnerability reporting structurally, satisfying
all five without further work here.

## What's deferred to upstream: documentation & policy (11)

`darnit` can auto-generate files that satisfy these controls, but we chose **not**
to create them in this repository:

| Control         | Level | Would document                            |
| --------------- | ----- | ----------------------------------------- |
| `OSPS-SA-01.01` | L2    | Design/architecture (actions & actors)    |
| `OSPS-SA-03.01` | L2    | Security assessment before major releases |
| `OSPS-SA-03.02` | L3    | Threat model                              |
| `OSPS-QA-06.02` | L3    | Testing instructions                      |
| `OSPS-QA-02.02` | L3    | SBOM delivered with compiled assets       |
| `OSPS-BR-07.02` | L3    | Secrets/credentials management policy     |
| `OSPS-DO-03.02` | L3    | How to verify release-author identity     |
| `OSPS-DO-05.01` | L3    | End-of-support (EOL) policy               |
| `OSPS-VM-04.02` | L3    | VEX policy (in `SECURITY.md`)             |
| `OSPS-VM-05.01` | L3    | SCA (dependency) remediation policy       |
| `OSPS-VM-06.01` | L3    | SAST remediation policy                   |

**Why deferred.** Much of this documentation — architecture, threat model,
security-assessment cadence, EOL and secrets policy, and the VEX/SCA/SAST
remediation policies that would live in `SECURITY.md` — is properly owned by the
_hosting project_ once upstreamed. Authoring standalone versions during
incubation risks producing throwaway artifacts that diverge from, or contradict,
`ossf/scorecard`'s own policies. This mirrors the deferral logic already applied
to the live provider and release automation in
[`init-impl.md`](init-impl.md#outcomes-and-next-steps).

**Two nuances to revisit if incubation lengthens:**

- `SBOM` (`QA-02.02`) is not a doc but a _build output_. Because this repo does
  intend to ship a compiled Go binary during incubation (`has_compiled_assets:
true`), if standalone releases begin before upstreaming, SBOM generation
  should be wired into the release workflow at that point.
- The VEX/SCA/SAST policies (`VM-04.02`, `VM-05.01`, `VM-06.01`) are small
  additions to the existing `SECURITY.md`. If a self-contained baseline posture
  is ever wanted before upstream, these are the cheapest wins and touch only one
  file.

## Needs manual verification (5)

These could not be determined automatically and relate to build/release
integrity and provenance — capabilities this project has **not** implemented yet:

| Control         | Level | Concerns                             |
| --------------- | ----- | ------------------------------------ |
| `OSPS-QA-01.01` | L1    | Public CI/build configuration        |
| `OSPS-BR-01.03` | L1    | Build/release pipeline integrity     |
| `OSPS-BR-01.04` | L3    | Build/release pipeline integrity     |
| `OSPS-BR-02.01` | L2    | Release provenance / signed releases |
| `OSPS-BR-02.02` | L3    | Release provenance / signed releases |

**Disposition:** these are satisfied naturally once signed, provenance-emitting
release automation exists — itself a deferred item in
[`init-impl.md`](init-impl.md#outcomes-and-next-steps). Treated as **open**
until release automation lands (here or upstream).

## Not applicable (2)

- `OSPS-QA-04.01` (L1) and `OSPS-QA-04.02` (L3) — subproject coordination.
  N/A because `has_subprojects: false`.

## Planned configuration (for when unblocked)

Recorded so it can be applied verbatim the moment the repository is public (or
the org is on a paid tier for the four branch-protection controls). Note that
`darnit`'s `enable_branch_protection` tool uses **classic** branch protection;
the plan below deliberately uses a **repository ruleset** instead (the modern,
layerable mechanism), so it would be applied via the GitHub API/UI rather than
that tool.

**Ruleset on `main` (solo-maintainer-friendly):**

- **Rules:** require a pull request with **1 approving review**; block
  **deletion**; block **force-pushes** (non-fast-forward); require **status
  checks** to pass.
- **Required status checks** (exact contexts, mapped from live check runs):
  `build-test`, `lint`, `super-linter`, `dependency-review`, `Run zizmor`.
  Deliberately excluded as "core CI" only: `Analyze (go)` / `Analyze (actions)`
  (CodeQL) and `Scorecard analysis`, which are slower/scheduled.
- **Bypass actor:** Repository `admin` role — so the sole maintainer can still
  merge while the rule applies to everyone else.
- **Enforcement:** active.

**Private vulnerability reporting:** enable via the GitHub API endpoint
`PUT /repos/{owner}/{repo}/private-vulnerability-reporting` — applicable only
after the repo is public.

## Remaining work / next steps

- [ ] **At upstream time:** confirm `ossf/scorecard`'s org-level settings cover
      `AC-03.01`, `AC-03.02`, `QA-03.01`, `QA-07.01`, and `VM-03.01`.
- [ ] **At upstream time:** rely on the hosting project's architecture, threat
      model, security-assessment, EOL, secrets, and VEX/SCA/SAST policies for the
      11 deferred documentation controls; delete any that this repo should not
      own.
- [ ] **When release automation lands:** add SBOM generation (`QA-02.02`) and
      signed provenance to satisfy the build/release-integrity controls
      (`QA-01.01`, `BR-01.03`, `BR-01.04`, `BR-02.01`, `BR-02.02`).
- [ ] **If incubation lengthens and a standalone posture is wanted:** apply the
      ruleset + PVR above (requires public or paid), and add the VEX/SCA/SAST
      sections to `SECURITY.md`.
- [ ] **Re-audit** after any of the above: `darnit` audit at Level 3 against this
      repo, refreshing this record.

## Tooling notes and darnit feedback

This assessment was the project's first end-to-end exercise of `darnit`, so a few
observations are worth capturing — both for whoever runs it here next and as
candidate issues to file upstream against
[`darnit`](https://github.com/darnitdevorg/darnit). (Exercising `darnit` here also
dogfoods the consumer described in
[`init-impl.md`](init-impl.md#integration-darnit).)

### Operational notes

- **Canonical repo vs. local remote.** The audit identifies the repository as
  `uwu-tools/scorecard-mcp` (its canonical home), while this working clone's
  `origin` is a personal fork (`justaugustus/scorecard-mcp`). Repo-scoped checks
  (branch protection, PVR) run against the canonical repo, not the fork.
- **Exact CI check contexts** (mapped from live check runs on `main`, for anyone
  configuring required status checks): `build-test`, `lint`, `super-linter`,
  `dependency-review`, `Run zizmor`, `Analyze (go)`, `Analyze (actions)`,
  `Scorecard analysis`, `update-go_modules-graph`. Contexts are the job _names_
  where a job sets one (e.g. `Run zizmor`, not `zizmor`), which matters for exact
  matching in a ruleset.
- **`get_project_config` can read stale.** After editing `.project/*.yaml`
  directly (for fields `confirm_project_data` cannot write), `get_project_config`
  kept returning pre-edit values within the session; the audit itself reads files
  fresh. Re-read the files on disk to confirm state.

### Candidate issues to file against `darnit`

1. **Support GitHub rulesets, not only classic branch protection.**
   `enable_branch_protection` targets
   `PUT /repos/{owner}/{repo}/branches/{branch}/protection` (classic). Rulesets
   are GitHub's current, layerable mechanism and are what this project chose. A
   ruleset-based remediation (or an option to select it) would match modern
   practice.
2. **Detect plan/visibility gating before recommending settings remediation.**
   On a private repo on a free plan, both rulesets and classic protection return
   `HTTP 403 ("Upgrade to Pro or make public")`. `darnit` currently reports
   `AC-03.01`, `AC-03.02`, `QA-03.01`, and `QA-07.01` as plain failures with
   remediation steps that will 403. Pre-checking visibility/plan and annotating
   "blocked — requires public or paid plan" would avoid dead-end remediation.
3. **`VM-03.01` remediation guidance is misleading for private repos.** The note
   offers "provide a security email in `SECURITY.md`" as an alternative, but the
   check is API-based and Private Vulnerability Reporting is a public-repo-only
   feature. On a private repo the control can never pass via the setting, and a
   documented `SECURITY.md` contact is not accepted. Either honor a documented
   contact when PVR is unavailable, or state the public-only constraint in the
   note.
4. **Solo-maintainer footgun in `enable_branch_protection` defaults.** The
   defaults (`required_approvals=1`, `enforce_admins=true`) on a single-maintainer
   repo lock the sole maintainer out of merging — no second approver exists and
   admins cannot bypass. A warning, or a solo-friendly preset with an admin bypass
   actor, would help.
5. **No supported writer for non-context `.project` fields.**
   `confirm_project_data` only sets `x_openssf_baseline.context` keys; the
   top-level fields (`description`, `slug`, `repositories`, `package_managers`)
   have no MCP writer, yet the tooling warns against hand-editing `.project/`.
   The generated files also do not document the expected shape for `repositories`
   or `package_managers`, so their structure has to be guessed.
6. **Unhelpful option descriptions in the context wizard.** The governance/enum
   questions surface placeholder descriptions (e.g. "Select 'bdfl'"); short
   explanations of each model would help users answer correctly.

## Conventions

- Project context lives in [`.project/darnit.yaml`](../.project/darnit.yaml);
  update it there (via `darnit`) rather than by hand so audits stay accurate.
- This record is a companion to [`init-impl.md`](init-impl.md) and
  [`mcp-design-audit.md`](mcp-design-audit.md); keep the three consistent when
  deferred items move.
