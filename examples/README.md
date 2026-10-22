# vigil Examples

These examples show common ways to use vigil as a local safety and evidence
layer before repository automation runs.

## Starter Config

- [`vigil.config.json`](vigil.config.json) is a copyable local configuration
  example.
- [`workflow-graph.config.json`](workflow-graph.config.json) demonstrates a
  workflow graph configuration shape.

Validate a config:

```bash
vigil config:validate
vigil config:report --json
```

## GitHub Actions

[`github-vigil.yml`](github-vigil.yml) shows a CI workflow that installs and
runs vigil checks.

Typical local workflow:

```bash
vigil workflow:local --dry-run
vigil workflow:local
vigil verify --json
```

## Plugin Reference

[`plugins/reference/`](plugins/reference/) contains a minimal executable plugin
reference for testing plugin protocol behavior and trust boundaries.

Relevant commands:

```bash
vigil plugins:conformance ./examples/plugins/reference/vigil-plugin-reference
vigil plugins:doctor
vigil plugins:list
```

## AI and Automation Context

Use machine-readable output when a script or coding agent needs stable context:

```bash
vigil help --json
vigil status --json
vigil verify --json
vigil guards:summary --json
```

Mutation commands should remain explicit and reviewable:

```bash
vigil --allow-mutation plan --output .vigil/plans/reviewed.json
vigil --allow-mutation apply .vigil/plans/reviewed.json
```
