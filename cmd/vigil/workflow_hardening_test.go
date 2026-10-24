package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	vigilcli "github.com/PayCal-Technologies/vigil-public/internal/cli"
	vigilconfig "github.com/PayCal-Technologies/vigil-public/internal/config"
	vigilplan "github.com/PayCal-Technologies/vigil-public/internal/plan"
	"github.com/PayCal-Technologies/vigil-public/internal/runner"
)

func TestMissingOptionalToolCannotSatisfyRequiredDependency(t *testing.T) {
	optional := false
	cfg := templateConfig("generic")
	cfg.Gates = []gateConfig{
		{Name: "optional", Command: "vigil-nonexistent-fixture-tool", ReadOnly: true, Required: &optional},
		{Name: "required", Command: "true", ReadOnly: true, DependsOn: []string{"optional"}},
		{Name: "independent", Command: "true", ReadOnly: true},
	}
	setupWorkflowGraphRepository(t, cfg)
	code := 0
	out := captureStdout(t, func() { code = workflowLocalContext(context.Background(), "", []string{"--json"}, false) })
	var payload struct {
		Results []gateResult `json:"results"`
	}
	decodeEnvelopeData(t, []byte(out), &payload)
	if code != vigilcli.ExitDependencyMissing || len(payload.Results) != 3 || payload.Results[1].Attempts != 0 || payload.Results[1].ExitCode != vigilcli.ExitDependencyMissing || payload.Results[2].State != string(runner.StateOK) {
		t.Fatalf("exit=%d results=%+v", code, payload.Results)
	}
}
func TestRetryBudgetIncludesBackoff(t *testing.T) {
	start := time.Now()
	result, attempts, _ := runGateAttempts(context.Background(), runner.Spec{Executable: "sh", Args: []string{"-c", "exit 1"}, Timeout: 80 * time.Millisecond}, gateConfig{ReadOnly: true, Tags: []string{"network"}, Retry: &vigilconfig.GateRetry{MaxAttempts: 5, Delay: "1s"}})
	if result.State != runner.StateTimedOut || attempts != 1 || time.Since(start) > time.Second {
		t.Fatalf("unbounded retries: result=%+v attempts=%d", result, attempts)
	}
}
func TestGateToolAndEnvironmentRequirements(t *testing.T) {
	root := t.TempDir()
	path, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := vigilplan.DigestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest = strings.TrimPrefix(digest, "sha256:")
	gate := gateConfig{Command: path, RequiredEnv: []string{"VIGIL_REQUIRED_TEST_ENV"}, ExecutableSHA256: digest, VersionArgs: []string{"--version"}, VersionPattern: "^git version "}
	t.Setenv("VIGIL_REQUIRED_TEST_ENV", "")
	if _, err = checkGateRequirements(context.Background(), gate, root, nil); err == nil {
		t.Fatal("missing environment accepted")
	}
	resolved, err := checkGateRequirements(context.Background(), gate, root, []string{"VIGIL_REQUIRED_TEST_ENV=privatevalue"})
	if err != nil || resolved != path {
		t.Fatalf("requirements: %s %v", resolved, err)
	}
	gate.ExecutableSHA256 = strings.Repeat("0", 64)
	if _, err = checkGateRequirements(context.Background(), gate, root, []string{"VIGIL_REQUIRED_TEST_ENV=privatevalue"}); err == nil || strings.Contains(err.Error(), "privatevalue") {
		t.Fatalf("bad digest not rejected safely: %v", err)
	}
	gate.ExecutableSHA256 = digest
	gate.VersionPattern = "^never-matches$"
	if _, err = checkGateRequirements(context.Background(), gate, root, []string{"VIGIL_REQUIRED_TEST_ENV=privatevalue"}); err == nil {
		t.Fatal("version mismatch accepted")
	}
}
func TestRequiredEnvironmentFailurePreventsCommandStart(t *testing.T) {
	t.Setenv("VIGIL_MISSING_FIXTURE_ENV", "")
	cfg := templateConfig("generic")
	cfg.Gates = []gateConfig{{Name: "blocked", Command: "mkdir -p .vigil; touch .vigil/must-not-run", Shell: true, ReadOnly: true, RequiredEnv: []string{"VIGIL_MISSING_FIXTURE_ENV"}}}
	setupWorkflowGraphRepository(t, cfg)
	code := 0
	_ = captureStdout(t, func() { code = workflowLocalContext(context.Background(), "", []string{"--json"}, false) })
	if code != vigilcli.ExitPolicyBlocked {
		t.Fatalf("exit %d", code)
	}
	if _, err := os.Stat(".vigil/must-not-run"); !os.IsNotExist(err) {
		t.Fatal("command ran despite unmet requirement")
	}
}

func TestEmptyWorkflowDoesNotPass(t *testing.T) {
	cfg := templateConfig("generic")
	cfg.Gates = []gateConfig{}
	setupWorkflowGraphRepository(t, cfg)
	code := 0
	out := captureStdout(t, func() { code = workflowLocalContext(context.Background(), "", []string{"--json"}, false) })
	if code != vigilcli.ExitUsage || !strings.Contains(out, "at least one") {
		t.Fatalf("empty workflow: %d %s", code, out)
	}
}

func TestEmptyTagSelectionDoesNotPass(t *testing.T) {
	setupWorkflowGraphRepository(t, templateConfig("generic"))
	code := 0
	out := captureStdout(t, func() {
		code = workflowLocalContext(context.Background(), "", []string{"--json", "--tag", "no-such-tag"}, false)
	})
	if code != vigilcli.ExitPolicyBlocked || !strings.Contains(out, "no selected gates") {
		t.Fatalf("empty tag selection: %d %s", code, out)
	}
}
