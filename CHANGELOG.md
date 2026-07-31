# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic
Versioning](https://semver.org/spec/v2.0.0.html).

<!-- textlint-disable -->

## [Unreleased]

### Added

* MCP server for OpenSSF Scorecard, written in Go using the official
  `modelcontextprotocol/go-sdk`, communicating over stdio.
* A `ResultProvider` seam with a `CachedRESTProvider` backed by the public
  Scorecard REST API (`api.scorecard.dev`).
* Tools: `get_repo_score`, `get_check_result`, `compare_repos`, `list_checks`,
  and `explain_check`.
* Read-only resources: `scorecard://checks` and `scorecard://checks/{name}`.
* Structured output with provenance (resolved commit SHA, scan date, Scorecard
  version, source), cached-data caveats, and CDLA Permissive 2.0 attribution.
* Project scaffolding: OpenSpec specifications, community-health files, GitHub
  templates and workflows, and an allstar-derived golangci-lint configuration.

[unreleased]: https://github.com/uwu-tools/scorecard-mcp/commits/main

<!-- textlint-enable -->
