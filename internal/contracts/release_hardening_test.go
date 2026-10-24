package contracts

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestReleaseBuilderRejectsUnboundSourceAndStableOverrides(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release builder uses bash")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	git := exec.Command("git", "rev-parse", "HEAD~1")
	git.Dir = root
	b, err := git.Output()
	if err != nil {
		t.Skip("fixture requires parent commit")
	}
	for _, tc := range []struct{ tag, commit, want string }{
		{"v0.2.0-beta.99", strings.TrimSpace(string(b)), "must match the source checkout HEAD"},
		{"v1.0.0", "", "stable releases do not allow"},
	} {
		cmd := exec.Command("bash", "scripts/build-release.sh")
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "RELEASE_TAG="+tc.tag, "VERSION="+strings.TrimPrefix(tc.tag, "v"), "COMMIT="+tc.commit, "ALLOW_DIRTY_RELEASE=1", "ALLOW_UNTAGGED_RELEASE=1")
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), tc.want) {
			t.Fatalf("release guard: %v %s", err, out)
		}
	}
}

func TestBlockedReleasePreservesPreviousArtifacts(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("release builder uses bash")
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	temporary := t.TempDir()
	fakeBin := filepath.Join(temporary, "bin")
	dist := filepath.Join(temporary, "dist")
	for _, dir := range []string{fakeBin, dist} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	sentinel := filepath.Join(dist, "previous-release")
	if err := os.WriteFile(sentinel, []byte("preserve me"), 0600); err != nil {
		t.Fatal(err)
	}
	fakeGo := `#!/bin/sh
output=""
while [ "$#" -gt 0 ]; do
 if [ "$1" = "-o" ]; then shift; output="$1"; fi
 shift
done
[ -n "$output" ] || exit 9
printf '#!/bin/sh\nprintf "blocked acceptance fixture\\n"\nexit 1\n' > "$output"
chmod +x "$output"
`
	if err := os.WriteFile(filepath.Join(fakeBin, "go"), []byte(fakeGo), 0700); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "scripts/build-release.sh")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"), "RELEASE_TAG=v0.2.0-beta.99", "VERSION=0.2.0-beta.99", "COMMIT=", "ALLOW_DIRTY_RELEASE=1", "ALLOW_UNTAGGED_RELEASE=1", "DIST_DIR="+dist)
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "blocked acceptance fixture") {
		t.Fatalf("blocked builder: %v %s", err, out)
	}
	b, err := os.ReadFile(sentinel)
	if err != nil || string(b) != "preserve me" {
		t.Fatal("failed build destroyed previous artifacts")
	}
	if _, err := os.Stat(dist + ".build.lock"); !os.IsNotExist(err) {
		t.Fatal("failed build left a writer lock")
	}
}
