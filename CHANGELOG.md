# Changelog

All notable public changes are documented here.

This project uses semantic version tags. Prerelease tags such as
`v0.2.0-beta.1` are public candidates; stable v1 tags are blocked until the v1
acceptance ledger is complete.

## Unreleased

### Added

- Add `cleanliness:scan` and guarded `cleanliness:clean` commands for reporting
  and deleting disposable local artifacts with `--dry-run`, JSON output, and
  opt-in logs/vendor cleanup.

### Changed

- Reuse the new cleanliness scanner for `checks:workspace-hygiene` so the
  read-only check and cleanup preview report the same disposable artifact set.

## v0.2.0-beta.1 - 2026-08-22

This public source prerelease validates vigil's new release and distribution
contract before any stable v1 claim. Its current operational status is recorded
in [`docs/release-state.json`](docs/release-state.json).

### Added

- Add detailed `TESTING.md` coverage guidance for unit, integration,
  acceptance, release, completion, schema, Windows, and troubleshooting checks.
- Add `help/testing.html` to the bundled vigil help system with direct links to
  the official vigil homepage, public help site, and GitHub repository.
- Add command completion generators for PowerShell, Elvish, Nushell, Xonsh,
  Tcsh, Clink, Yash, and Oil-family shells while preserving Bash, Zsh, and Fish.
- Expand release cross-build targets to Windows amd64 and arm64 archives with
  `.exe` entrypoints for both `vigil` and `vigil-plugin-publisher`.
- Add beginner-facing command aliases and guided language for setup, checks,
  explanation, status, learning, and safe next actions.
- Typed command registry used by help, command discovery, completions, manpages,
  and machine-readable command metadata.
- Versioned configuration, output, plan, run-artifact, plugin protocol, plugin
  index, publisher, lock, and trust schemas.
- Digest-bound reviewed plans that fail closed when execution inputs change.
- Cancellable direct-argv command execution with explicit shell opt-in.
- Git-visible mutation detection for read-only gates, including workflow graph
  execution with bounded parallel read-only groups.
- Embedded official packs that work without a vigil source checkout.
- Digest-bound subprocess plugin execution with explicit local capability
  approval and conformance testing.
- Release automation for signed macOS archives, Linux archives, checksums, SPDX
  SBOMs, Sigstore checksum bundles, GitHub attestations, native archive smoke,
  and generated Homebrew formula validation.
- v1 acceptance gate reporting through `v1-acceptance-gate.json`.

### Changed

- Add presentation-oriented terminology that describes checks, file changes,
  access, saved reports, reviewed plans, feature collections, and executable
  extensions in plainer language without changing machine contracts.
- Clarify README and troubleshooting wording around executable extensions:
  access approval is not an operating-system sandbox.
- Polish release, Homebrew, and GitHub repository metadata.

- `extensions:*` commands remain as compatibility aliases for declarative packs;
  executable plugins now have their own plugin lifecycle and protocol.
- Workflow execution now exposes stable execution states and versioned machine
  output envelopes.
- Release builds require exact semantic-version metadata and annotated tags
  unless local development overrides are explicitly set.

### Known Gaps

- Stable `v1.0.0` remains blocked until external and operational acceptance
  criteria are independently verified.
- The `v0.2.0-beta.1` source tag and GitHub prerelease exist, but signed
  archives, notarization evidence, checksums, SBOM, Sigstore bundles, GitHub
  attestations, native-smoke results, and Homebrew publication remain pending.
- Public Homebrew install instructions should remain disabled until the project
  tap publishes and validates a formula from public release URLs.

## v0.1.0

- Initial public source release.

## Unreleased execution and release hardening

Configuration schema 3 and plan schema 1 gain optional `required_env`,
`executable_sha256`, `version_args` and `version_pattern` gate fields. This is an
additive compatibility decision; the command catalogue, output envelope and exit
code meanings remain unchanged. Plans containing these fields need an updated
binary and must be reviewed again because their digest includes the requirements.

Skipped optional tools no longer satisfy dependencies. Required blocked
dependents fail the workflow. A gate timeout now covers retries and backoff as
one budget. Direct runner calls default to five minutes, with a 24 hour ceiling.
Release builds enforce checkout identity, reject stable override flags, preserve
previous artifacts on failure and bound release jobs to 45 minutes.
