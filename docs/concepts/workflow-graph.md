# Workflow Graph

vigil schema 3 gates form a validated directed acyclic graph. Reviewed plans
copy the complete graph and bind it into the plan digest.

For the full safety flow, including command access, mutation confirmation, and
known non-sandbox limits, see [Safety Model](safety-model.md).

## Ordering

- `depends_on` names prerequisite gates.
- Declaration order is the deterministic tie-breaker for ready gates.
- Gates without `parallel_group` run one at a time, preserving the behavior of
  existing configurations.
- Ready, read-only gates with the same non-empty `parallel_group` may run
  together. `--jobs` bounds the batch from 1 through 32 and defaults to 4.
- Mutating gates cannot declare `parallel_group` and always run exclusively.
- `--tag` selects matching gates plus their transitive dependencies.

Results are returned in declaration order even when completion order differs.
JSONL start and finish events are emitted in deterministic batch order.

## Failure

The default remains fail-fast. `continue_on_error: true` allows independent
gates to continue after an ordinary failure, missing required tool, or timeout.
Dependents of a failed gate are reported as `skipped`; unrelated ready gates
can still run. Policy blocks, cancellation, internal failures, and mutation
violations always halt scheduling.
If vigil detects an impossible scheduler transition, it returns an internal
failure instead of panicking.

`required` defaults to `true`. Setting it to `false` changes only missing-tool
behavior: an unavailable executable becomes a successful `skipped` result so
independent work may proceed. Dependents require an actual success; a required
dependent that cannot run fails the workflow. It does not turn command failures
into successes.

## Retry

Retries are intentionally narrow:

- the gate must be read-only and carry the explicit `network` tag;
- `max_attempts` is between 2 and 5;
- `on` may contain `failed` and `timed_out`;
- `delay` is at most five minutes;
- blocked, cancelled, missing-tool, mutation, and internal states are never
  retried.

The gate timeout applies to preflight, all attempts and retry delays as one
budget. The reported duration includes retry delay. Private stdout and stderr logs retain up to 64 MiB per stream and append
`[truncated]` when the bound is reached. A run shares a 512 MiB log budget and
fails closed if another stream cannot reserve its marker; result summaries
expose truncation flags rather than implying that bounded evidence is complete.

## Execution Context

`cwd` is relative to the repository root. vigil resolves symlinks before
execution and blocks a directory that escapes the repository.

In-repository plan and run-artifact paths must be Git-ignored. vigil resolves
their nearest existing ancestor, rejects symlink redirection outside the
repository, and never writes inside Git metadata. Explicit paths that are
already outside the repository remain supported.

`environment` contains literal repository-owned values. Variable names use the
portable `[A-Za-z_][A-Za-z0-9_]*` form. Repository configuration cannot
override `VIGIL_*`; vigil supplies:

```text
VIGIL_GATE_NAME
VIGIL_PLAN_ID
VIGIL_REPOSITORY_ROOT
VIGIL_RUN_ID
```

The process otherwise inherits the invoking environment. Secrets should not be
stored in repository configuration.

## Declared Artifacts

Each artifact declaration has a repository-confined path relative to the gate
`cwd`, an optional kind and media type, and `required` semantics that default to
true. After execution vigil requires each mandatory artifact to be a regular,
non-symlink file beneath the repository, computes its SHA-256 digest, and adds
it to the common output envelope. Missing optional artifacts become warnings.

Artifact declarations verify outputs; they do not authorize mutation. A
read-only gate that changes Git-visible files still fails mutation detection.

## Parallel Mutation Evidence

vigil fingerprints the Git-visible workspace once before and once after an
explicit parallel batch. If the fingerprint changes, every gate in that batch
is marked `mutation_detected` because vigil cannot safely attribute the change
to one process. The diagnostic reports that a command in the named parallel
group changed the fingerprint and that individual attribution is unavailable.
This is fail-closed and preserves exit code `6`.

`read_only` remains repository verification, not an operating-system sandbox.
Ignored files, caches, network state, and external services are outside the
fingerprint boundary.


## Required prerequisites and time budgets

A workflow with no selected gates fails, including a tag filter that matches no
checks. Missing required dependencies are failures in JUnit reports as well as
in the command exit status.

An optional gate whose executable is missing remains a successful advisory skip
when nothing depends on it. A skipped gate cannot satisfy `depends_on`. Its
transitive dependents are skipped without execution, and an unmet required
dependent makes the workflow fail. An earlier failure keeps its original exit
classification. Independent work may continue according to the existing policy.

Gates may declare these optional requirements. They are included in reviewed
plans and their digests:

```json
{
  "name": "remote-check",
  "command": "audit-tool",
  "args": ["check"],
  "read_only": true,
  "tags": ["network"],
  "required_env": ["AUDIT_TOKEN"],
  "version_args": ["--version"],
  "version_pattern": "^audit-tool 2\\.4\\.[0-9]+\\n?$",
  "timeout": "60s",
  "retry": {"max_attempts": 2, "delay": "2s"}
}
```

`required_env` checks the effective gate environment for nonempty values without
printing them. `version_args` and `version_pattern` must be supplied together for
argv commands. Version probes have a five second limit and at most 64 KiB of
captured output. `executable_sha256`, when supplied, must contain the exact 64
lowercase hexadecimal SHA256 characters for the executable on this platform.
The executable is resolved to a fixed path and checked before execution and
again before retries. A missing requirement blocks the gate. Vigil does not
install tools, acquire credentials or update a dependency to satisfy it.

The gate timeout covers preflight, every attempt and retry delays. A retry does
not reset that budget. Direct runner calls without a timeout use five minutes;
timeouts over 24 hours and capture limits over 64 MiB are rejected. Output summary
truncation remains explicitly reported and does not mean the command failed.
Credential existence is not evidence that a provider is healthy or authorized;
use a bounded check against the actual service to establish that requirement.
