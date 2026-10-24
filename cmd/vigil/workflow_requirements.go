package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PayCal-Technologies/vigil-public/internal/runner"
)

// Requirements are part of the reviewed gate plan. Probes inherit its directory,
// environment and deadline, and neither acquire nor install missing tools.
func checkGateRequirements(ctx context.Context, gate gateConfig, dir string, env []string) (string, error) {
	values := map[string]string{}
	for _, entry := range append(os.Environ(), env...) {
		if k, v, ok := strings.Cut(entry, "="); ok {
			values[k] = v
		}
	}
	for _, key := range gate.RequiredEnv {
		if strings.TrimSpace(values[key]) == "" {
			return "", fmt.Errorf("required environment variable %s is unavailable", key)
		}
	}
	if gate.ExecutableSHA256 == "" && gate.VersionPattern == "" {
		return "", nil
	}
	if gate.Shell {
		return "", fmt.Errorf("tool requirements require argv execution")
	}
	path := gate.Command
	if strings.ContainsAny(path, "/\\") && !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	path, err := exec.LookPath(path)
	if err != nil {
		return "", fmt.Errorf("required tool %s unavailable: %w", gate.Command, err)
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	verify := func() error {
		if gate.ExecutableSHA256 == "" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || info.Size() > 512<<20 {
			return fmt.Errorf("tool executable is not a regular file within 512 MiB")
		}
		h := sha256.New()
		n, err := io.Copy(h, io.LimitReader(f, (512<<20)+1))
		if err != nil {
			return err
		}
		if n > 512<<20 || hex.EncodeToString(h.Sum(nil)) != gate.ExecutableSHA256 {
			return fmt.Errorf("tool executable digest does not match the reviewed gate requirement")
		}
		return nil
	}
	if err := verify(); err != nil {
		return "", err
	}
	if gate.VersionPattern != "" {
		pattern, err := regexp.Compile(gate.VersionPattern)
		if err != nil {
			return "", err
		}
		r := runner.Run(ctx, runner.Spec{Name: "tool version preflight", Executable: path, Args: gate.VersionArgs, Dir: dir, Env: env, Timeout: 5 * time.Second, CaptureLimit: 65536})
		if r.State != runner.StateOK || r.Truncated {
			return "", fmt.Errorf("tool version probe failed (%s); output must fit within 64 KiB", r.State)
		}
		if !pattern.MatchString(r.Output) {
			return "", fmt.Errorf("tool version does not meet the reviewed gate requirement")
		}
	}
	if err := verify(); err != nil {
		return "", err
	}
	return path, nil
}
