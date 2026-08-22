package cleanliness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsAndDeletesDisposableArtifacts(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".phpunit.result.cache", "cache")
	writeFile(t, root, "strings/en.txt.bak", "backup")
	writeFile(t, root, "src/change.patch.rej", "reject")
	writeFile(t, root, "tmp/vigil-gates.hooks-pre-commit.123/gate.log", "ok")
	writeFile(t, root, "test-results/report.xml", "<testsuite/>")
	writeFile(t, root, ".git/keep.tmp", "git")
	writeFile(t, root, "logs/app.log", "log")
	writeFile(t, root, "storage/app.tmp", "runtime")
	writeFile(t, root, "vendor/package/.phpunit.result.cache", "dependency")

	report, err := Scan(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	paths := findingPaths(report)
	for _, want := range []string{
		".phpunit.result.cache",
		"strings/en.txt.bak",
		"src/change.patch.rej",
		"tmp/vigil-gates.hooks-pre-commit.123",
		"test-results",
	} {
		if !paths[want] {
			t.Fatalf("missing cleanup finding %q in %#v", want, report.Findings)
		}
	}
	for _, blocked := range []string{".git/keep.tmp", "logs/app.log", "storage/app.tmp", "vendor/package/.phpunit.result.cache"} {
		if paths[blocked] {
			t.Fatalf("unsafe cleanup finding %q in %#v", blocked, report.Findings)
		}
	}

	report, err = Scan(root, Options{Delete: true})
	if err != nil {
		t.Fatal(err)
	}
	if report.DeletedCount == 0 {
		t.Fatal("expected deleted artifacts")
	}
	for _, gone := range []string{".phpunit.result.cache", "strings/en.txt.bak", "tmp/vigil-gates.hooks-pre-commit.123"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(gone))); !os.IsNotExist(err) {
			t.Fatalf("%s still exists after cleanup", gone)
		}
	}
	for _, kept := range []string{".git/keep.tmp", "logs/app.log", "storage/app.tmp", "vendor/package/.phpunit.result.cache"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(kept))); err != nil {
			t.Fatalf("%s should have been kept: %v", kept, err)
		}
	}
}

func TestScanIncludesLogsAndVendorOnlyWhenRequested(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "logs/app.log", "log")
	writeFile(t, root, "node_modules/pkg/index.js", "js")
	writeFile(t, root, "vendor/pkg/index.php", "php")

	report, err := Scan(root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("default scan found protected dirs: %#v", report.Findings)
	}

	report, err = Scan(root, Options{IncludeLogs: true, IncludeVendor: true})
	if err != nil {
		t.Fatal(err)
	}
	paths := findingPaths(report)
	for _, want := range []string{"logs", "node_modules", "vendor"} {
		if !paths[want] {
			t.Fatalf("missing requested finding %q in %#v", want, report.Findings)
		}
	}
}

func TestScanRejectsNonDirectoryRoot(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "fixture.txt")
	if err := os.WriteFile(file, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Scan(file, Options{}); err == nil {
		t.Fatal("expected non-directory root to fail")
	}
}

func findingPaths(report Report) map[string]bool {
	paths := map[string]bool{}
	for _, finding := range report.Findings {
		paths[finding.Path] = true
	}
	return paths
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
