# vigil GitHub Visibility Action Plan

This plan tracks repository changes that make vigil easier to discover,
evaluate, trust, and share.

## Current Baseline

| Area | Status |
| --- | --- |
| Public repository | Complete |
| README with badges, install, screenshots, and official links | Complete |
| License | Complete |
| Security policy | Complete |
| Contributing guide | Complete |
| Code of conduct | Complete |
| Issue and PR templates | Complete |
| Detailed docs and help system | Complete |
| Testing guide | Complete |
| Examples directory | Present; indexed in this visibility pass |
| Social preview asset | Added in this visibility pass |
| GitHub topics/homepage | Update through GitHub metadata |

## Priority Actions

1. Keep the README opening focused on vigil's value: local preflight checks,
   explicit mutation boundaries, evidence, and agent-safe automation.
2. Keep release status unambiguous. The current public prerelease, legacy
   bootstrap release, and v1 gate should stay easy to distinguish.
3. Keep `examples/` copy-pasteable and linked from README for first-time users,
   CI authors, and AI/coding agents.
4. Keep `help/` and `docs/` cross-linked with the official PCT help site.
5. Set repository metadata:
   - Homepage: `https://paycaltech.com/vigil/`
   - Topics: `go`, `cli`, `devtools`, `preflight`, `automation`,
     `agent-safety`, `local-first`, `ci`, `security-tools`, `github-actions`
6. Upload `.github/social-preview.png` in GitHub repository settings as the
   social preview image.
7. Keep releases rich: changed commands, compatibility notes, verification,
   checksums/provenance status, and links to the official help site.

## Recurring Release Checklist

- [ ] README install/version text matches the release.
- [ ] `CHANGELOG.md` and GitHub release notes explain user-visible changes.
- [ ] `TESTING.md` reflects current test coverage.
- [ ] `examples/` still matches current commands and config schema.
- [ ] Help HTML and screenshots reflect current command behavior.
- [ ] GitHub topics and homepage still match the project.
- [ ] Social preview remains current with the public positioning.
