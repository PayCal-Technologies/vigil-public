# vigil Testing Guide

This document explains how to test vigil locally, what each test suite covers,
and which checks are expected before a pull request, release candidate, or
high-risk change.

vigil is a local safety and evidence layer around repository automation. Its
tests therefore cover more than ordinary command success. They verify command
contracts, mutation guards, output schemas, release artifacts, plugin trust,
configuration migration, process cancellation, performance budgets, and
cross-platform behavior.

## Official Resources

- vigil homepage: [sltd.ca/vigil](https://sltd.ca/vigil/)
- vigil help: [sltd.ca/vigil/help](https://sltd.ca/vigil/help/)
- GitHub repository: [github.com/PayCal-Technologies/vigil-public](https://github.com/PayCal-Technologies/vigil-public)
- Created by [Starlit Digital](https://sltd.ca/)

## Quick Start

Run the default test suite from the repository root:

```bash
go test ./...
```

For a normal code change, run:

```bash
gofmt -w $(find . -name '*.go' -not -path './graft/*')
go test ./...
```

For a change that touches command execution, plugins, plans, output formats,
release tooling, schema files, or safety policy, run the fuller local matrix:

```bash
go test ./...
go test -race ./...
go vet ./...
scripts/run-fuzz-smoke.sh
scripts/check-performance.sh
```

If you change GitHub Actions or shell scripts, also run:

```bash
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.9 .github/workflows/*.yml
bash -n scripts/*.sh
```

## What Exists Today

Current test inventory:

| Area | Count |
| --- | ---: |
| Go test files | 36 |
| `func Test...` tests | 297 |
| `func Fuzz...` targets | 7 |

Package-level coverage:

| Package | Test/fuzz functions | Main coverage |
| --- | ---: | --- |
| `cmd/vigil` | 106 | CLI dispatch, command registry behavior, mutation confirmation, workflows, hooks, plans, Sentry mode, setup, config commands, plugin CLI, pack commands, cleanliness commands, beginner-facing output, machine output adapters. |
| `cmd/vigil-plugin-publisher` | 1 | Offline publisher key generation and signed index authoring utility. |
| `internal/acceptance` | 29 | v1 acceptance ledger, stable gate, repository evidence, operational evidence, external evidence, schema-bound gate reports. |
| `internal/atomicfile` | 4 | Atomic writes, backup behavior, private mode, symlink rejection, fuzz round trips. |
| `internal/cache` | 2 | Bounded LRU behavior and concurrent use. |
| `internal/cli` | 5 | Command registry validation, exit taxonomy, structured flags and arguments, invocation context. |
| `internal/cleanliness` | 3 | Disposable artifact discovery, protected directory defaults, opt-in logs/vendor cleanup, guarded deletion behavior. |
| `internal/completion` | 2 | Registry-driven completion generation and unsupported-shell rejection. |
| `internal/config` | 7 | Config defaults, migration, validation, environment file path confinement, unknown-field preservation, config fuzzing. |
| `internal/contracts` | 32 | Published JSON schemas, workflow contracts, release workflow policy, acceptance schema compatibility. |
| `internal/externalevidence` | 9 | Independent review and external evidence report validation. |
| `internal/git` | 12 | Git root/status parsing, mutation fingerprints, dirty submodules, large dirty repos, special files. |
| `internal/hooks` | 3 | Hook chaining, rollback, symlink rejection. |
| `internal/operationalevidence` | 5 | Operational release evidence validation. |
| `internal/output` | 17 | JSON, JSONL, JUnit, SARIF, GitHub annotations, envelopes, status mapping, UTF-8 trimming. |
| `internal/packs` | 13 | Embedded packs, repository/user precedence, manifest validation, path confinement, pack fuzzing. |
| `internal/plan` | 8 | Reviewed plan IDs, private strict reads/writes, symlink and oversize rejection, digesting, golden compatibility. |
| `internal/plugins` | 19 | Plugin lifecycle, conformance, trust/revocation, signed indexes, HTTPS acquisition, strict protocol fuzzing. |
| `internal/releasearchive` | 2 | Reproducible release archives and symlink rejection. |
| `internal/runartifact` | 6 | Private run logs, manifest validation, bounded logs, refusal to overwrite logs. |
| `internal/runner` | 14 | Direct argv execution, explicit shell mode, timeouts, cancellation, output truncation, command-line parsing fuzzing. |
| `internal/support` | 7 | Support bundle determinism, private permissions, redaction, redaction fuzzing. |
| `internal/workflow` | 5 | Workflow graph dependency filtering, bounded parallel groups, fail-closed state transitions. |

## CI Quality Matrix

The main quality workflow is `.github/workflows/vigil.yml`.

On pull requests and pushes to `main`, it runs on:

- `ubuntu-latest`
- `macos-latest`

The workflow performs:

1. `gofmt` check.
2. GitHub Actions syntax validation with `actionlint`.
3. Shell syntax validation with `bash -n scripts/*.sh`.
4. `go test ./...`.
5. `go test -race ./...`.
6. Fuzz smoke targets through `scripts/run-fuzz-smoke.sh`.
7. `go vet ./...`.
8. `staticcheck`.
9. `govulncheck`.
10. Local release-like binary build.
11. Startup performance budget enforcement.
12. External plugin conformance with the reference plugin.
13. `vigil verify --json`.
14. `vigil workflow:local --json`.
15. Release evidence rebuild and reproducibility check on Linux.
16. Clean-directory binary smoke.

## Test Commands By Purpose

Use this table to choose the right command.

| Goal | Command |
| --- | --- |
| Run all normal tests | `go test ./...` |
| Catch data races and cancellation bugs | `go test -race ./...` |
| Check Go static correctness | `go vet ./...` |
| Run lint/static analysis used by CI | `go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...` |
| Check known Go vulnerabilities | `go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...` |
| Run fuzz smoke targets | `scripts/run-fuzz-smoke.sh` |
| Enforce startup performance budgets | `scripts/check-performance.sh` |
| Validate GitHub workflow files | `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.9 .github/workflows/*.yml` |
| Validate shell scripts parse | `bash -n scripts/*.sh` |
| Build the main binary | `CGO_ENABLED=0 go build -trimpath -o bin/vigil ./cmd/vigil` |
| Build the publisher utility | `CGO_ENABLED=0 go build -trimpath -o bin/vigil-plugin-publisher ./cmd/vigil-plugin-publisher` |
| Run vigil's self-checks | `bin/vigil verify --json` |
| Run configured local gates | `bin/vigil workflow:local --json` |
| Test reference plugin conformance | `bin/vigil --config examples/vigil.config.json --allow-mutation plugins:conformance --file examples/plugins/reference/vigil-plugin-reference --execute --json` |

## Fuzz Smoke Targets

`scripts/run-fuzz-smoke.sh` runs each fuzz target for a short bounded interval.
The default fuzz time is `2s`. Override it with `FUZZ_TIME`.

```bash
FUZZ_TIME=10s scripts/run-fuzz-smoke.sh
```

Targets:

| Package | Fuzz target | Purpose |
| --- | --- | --- |
| `internal/config` | `FuzzConfigDocumentRoundTrip` | Config parser and unknown-field preservation. |
| `internal/packs` | `FuzzPackManifestParsing` | Pack manifest parsing and validation robustness. |
| `internal/packs` | `FuzzPathConfinement` | Repository path confinement and traversal rejection. |
| `internal/runner` | `FuzzCommandLineParsing` | Shell-like command-line parsing boundaries. |
| `internal/atomicfile` | `FuzzAtomicWriteRoundTrip` | Atomic write round-trip safety. |
| `internal/plugins` | `FuzzStrictPluginHandshakeJSON` | Plugin protocol handshake JSON strictness. |
| `internal/support` | `FuzzSupportRedaction` | Support bundle redaction resilience. |

## Performance Tests

Performance is a release contract for fast read-only startup commands.

Run:

```bash
scripts/check-performance.sh
```

The script builds a release-like binary and measures new-process startup in an
empty directory with empty user pack and plugin roots.

Enforced median budgets:

| Operation | Budget |
| --- | ---: |
| `vigil version` | 20 ms |
| `vigil help` | 50 ms |
| `vigil list --json` | 100 ms |
| `vigil setup --dry-run --json` | 500 ms |

The checker also records p95 as diagnostic evidence. Median is the enforced
budget. If a performance test fails, re-run on an idle machine before changing
code. If it still fails, compare `version` with `list --json` to distinguish
process startup cost from discovery cost.

## Release And Archive Tests

Release tooling is tested both by Go tests and shell scripts.

Important commands:

```bash
RELEASE_TAG=v0.0.0-quality.1 \
VERSION=0.0.0-quality.1 \
COMMIT="$(git rev-parse HEAD)" \
SOURCE_DATE_EPOCH="$(git show -s --format=%ct HEAD)" \
ALLOW_UNTAGGED_RELEASE=1 \
scripts/build-release.sh
```

Then verify reproducibility:

```bash
scripts/check-release-reproducibility.sh dist
```

Smoke-test a native archive:

```bash
scripts/smoke-release-archive.sh dist/vigil_VERSION_GOOS_GOARCH.tar.gz VERSION GOOS GOARCH
```

The smoke script verifies:

- archive path confinement;
- no symlink or special-file entries;
- expected executable presence;
- schema count and JSON validity;
- acceptance gate report presence;
- binary version behavior;
- embedded pack behavior;
- platform match for native smoke.

## v1 Acceptance Tests

vigil has a separate v1 acceptance model because `v1.0.0` must not be claimed
until all required evidence is verified.

Primary files:

- `docs/v1-acceptance-matrix.md`
- `docs/v1-acceptance.json`
- `scripts/v1-acceptance-check.go`
- `internal/acceptance/*_test.go`
- `internal/contracts/v1_acceptance*_test.go`

Run the gate:

```bash
go run ./scripts/v1-acceptance-check.go --version 0.0.0-quality.1
go run ./scripts/v1-acceptance-check.go --version 0.0.0-quality.1 --json
```

Stable v1 tags intentionally fail closed while any acceptance criterion remains
pending.

## Golden Fixtures And Compatibility Tests

Some tests compare runtime output or command metadata against checked-in
fixtures. These are deliberate compatibility gates.

Important examples:

- `cmd/vigil/testdata/v1-command-contract.json`
- `internal/output/testdata/envelope-v1-success.json`
- `internal/output/testdata/envelope-v1-failure.json`
- `internal/plan/testdata/plan-v1.json`

Only update golden fixtures when the compatibility change is intentional and
reviewed. For the command contract fixture:

```bash
UPDATE_VIGIL_GOLDENS=1 go test ./cmd/vigil -run TestV1CommandContractGolden -count=1
```

After regenerating a golden, inspect the diff carefully. A valid diff should be
small, explainable, and documented in `CHANGELOG.md` when it affects public
behavior.

## Cross-Platform Testing

vigil is built as `CGO_ENABLED=0` Go binaries. The release script currently
cross-builds archives for:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`
- `windows/amd64`
- `windows/arm64`

Local cross-build smoke:

```bash
for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64; do
  os="${target%/*}"
  arch="${target#*/}"
  suffix=""
  if [ "$os" = windows ]; then suffix=".exe"; fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -o "/tmp/vigil-${os}-${arch}${suffix}" ./cmd/vigil
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -o "/tmp/vigil-plugin-publisher-${os}-${arch}${suffix}" ./cmd/vigil-plugin-publisher
done
```

Cross-building proves the code compiles for the target. It does not replace
native archive smoke on the actual operating system.

## When To Run Which Suite

Small documentation-only change:

```bash
git diff --check
```

Go-only implementation change:

```bash
gofmt -w <changed .go files>
go test ./...
```

Command registry, output, config, plan, runner, plugin, or workflow change:

```bash
go test ./...
go test -race ./...
scripts/run-fuzz-smoke.sh
```

Performance-sensitive startup/discovery change:

```bash
go test ./...
scripts/check-performance.sh
```

Release, archive, Homebrew, signing, checksum, or workflow change:

```bash
go test ./...
go test -race ./...
go vet ./...
scripts/run-fuzz-smoke.sh
scripts/check-performance.sh
bash -n scripts/*.sh
go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.9 .github/workflows/*.yml
```

Security-sensitive path, file, plugin, config, or command-execution change:

```bash
go test ./...
go test -race ./...
scripts/run-fuzz-smoke.sh
go run golang.org/x/vuln/cmd/govulncheck@v1.6.0 ./...
```

## Common Failures

`gofmt` failure:

- Run `gofmt -w` on the named files.

Golden fixture failure:

- Do not blindly regenerate.
- Confirm the public contract changed intentionally.
- Regenerate only the relevant fixture.
- Review the diff before committing.

Race detector failure:

- Treat it as a correctness bug.
- Look for shared maps, slices, caches, process state, or test globals.
- Add a focused regression test if the race came from production code.

Fuzz smoke failure:

- Reduce the input if possible.
- Add the reduced case as a normal unit test.
- Keep the fuzz target broad enough to catch future variants.

Performance failure:

- Re-run on an idle machine.
- Separate process startup from discovery work.
- Prefer removing unnecessary startup work over raising budgets.

Mutation violation failure:

- Check whether a read-only command wrote to the Git-visible workspace.
- If the mutation is expected, the command contract should be conditional-write
  or write, and it must require explicit mutation approval.

## Adding Tests For New Code

Use these rules of thumb:

- New command: add command registry tests, help/list/JSON behavior, mutation
  guard behavior, and at least one CLI-level success test.
- New output format: add envelope/schema tests and escaping tests.
- New config field: add defaulting, validation, migration, and unknown-field
  preservation tests.
- New file write: add private permission, symlink rejection, path confinement,
  and rollback/atomicity tests.
- New process execution behavior: add timeout, cancellation, stderr/stdout
  boundary, and missing-tool tests.
- New release artifact: add archive, reproducibility, checksum, and smoke
  coverage.
- New plugin capability: add conformance, trust policy, revocation, and digest
  drift tests.

Prefer small focused unit tests for pure logic and CLI-level tests for behavior
users or agents rely on.

## Before Pushing

Minimum:

```bash
git status --short
gofmt -w <changed .go files>
go test ./...
git diff --check
```

Recommended for non-trivial changes:

```bash
go test -race ./...
scripts/run-fuzz-smoke.sh
scripts/check-performance.sh
```

Commit only after the relevant suite passes or the failure is clearly unrelated
and documented.
