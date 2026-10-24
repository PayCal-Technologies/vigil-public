package config

import "testing"

func TestGateRejectsInvalidToolRequirementsAndTimeout(t *testing.T) {
	for _, gate := range []Gate{
		{Name: "x", Command: "git", ReadOnly: true, RequiredEnv: []string{"BAD=NAME"}},
		{Name: "x", Command: "git", ReadOnly: true, ExecutableSHA256: "wrong"},
		{Name: "x", Command: "git", ReadOnly: true, VersionArgs: []string{"--version"}},
		{Name: "x", Command: "git", ReadOnly: true, VersionArgs: []string{"--version"}, VersionPattern: "["},
		{Name: "x", Command: "git", Shell: true, VersionArgs: []string{"--version"}, VersionPattern: "git"},
		{Name: "x", Command: "git", ReadOnly: true, Timeout: "25h"},
	} {
		if len(GateIssues([]Gate{gate})) == 0 {
			t.Fatalf("invalid requirements accepted: %+v", gate)
		}
	}
}
