# Vigil AI Video Transcript

## Production Notes

Internal direction: Use this transcript as the source for an AI-generated explainer and product walkthrough video. Treat bracketed `Direction` notes as scene, pacing, screen-capture, animation, and emphasis guidance. Treat fenced code blocks as on-screen terminal, editor, or overlay snippets. The narrator text can be read directly.

Internal direction: Visual style should feel like a focused developer tool walkthrough: terminal, editor, repository graph, policy gates, structured JSON output, and concise callouts. Avoid implying that Vigil is a hosted CI system, deployment orchestrator, secrets manager, or operating-system sandbox. Repeatedly distinguish visibility, policy, approval, and evidence from confinement.

Internal direction: Suggested runtime is 10 to 14 minutes for a complete feature video, or 4 to 6 minutes if edited down. Keep the full transcript available for long-form documentation, launch pages, demos, and agent-generated chaptered videos.

## Title Card

[Direction: Fade in on a repository opened in an editor. Show `vigil.config.json`, a terminal prompt, and a clean Git status. Overlay the title: "Vigil: Policy-Aware Preflight for Software Projects."]

Narrator:

Before you publish, push, deploy, or hand off a software project to another person or coding agent, there are five questions you need answered.

What will run?

Why will it run?

What can it access or change?

What evidence will it produce?

And can the exact reviewed plan run later without being reinterpreted?

Vigil is a policy-aware repository preflight engine built to answer those questions.

It makes local automation visible, reviewable, executable, and verifiable.

## Introduction

[Direction: Show a fast montage: a developer about to push code, an AI coding agent preparing a change, a release workflow, and a terminal running checks.]

Narrator:

Modern projects have more automation than ever. Test runners, linters, security scanners, release scripts, dependency tools, accessibility checks, README generators, Git hooks, CI helpers, and AI-generated commands all live beside the source code.

That automation is useful, but it also creates risk. A command can run more than intended. A read-only check can quietly write files. A setup script can run after the project changed. A plugin can request broad local access without a clear review trail.

Vigil's job is to reduce accidental and over-broad automation.

It does not pretend to be a kernel sandbox. Approved commands still run as the user. Instead, Vigil makes the boundaries explicit: the command contract, the requested access, the mutation approval path, the repository fingerprint, the plan digest, and the final machine-readable evidence.

[Direction: Animate a simple pipeline.]

```text
typed command registry
        |
access and mutation policy
        |
reviewed plan checks
        |
bounded command execution
        |
repository mutation evidence
        |
structured output
```

Narrator:

The result is a local command surface that humans, agents, and CI systems can inspect and trust at the policy level.

## What Vigil Is

[Direction: Show a plain definition card. Keep it restrained and technical.]

Narrator:

Vigil is a standalone CLI for repository-local preflight checks.

It owns repository policy, deterministic command discovery, preflight planning, approval, process execution, cancellation, result classification, mutation detection, machine-readable evidence, and local run artifacts.

It is designed for teams and coding agents that need to know which automation is safe to run, which automation changes files, and whether the result matches the reviewed setup.

[Direction: Show "Not a hosted CI service", "Not a deployment orchestrator", "Not a secrets manager", "Not an OS sandbox" as clear boundary labels.]

Narrator:

Vigil is not a hosted CI service.

It is not a deployment orchestrator.

It is not a secrets manager.

And it is not an operating-system sandbox.

That boundary matters. Vigil focuses on review, policy, verification, and evidence. It helps you catch unexpected Git-visible project changes and stale reviewed plans. It does not claim to contain every possible side effect of an approved process.

## Installation And First Run

[Direction: Show a terminal building Vigil from source, then running the beginner command surface.]

Narrator:

You can install the current tagged release with Go, install the development preview from `main`, or build the current checkout.

```bash
go install github.com/PayCal-Technologies/vigil-public/cmd/vigil@v0.1.0
go install github.com/PayCal-Technologies/vigil-public/cmd/vigil@main
go build -trimpath -o bin/vigil ./cmd/vigil
```

Narrator:

The beginner-friendly command surface on `main` includes simple commands for setup, explanation, checking, status, and learning.

```bash
vigil
vigil setup
vigil explain
vigil check
vigil status
vigil learn hooks
```

[Direction: Show `vigil` without a config. Present a guided setup prompt, then a generated `vigil.config.json`.]

Narrator:

Running `vigil` with no command introduces the project and, on an interactive terminal, can guide first-time setup. The simple command `vigil check` runs the same configured local workflow as `vigil workflow:local`, but gives beginners a direct entry point.

## The Core Workflow

[Direction: Split screen: left side `vigil.config.json`, right side terminal. Highlight one gate at a time.]

Narrator:

Vigil reads a versioned JSON configuration, normally named `vigil.config.json`.

The current schema is `3`. A basic gate names an executable, passes arguments as an argv array, declares whether the gate is read-only, assigns tags, and can set a timeout.

```json
{
  "name": "go test",
  "command": "go",
  "args": ["test", "./..."],
  "read_only": true,
  "tags": ["test", "pre-push"],
  "timeout": "5m"
}
```

Narrator:

By default, Vigil executes commands directly with argv. Shell behavior is explicit. If a check needs shell syntax, the config says so.

```json
{
  "name": "report",
  "command": "go test ./... | tee test.log",
  "shell": true,
  "read_only": false,
  "timeout": "10m"
}
```

Narrator:

This distinction prevents accidental shell interpretation and makes the command contract easier to inspect.

## Full Configuration Example

[Direction: Show a complete starter config. Highlight coordination, gates, extensions, and plugins.]

Narrator:

A typical Vigil config identifies the project, describes coordination policy, lists ordered workflow gates, and controls built-in feature collections and executable plugins.

```json
{
  "schema_version": "3",
  "profile": "go-tool",
  "project": "example-vigil-project",
  "coordination": {
    "mode": "policy-aware-preflight",
    "authoritative_surfaces": [
      "GitHub",
      "reviewed repository configuration",
      "active workflow files"
    ],
    "mutation_requires": [
      "explicit-confirmation",
      "clean-config"
    ]
  },
  "gates": [
    {
      "name": "go test",
      "command": "go",
      "args": ["test", "./..."],
      "read_only": true,
      "tags": ["test", "pre-push"]
    },
    {
      "name": "vigil verify",
      "command": "vigil",
      "args": ["verify", "--json"],
      "read_only": true,
      "tags": ["vigil", "pre-commit"]
    }
  ],
  "extensions": {
    "enabled": true,
    "manifest_root": "extensions",
    "allowed_kinds": ["custom"],
    "enabled_ids": ["scribe", "repo-health", "github-cicd"],
    "disabled_ids": [],
    "require_private": false
  },
  "plugins": {
    "mode": "enabled",
    "local": "allow",
    "require_signed": false,
    "min_signature_threshold": 0,
    "allowed_ids": [],
    "denied_ids": [],
    "allowed_publisher_key_ids": [],
    "denied_capabilities": []
  }
}
```

Narrator:

The configuration is bounded and validated. It must be a regular, non-symlink JSON file no larger than four MiB. The same reader is used for validation, setup, migration, repair, pack settings, and startup.

## Command Discovery And Explanation

[Direction: Show `vigil list`, then `vigil explain workflow:local`, then `vigil guards:summary`. Use callouts for source, access, and formats.]

Narrator:

Vigil's command registry is the authoritative source of command identity.

Every command has a typed contract: name, aliases, access class, capabilities, binding, stability, timeout, tools, network behavior, and supported output formats.

You can list commands, explain a command, or summarize which commands are read-only and which can write.

```bash
vigil list
vigil explain workflow:local
vigil guards:summary
vigil tools:catalog --json
```

Narrator:

The same registry drives help, command discovery, completions, manpages, policy checks, plugin wrapping, and machine output.

Unknown access values, missing handlers, duplicate names, invalid aliases, unsupported bindings, and malformed output contracts fail before dispatch.

## Beginner And Advanced Commands

[Direction: Present commands in clusters. Avoid dumping all commands visually at once; animate categories.]

Narrator:

Vigil includes a beginner command surface:

```bash
vigil check
vigil setup
vigil explain
vigil status
vigil learn
vigil fix
vigil next
```

Narrator:

And an advanced command interface for deeper workflows:

```bash
vigil config:validate
vigil config:migrate
vigil config:repair
vigil config:report
vigil workflow:local
vigil plan
vigil apply
vigil verify
vigil support:bundle
```

Narrator:

Compatibility aliases remain available. For example, `vigil init` maps to the setup wizard, `vigil list` maps to command listing, and `vigil check` maps to the configured local workflow.

## Access Classes And Mutation Approval

[Direction: Show three labeled cards: `read`, `write`, `conditional-write`. Then show a command blocked until `--allow-mutation` is supplied.]

Narrator:

Vigil separates declared access from observed evidence.

Commands declare one of three access classes.

`read` means no declared repository or external mutation.

`write` means mutation is the normal operation.

`conditional-write` means mutation is activated by registered write flags.

Write paths require explicit authorization, usually with `--allow-mutation`.

```bash
vigil plan --json
vigil --allow-mutation plan --output .vigil/plans/reviewed.json
```

Narrator:

The second command writes a plan file. Writing evidence is itself a filesystem mutation, so the command requires explicit approval.

## Repository Mutation Detection

[Direction: Animate a Git-visible fingerprint before and after a check. Show status, binary diff, staged diff, and untracked content hashes being captured.]

Narrator:

For read-only workflow gates, Vigil fingerprints the Git-visible repository before and after execution.

It records Git porcelain status, unstaged binary diff, staged binary diff, and names and content digests of untracked, non-ignored files.

If the fingerprint changes, Vigil reports `mutation_detected` and exits with code `6`, even if the underlying command exited successfully.

```text
before fingerprint
        |
run declared read-only gate
        |
after fingerprint
        |
same? pass
different? mutation_detected
```

Narrator:

Fingerprinting fails closed. If Vigil cannot safely collect complete Git-visible evidence, it blocks the read-only gate instead of pretending the workspace is clean.

[Direction: Add a boundary callout.]

Narrator:

This is repository verification, not an operating-system sandbox. Ignored files, files outside the repository, caches, databases, sockets, network state, external services, environment changes, and credential stores are outside this fingerprint boundary.

## Workflow Graphs

[Direction: Show a directed acyclic graph of gates: generated code, unit tests, integration tests, security scan, verify. Show read-only gates in one color and write gates in another.]

Narrator:

Vigil schema `3` gates form a validated directed acyclic graph.

Gates can depend on other gates, run in deterministic order, opt into bounded parallelism, continue after ordinary failures, declare artifacts, and retry carefully selected network checks.

```json
{
  "name": "integration tests",
  "command": "go",
  "args": ["test", "./integration/..."],
  "read_only": true,
  "depends_on": ["generated-code check"],
  "parallel_group": "tests",
  "continue_on_error": true,
  "required": true,
  "retry": {
    "max_attempts": 2,
    "delay": "1s",
    "on": ["failed"]
  },
  "cwd": "services/api",
  "environment": {"CI": "true"},
  "artifacts": [
    {
      "path": "reports/junit.xml",
      "kind": "junit",
      "media_type": "application/xml"
    }
  ],
  "tags": ["network", "pre-push"],
  "timeout": "5m"
}
```

Narrator:

Ungrouped gates run one after another. Ready read-only gates with the same non-empty `parallel_group` may run together. The `--jobs` flag bounds parallel work from one through thirty-two and defaults to four.

Mutating gates always run exclusively. A batch with any write-capable gate cannot run concurrently.

```bash
vigil workflow:local --dry-run
vigil workflow:local --tag pre-push
vigil workflow:local --jobs 4
```

Narrator:

If a parallel read-only batch changes the workspace, Vigil marks every gate in the batch as `mutation_detected`. That is deliberate. Assigning the mutation to one process would be guesswork, so Vigil reports group-level evidence.

## Reviewed Plans

[Direction: Show a plan being created, reviewed as JSON, stored privately, and later applied. Then show a stale plan being blocked after the repository changes.]

Narrator:

Reviewed plans are one of Vigil's most important features.

A plan is a complete, versioned description of the configured gates and the evidence that makes those gates safe to execute without reinterpretation.

Create a plan:

```bash
vigil plan --json
```

Write a private reviewed plan:

```bash
vigil --allow-mutation plan --output .vigil/plans/reviewed.json
```

Apply the reviewed plan later:

```bash
vigil --allow-mutation apply .vigil/plans/reviewed.json
```

Narrator:

When applying a plan, Vigil validates the plan schema and plan ID, then recomputes every reviewed input.

The apply is blocked if any of these changed: the Vigil executable digest, config path, config content digest, repository root, `HEAD`, Git-visible workspace digest, active command-registry digest, or active pack-registry digest.

There is intentionally no stale-plan override. If the reviewed inputs changed, generate and review a new plan.

[Direction: Show "No reinterpretation" as a large overlay.]

Narrator:

That means an approved plan is not silently reinterpreted against a different repository, a different config, a different binary, a different command list, or a different pack set.

## Structured Output

[Direction: Show a terminal with human-readable output, then switch to JSON output feeding an automation script.]

Narrator:

Vigil is built for humans and automation.

Human-facing commands stream useful terminal output. Machine-facing commands can emit stable JSON envelopes, JSONL event streams, JUnit XML, SARIF, and GitHub Actions annotations.

```bash
vigil verify --json
vigil workflow:local --format=jsonl
vigil verify --format=junit
vigil checks:staged-sensitive --format=sarif
vigil checks:workspace-hygiene --format=github
```

Narrator:

The common JSON envelope uses schema version `1`.

```json
{
  "schema_version": "1",
  "command": "verify",
  "status": "ok",
  "exit_code": 0,
  "started_at": "2026-07-31T18:00:00Z",
  "finished_at": "2026-07-31T18:00:00.125Z",
  "duration_ms": 125,
  "warnings": [],
  "errors": [],
  "data": {},
  "artifacts": []
}
```

Narrator:

All fields are present. Warnings, errors, and artifacts are arrays, not null. Command-specific data lives under `data`.

Programs should branch on `exit_code` and use `status` for display.

## Exit Codes

[Direction: Use a compact table overlay. Highlight codes as results appear in short demo clips.]

Narrator:

Vigil has a stable exit taxonomy.

```text
0  success
1  check failed
2  usage or configuration error
3  policy blocked execution
4  required dependency missing
5  timeout or cancellation
6  mutation violation
7  internal Vigil failure
```

Narrator:

Aggregate commands return the most specific applicable non-zero code. Unknown child or handler codes normalize to internal failure, so new public meanings do not appear accidentally.

## Run Artifacts And Logs

[Direction: Show `.vigil/runs/<run-id>/manifest.json`, stdout logs, stderr logs, and artifact digests. Mark the directory as private and Git-ignored.]

Narrator:

Saved local reports and logs are opt-in because writing them changes files.

```bash
vigil --allow-mutation workflow:local --artifacts
```

Narrator:

Run artifact directories start with a private manifest. Logs are bounded. Artifact paths are checked. In-repository plan and artifact paths are expected to be ignored so Vigil's own evidence does not invalidate a read-only check.

Declared artifacts are verified after execution. They do not authorize mutation.

If a read-only gate writes a Git-visible JUnit file, that gate still fails mutation detection unless the path is outside the fingerprint boundary or the gate is correctly declared as file-changing.

## Setup, Config, And Migration

[Direction: Show `vigil setup`, `vigil config:validate`, then an old config being migrated with a backup file.]

Narrator:

Vigil includes commands for setup, validation, migration, repair, templates, and redacted reports.

```bash
vigil setup
vigil setup:wizard
vigil config:init
vigil config:template --profile go-tool
vigil config:validate --json
vigil config:migrate --json
vigil --allow-mutation config:migrate --write
vigil config:repair
vigil config:report --json
```

Narrator:

Migration previews changes by default. Writing a migrated config requires mutation authorization. Vigil preserves unknown document fields, converts legacy authority policy into the current coordination model, converts simple legacy command strings into argv, marks shell-dependent command strings explicitly, and refuses to downgrade a future schema.

## Git Hooks

[Direction: Show a repository installing hooks. Show pre-commit and pre-push checks running through Vigil.]

Narrator:

Vigil can install safe Git hook shims, inspect hook state, run hook checks from config, and uninstall while restoring chained hooks.

```bash
vigil hooks:doctor
vigil --allow-mutation hooks:install
vigil hooks:pre-commit
vigil hooks:pre-push
vigil --allow-mutation hooks:uninstall
```

Narrator:

Hook installation and removal are file-changing operations, so they use the same mutation approval model as other writes.

## Local Readiness And Diagnostics

[Direction: Show a developer diagnosing a project with read-only commands. Keep the terminal output clean and quick.]

Narrator:

For day-to-day readiness, Vigil provides status, doctor, verification, dependency inventory, public-boundary checks, workspace hygiene checks, and redacted support bundles.

```bash
vigil status --json
vigil doctor --json
vigil verify --json
vigil deps:inventory
vigil checks:workspace-hygiene --json
vigil checks:staged-sensitive --format=sarif
vigil support:bundle --dry-run
vigil --allow-mutation support:bundle
```

Narrator:

Support bundles are local-only by default, redacted, written atomically, and private. Preview before writing and review before sharing.

## Built-In Feature Collections

[Direction: Show the `extensions/` directory and then `vigil extensions:list --json`. Use the term "feature collections" and mention compatibility with `extensions:*` names.]

Narrator:

Vigil ships embedded official feature collections. These are declarative packs: command and policy metadata bound to built-in implementations.

They are not arbitrary code loaded from JSON.

```bash
vigil extensions:list
vigil extensions:doctor --json
```

Narrator:

Official, user, and repository packs have deterministic precedence.

```text
embedded official < user < repository
```

Core commands cannot be replaced by pack metadata. Pack roots are confined to safe repository or user locations, and invalid manifests fail closed.

## Pack Policy

[Direction: Show the `extensions` policy block from config. Toggle enabled IDs and allowed kinds.]

Narrator:

Repository config can disable all packs, restrict allowed kinds, require private manifests, allow only selected IDs, disable selected IDs, and choose a repository manifest root.

```json
{
  "enabled": true,
  "manifest_root": "extensions",
  "allowed_kinds": ["custom"],
  "enabled_ids": [],
  "disabled_ids": [],
  "require_private": false
}
```

Narrator:

Each pack command contract includes command name, access, capabilities, implementation binding, timeout, stability, required tools, network behavior, expected output formats, usage, and description.

## Extension Examples

[Direction: Present extension demos as short chapters. Show each as a terminal command and one sentence of output purpose.]

Narrator:

The public repository includes several feature collection examples and built-in command contracts.

Accessibility adapters expose local accessibility checks:

```bash
vigil a11y:inventory
vigil a11y:pa11y https://example.test
vigil a11y:lighthouse https://example.test
vigil a11y:playwright
vigil a11y:smoke --json
```

Dependency adapters expose package and scanner checks:

```bash
vigil deps:inventory
vigil deps:why react
vigil checks:dependency-security --json
vigil npm:audit
vigil composer:validate
vigil php:lint --json
vigil phpstan:analyse
vigil javascript:quality --json
```

Deployment verification checks a public endpoint:

```bash
vigil deploy:verify --url https://example.test --json
```

File iteration demonstrates stable repository file traversal:

```bash
vigil files:iterate --root=. --glob='**/*.go' --format=jsonl
```

GitHub CI helpers generate workflow scaffolding:

```bash
vigil github:init-ci
vigil github:init-ci --write
```

Release policy checks generic release readiness:

```bash
vigil checks:release-policy --json
```

Repository health reports lightweight Git history diagnostics:

```bash
vigil repo:health --json
vigil history:diagnose --json
```

Scribe updates a managed README snapshot block:

```bash
vigil readme:generate
vigil readme:check
vigil readme:generate --dry-run
```

Security adapters expose local scanner contracts:

```bash
vigil security:gitleaks
```

Test adapters expose generic test helpers:

```bash
vigil tests:history --json
vigil tests:affected --json
```

## Executable Plugins

[Direction: Switch from "packs" to "plugins" with a clear boundary graphic. Packs are metadata. Plugins are separately installed executables.]

Narrator:

Packs and plugins are different.

Packs are declarative metadata.

Plugins are separately installed executables that communicate with Vigil through a strict subprocess protocol.

Plugins do not run inside the Vigil process, and a JSON manifest alone is not an executable plugin.

## Plugin Lifecycle

[Direction: Show local plugin install, list, doctor, update, and remove. Highlight local capability approval.]

Narrator:

Local-file plugin installation remains available when repository policy permits it.

```bash
vigil --allow-mutation plugins:install \
  --file ./vigil-plugin-example \
  --approve filesystem:read \
  --approve process

vigil plugins:list --json
vigil plugins:doctor --json

vigil --allow-mutation plugins:update \
  --file ./vigil-plugin-example \
  --approve-all

vigil --allow-mutation plugins:remove example
```

Narrator:

The user must approve every aggregate capability individually or explicitly approve all. Capability approval records what the plugin declares and what the user approved. It does not sandbox the plugin.

## Signed Plugin Acquisition

[Direction: Show a publisher key, signed index, exact plugin ID and semantic version, then trust and install steps.]

Narrator:

Signed-index installation adds publisher verification and exact release selection.

```bash
vigil --allow-mutation plugins:trust-publisher \
  --key ./publisher.pub \
  --name "Example Publisher"

vigil plugins:index:verify --index ./index-v1.json --json

vigil --allow-mutation plugins:install \
  --index ./index-v1.json \
  --id example \
  --version 1.2.3 \
  --approve filesystem:read

vigil plugins:publishers --json
vigil --allow-mutation plugins:revoke-publisher sha256:<key-digest>
```

Narrator:

There is no implicit latest selection. Exact ID and semantic version are mandatory.

Publisher trust never implies capability approval.

## Plugin Trust And Discovery

[Direction: Show `vigil.plugins.lock.json` and local trust state under the user config directory. Highlight digest-bound identity.]

Narrator:

Each repository records selected plugins in `vigil.plugins.lock.json`.

The lock pins plugin ID, semantic version, executable digest, metadata digest, protocol and host API versions, capabilities, command names, acquisition type, signed index digest, publisher signer IDs, signature threshold, and exact binding.

```text
plugin:<id>@<version>#sha256:<digest>
```

Narrator:

Local approvals and revocations live outside the repository in the platform user configuration directory. Repository lock state is shareable. User trust state is local and private.

During discovery, Vigil verifies strict lock and trust documents, revocation state, repository policy, publisher threshold, installed path, executable digest, bounded handshake, lock agreement, trust agreement, and command-name collisions.

Only plugins that pass every check enter the typed command registry.

## Plugin Protocol

[Direction: Show a plugin executable receiving a handshake and execute call. Animate strict JSON in and out.]

Narrator:

The plugin protocol is intentionally simple.

```text
<plugin> handshake --protocol-version=1
<plugin> execute --protocol-version=1 --command=<namespaced-command>
```

Narrator:

The handshake writes strict JSON to standard output. Execution reads one strict JSON request on standard input and returns one strict JSON response on standard output.

Unknown fields, missing required fields, extra JSON values, unsupported versions, malformed command contracts, oversized output, and request-ID mismatches fail closed.

Plugin authors can validate an executable without installing or trusting it.

```bash
vigil plugins:conformance \
  --file ./vigil-plugin-example \
  --json

vigil --allow-mutation plugins:conformance \
  --file ./vigil-plugin-example \
  --execute \
  --timeout 30s \
  --json
```

## Plugin Policy

[Direction: Show a config snippet denying local installs and requiring signed publishers.]

Narrator:

Repository config can control plugin acquisition and capability policy.

```json
{
  "mode": "enabled",
  "local": "deny",
  "require_signed": true,
  "min_signature_threshold": 2,
  "allowed_ids": [],
  "denied_ids": [],
  "allowed_publisher_key_ids": [],
  "denied_capabilities": ["secrets"]
}
```

Narrator:

Policy is enforced during installation and every discovery. Tightening repository policy immediately removes noncompliant locked commands from the active registry without rewriting the lock.

## GitHub Actions And CI Use Case

[Direction: Show a repository adding a CI workflow generated by Vigil, then a GitHub Actions run executing `vigil verify` and `vigil workflow:local`.]

Narrator:

Vigil can help bridge local preflight and CI.

Use the GitHub helper to generate workflow scaffolding:

```bash
vigil github:init-ci
vigil github:init-ci --write
```

Or run Vigil directly in CI:

```yaml
name: Vigil

on:
  pull_request:
  push:
    branches: [main]

jobs:
  vigil:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - run: go install github.com/PayCal-Technologies/vigil-public/cmd/vigil@main
      - run: vigil verify --json
      - run: vigil workflow:local --format=jsonl
```

Narrator:

In CI, Vigil's structured output, fixed exit taxonomy, and repository policy make it easier to distinguish ordinary check failures from blocked policy, missing dependencies, interrupted runs, mutation violations, and internal failures.

## Example Use Case: AI Coding Agent Handoff

[Direction: Show an AI agent completing code, then running Vigil before handing back to a human reviewer.]

Narrator:

In an AI-assisted coding workflow, a coding agent can run Vigil before handoff.

```bash
vigil status
vigil explain workflow:local
vigil workflow:local --dry-run
vigil check
```

Narrator:

The agent can show what will run, why it will run, and whether any supposedly read-only check changed the repository.

For higher-assurance handoff, the agent can create a reviewed plan for the human to inspect.

```bash
vigil --allow-mutation plan --output .vigil/plans/agent-handoff.json
```

The human can later apply that exact plan:

```bash
vigil --allow-mutation apply .vigil/plans/agent-handoff.json
```

If the repository changed after review, Vigil blocks the apply and asks for a new plan.

## Example Use Case: Pre-Push Local Gate

[Direction: Show a developer preparing to push. A pre-push tag selects tests and verification gates.]

Narrator:

Before pushing, a developer can run only the gates tagged for pre-push, including transitive dependencies.

```bash
vigil workflow:local --tag pre-push
```

Config:

```json
{
  "name": "go test",
  "command": "go",
  "args": ["test", "./..."],
  "read_only": true,
  "tags": ["test", "pre-push"]
}
```

Narrator:

If the test command writes unexpected Git-visible files, the gate fails with a mutation violation. That gives the developer a direct signal that the check is not truly read-only or that generated files need to be handled intentionally.

## Example Use Case: Release Readiness

[Direction: Show a release manager running checks, building an archive, and verifying checksums.]

Narrator:

For release preparation, Vigil combines local readiness checks, release policy checks, signed release archive tooling, checksums, SBOM support, Sigstore bundles, GitHub attestations, and v1 acceptance evidence.

```bash
vigil verify --json
vigil checks:release-policy --json
scripts/check-github-release-readiness.sh --tag v0.2.0-beta.1
RELEASE_TAG=v0.2.0-beta.1 scripts/build-release.sh
RELEASE_TAG=v0.2.0-beta.1 scripts/check-release-reproducibility.sh dist
shasum -a 256 -c dist/SHA256SUMS
```

Narrator:

Vigil's stable `v1.0.0` gate is intentionally blocked until the checked-in acceptance ledger has verified evidence for every required safety invariant and acceptance criterion.

## Example Use Case: Dependency And Security Checks

[Direction: Show a project with JavaScript, PHP, and Go files. Commands light up based on available local tools.]

Narrator:

Vigil can inventory common dependency manifests and run local dependency or scanner adapters.

```bash
vigil deps:inventory
vigil deps:why react
vigil checks:dependency-security --json
vigil npm:audit
vigil composer:validate
vigil php:lint --json
vigil security:gitleaks
```

Narrator:

If a required tool is unavailable, Vigil reports dependency missing with exit code `4` instead of hiding the problem.

## Example Use Case: Accessibility Checks

[Direction: Show a local web app and browser page, then terminal commands for accessibility adapters.]

Narrator:

For web projects, accessibility adapters can wrap local tools like Pa11y, Lighthouse, and Playwright.

```bash
vigil a11y:inventory
vigil a11y:pa11y https://example.test
vigil a11y:lighthouse https://example.test
vigil a11y:playwright
vigil a11y:smoke --json
```

Narrator:

These are public command contracts around local tools. Missing tools report clearly, and projects can also wire custom accessibility checks directly into `vigil.config.json`.

## Example Use Case: README And Repository Health

[Direction: Show README managed block updating, then a repo health diagnostic report.]

Narrator:

Vigil can help keep repository metadata current.

Scribe manages a README snapshot block without taking ownership of the rest of the file.

```bash
vigil readme:generate
vigil readme:check
vigil readme:generate --dry-run
```

Repository health diagnostics summarize lightweight Git history signals.

```bash
vigil repo:health
vigil repo:health --json
vigil history:diagnose --json
```

## Example Use Case: Public Boundary Hygiene

[Direction: Show a public-source repository. Overlay examples: local editor artifacts, OS files, sensitive staged files, assistant artifacts.]

Narrator:

Vigil includes public-boundary and workspace hygiene checks.

```bash
vigil checks:public-assumptions
vigil checks:public-parity
vigil checks:workspace-hygiene
vigil checks:staged-sensitive --format=sarif
vigil checks:tracked-assistant-artifacts
```

Narrator:

These checks are useful before publishing, open-sourcing, or handing a repository to another system.

## Example Use Case: Plugin Author Workflow

[Direction: Show a plugin author editing `vigil-plugin-example`, running conformance, then preparing a signed index.]

Narrator:

Plugin authors can build standalone executables in any language that can read and write strict JSON over standard streams.

The executable name follows the pattern:

```text
vigil-plugin-<id>
```

A plugin declares namespaced commands, access, capabilities, stability, timeout, network behavior, required tools, output formats, flags, positional arguments, usage, and examples.

Authors run conformance before distribution:

```bash
vigil plugins:conformance \
  --file ./vigil-plugin-example \
  --json

vigil --allow-mutation plugins:conformance \
  --file ./vigil-plugin-example \
  --execute \
  --timeout 30s \
  --json
```

Narrator:

Conformance checks repeated handshake determinism, command contracts, acquisition policy, metadata policy, request and response shape, artifacts, and executable digest stability.

## Completions And Manpages

[Direction: Show shell completion suggestions and a generated manpage.]

Narrator:

Because Vigil's command registry is typed, completions and manpages are generated from the same source of truth.

```bash
vigil completion bash
vigil completion zsh
vigil completion fish
vigil completion powershell
vigil completion nushell
vigil manpage
vigil manpage:generate
vigil manpage:install
```

Narrator:

That keeps help, command listing, completion, documentation, and execution policy aligned.

## Performance And Bounded Execution

[Direction: Show timeouts, cancellation, and bounded logs. Use a process tree animation for cancellation.]

Narrator:

Vigil runs commands through a bounded process runner.

Each process has a default or per-check timeout. Ctrl+C and SIGTERM cancel the active process group. Plugin handshakes time out. Plugin command timeouts are capped. Output and artifact sizes are bounded.

```json
{
  "name": "slow integration test",
  "command": "go",
  "args": ["test", "./integration/..."],
  "read_only": true,
  "timeout": "10m"
}
```

Narrator:

Vigil also includes performance scripts for startup and release checks.

```bash
scripts/check-performance.sh
go run ./scripts/performance-check.go
```

## Release And Distribution Features

[Direction: Show a release archive being assembled: binary, README, LICENSE, manpage, completions, schemas, checksums, SBOM, Sigstore bundle, attestation, v1 acceptance gate JSON.]

Narrator:

Vigil's release workflow is designed for static macOS, Linux, and Windows archives for amd64 and arm64.

Release candidates can include the binary, README, license, manpage, generated shell completions, public schemas, checksums, an SPDX SBOM, a Sigstore bundle, GitHub build-provenance attestations, and `v1-acceptance-gate.json`.

macOS binaries are Developer ID-signed and notarized before final checksums are created.

Verification can include checksums, GitHub attestations, and Sigstore verification.

```bash
shasum -a 256 -c SHA256SUMS
gh attestation verify vigil_0.2.0-beta.1_linux_amd64.tar.gz \
  --repo PayCal-Technologies/vigil-public
cosign verify-blob SHA256SUMS \
  --bundle SHA256SUMS.sigstore.json \
  --certificate-identity \
    https://github.com/PayCal-Technologies/vigil-public/.github/workflows/release.yml@refs/tags/v0.2.0-beta.1 \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```

Narrator:

Homebrew packaging automation exists, but the project does not advertise Homebrew installation until the public tap formula is tested and published.

## Acceptance Evidence

[Direction: Show `docs/v1-acceptance.json` linked to tests, schemas, scripts, docs, operational reports, and external reports.]

Narrator:

Vigil's v1 release gate is evidence-backed.

The canonical ledger is `docs/v1-acceptance.json`. It records required safety invariants and acceptance criteria, and each verified item must cite real evidence.

Evidence can be Go tests, documents, schemas, scripts, workflows, operational reports, or external reports.

```text
docs/v1-acceptance.json
        |
        v
Go tests, schemas, scripts, documents,
operational reports, and external review reports
```

Narrator:

Pre-v1 and prerelease builds can proceed while operational or external criteria remain pending. Stable v1 releases fail closed until every required criterion is verified.

## Troubleshooting Flow

[Direction: Show a "start here" diagnostic checklist. Each command appears as a terminal line.]

Narrator:

When something goes wrong, start with read-only diagnostics.

```bash
vigil version
vigil status --json
vigil doctor --json
vigil config:validate --json
vigil extensions:doctor --json
vigil plugins:doctor --json
```

Narrator:

If config is rejected, preview migration and inspect structured issue codes.

```bash
vigil config:migrate --json
```

If a reviewed plan no longer matches, generate and review a new plan.

If a read-only check changed files, inspect the private run directory, stdout and stderr logs, optional diffs, and artifact paths.

If a required executable is missing, install the tool or intentionally mark the gate optional.

If a plugin is blocked, inspect the plugin list, doctor output, publisher state, digest, metadata, capabilities, revocations, and repository policy.

## Full Feature Summary

[Direction: Use this as a chaptered recap. Keep the visuals moving through a grid of features.]

Narrator:

Vigil provides a versioned JSON configuration model with validation, migration, repair, templates, and redacted reporting.

It provides beginner commands for setup, checking, status, learning, and explanation.

It provides an advanced CLI with typed command discovery, help, completions, manpages, workflow execution, reviewed plans, verification, hooks, support bundles, plugins, packs, and diagnostics.

It executes commands directly through argv by default, with explicit shell opt-in.

It supports process timeouts, cancellation, signal handling, stable execution states, and bounded logs.

It detects when read-only workflow gates change Git-visible tracked or untracked project files.

It supports workflow DAGs with dependencies, deterministic ordering, tag selection, bounded read-only parallelism, continue-on-error behavior, optional required tools, narrow network retries, working directories, environment values, and declared artifacts.

It supports reviewed plans that bind execution to the reviewed binary, config, repository state, command registry, pack registry, and workspace digest.

It emits structured machine output through stable JSON envelopes, JSONL events, JUnit, SARIF, and GitHub annotation formats.

It has a fixed exit-code taxonomy for success, check failure, invalid usage or config, policy block, missing dependency, interruption, mutation violation, and internal failure.

It includes safe Git hook management for pre-commit and pre-push checks.

It includes local readiness and repository hygiene checks, including staged sensitive file scanning, tracked assistant artifact detection, dependency inventory, public assumption checks, public parity checks, workspace hygiene checks, and support bundles.

It includes embedded official feature collections for accessibility adapters, dependency adapters, deployment verification, file iteration, GitHub CI helpers, release policy, repository health, README Scribe, security adapters, and test adapters.

It supports declarative packs with deterministic precedence, repository confinement, policy controls, command contracts, and fail-closed validation.

It supports executable plugins through a strict subprocess protocol, digest-bound locks, local trust, publisher trust, signed indexes, capability approval, conformance checks, sanitized execution environments, bounded handshakes, artifact verification, command namespacing, and repository policy enforcement.

It includes release tooling for deterministic archives, checksums, macOS signing and notarization, reproducibility checks, GitHub release readiness, Homebrew formula generation, SBOMs, Sigstore bundles, attestations, and stable v1 acceptance gates.

It maintains public schemas for configuration, output, JSONL events, plans, packs, run artifact manifests, plugin protocol, plugin conformance, plugin indexes, publisher trust, plugin locks, external evidence, operational evidence, acceptance ledgers, and release gates.

And throughout all of those features, it preserves the central boundary: Vigil is a policy and evidence layer. It makes automation reviewable and verifiable. It does not claim to sandbox approved processes.

## Closing

[Direction: Return to the opening repository view. A terminal runs `vigil check`, then `vigil verify --json`. End on a clean structured result.]

Narrator:

Vigil helps people and coding agents run repository automation with clearer contracts.

Before a command runs, Vigil can show what it is, where it came from, what it can access, whether it may write, what policy applies, and what evidence it should produce.

During execution, Vigil bounds the process, handles cancellation, and checks read-only gates against Git-visible mutation evidence.

After execution, Vigil returns stable human and machine-readable results.

And when a plan has been reviewed, Vigil can apply that exact plan later only if the binary, config, repository, command registry, pack registry, and workspace state still match.

That is Vigil's promise: not invisible trust, not a sandbox claim, and not another opaque automation layer.

Vigil gives local project automation a reviewable contract, an approval path, and evidence you can inspect.

## Optional Short Video Cutdown

Internal direction: For a shorter video, keep only these sections: Title Card, Introduction, What Vigil Is, The Core Workflow, Access Classes And Mutation Approval, Repository Mutation Detection, Reviewed Plans, Structured Output, Built-In Feature Collections, Executable Plugins, Example Use Case: AI Coding Agent Handoff, Full Feature Summary, Closing.

Internal direction: In the short version, show only these code snippets:

```bash
vigil setup
vigil explain
vigil check
vigil status
```

```json
{
  "name": "go test",
  "command": "go",
  "args": ["test", "./..."],
  "read_only": true,
  "tags": ["test", "pre-push"],
  "timeout": "5m"
}
```

```bash
vigil --allow-mutation plan --output .vigil/plans/reviewed.json
vigil --allow-mutation apply .vigil/plans/reviewed.json
```

```bash
vigil workflow:local --format=jsonl
vigil verify --json
```

```text
0 success
3 policy blocked
4 dependency missing
6 mutation violation
```

Internal direction: End the short video with the phrase: "Vigil gives local project automation a reviewable contract, an approval path, and evidence you can inspect."
