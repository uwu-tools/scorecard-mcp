# How to Contribute

Thanks for your interest in contributing to `scorecard-mcp`! Here are a few
general guidelines on contributing and reporting bugs that we ask you to review.
Following these guidelines helps to communicate that you respect the time of the
contributors managing and developing this open source project. In return, they
should reciprocate that respect in addressing your issue, assessing changes, and
helping you finalize your pull requests.

Please note that all of your interactions in the project are subject to our
[Code of Conduct](/CODE_OF_CONDUCT.md). This includes creation of issues or pull
requests, commenting on issues or pull requests, and extends to all interactions
in any real-time space e.g., Slack, Discord, etc.

## Reporting Issues

Before reporting a new issue, please ensure that the issue was not already
reported or fixed by searching through our [issues
list](https://github.com/uwu-tools/scorecard-mcp/issues).

When creating a new issue, please be sure to include a **title and clear
description**, as much relevant information as possible, and, if possible, a
test case.

**If you discover a security bug, please do not report it through GitHub issues.
Instead, please see security procedures in [SECURITY.md](/SECURITY.md).**

## Sending Pull Requests

Before sending a new pull request, take a look at existing pull requests and
issues to see if the proposed change or fix has been discussed in the past, or
if the change was already implemented but not yet released.

We expect new pull requests to include tests for any affected behavior, and, as
we follow semantic versioning, we may reserve breaking changes until the next
major version release.

### Developer Certificate of Origin (DCO)

This project requires a [Developer Certificate of Origin][dco] (DCO) sign-off on
every commit. Add one automatically with the `-s` flag:

```sh
git commit -s -m "your message"
```

This appends a `Signed-off-by: Your Name <your@email>` trailer certifying that
you wrote or have the right to submit the change.

[dco]: https://developercertificate.org/

### Development

- The server is written in Go; see [`README.md`](README.md) for build and run
  instructions and [`AGENTS.md`](AGENTS.md) for the conventions and tooling
  (including the spec-driven [OpenSpec](https://openspec.dev) workflow under
  `openspec/`).
- Before opening a pull request, please make sure the following pass:

  ```sh
  go build ./...
  go test ./...
  golangci-lint run ./...
  ```

## Other Ways to Contribute

We welcome anyone that wants to contribute to `scorecard-mcp` to triage and
reply to open issues to help troubleshoot and fix existing bugs. Here is what
you can do:

- Help ensure that existing issues follow the recommendations from the
  _[Reporting Issues](#reporting-issues)_ section, providing feedback to the
  issue's author on what might be missing.
- Review existing pull requests, and test patches against real applications that
  use `scorecard-mcp`.
- Write a test, or add a missing test case to an existing test.

Thanks again for your interest in contributing to `scorecard-mcp`!

:heart:
