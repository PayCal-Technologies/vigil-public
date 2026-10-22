# GitHub Actions Helper Extension

This is the canonical public GitHub extension for running vigil preflight checks
inside GitHub Actions.

```bash
vigil github:init-ci
vigil github:init-ci --write
```

The generated workflow installs vigil, runs `vigil verify --json`, and then
runs the configured checks from `vigil.config.json`.
