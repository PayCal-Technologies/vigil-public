package cleanliness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Finding struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Bytes  int64  `json:"bytes"`
	Reason string `json:"reason"`
}

type Report struct {
	Status       string    `json:"status"`
	Deleted      bool      `json:"deleted"`
	Findings     []Finding `json:"findings"`
	DeletedCount int       `json:"deleted_count,omitempty"`
	DeletedBytes int64     `json:"deleted_bytes,omitempty"`
}

type Options struct {
	Delete        bool
	IncludeLogs   bool
	IncludeVendor bool
}

func Scan(repoRoot string, opts Options) (Report, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Report{}, err
	}
	if info, err := os.Stat(root); err != nil {
		return Report{}, err
	} else if !info.IsDir() {
		return Report{}, fmt.Errorf("repository root is not a directory: %s", repoRoot)
	}

	findings, err := walkTrash(root)
	if err != nil {
		return Report{}, err
	}
	for _, spec := range specs(opts) {
		matches, err := filepath.Glob(filepath.Join(root, filepath.FromSlash(spec.Pattern)))
		if err != nil {
			return Report{}, err
		}
		for _, match := range matches {
			if !safeCleanupTarget(root, match) {
				continue
			}
			info, err := os.Lstat(match)
			if err != nil {
				continue
			}
			if spec.DirOnly && !info.IsDir() {
				continue
			}
			if spec.FileOnly && info.IsDir() {
				continue
			}
			rel, err := filepath.Rel(root, match)
			if err != nil {
				continue
			}
			findings = append(findings, Finding{
				Path:   filepath.ToSlash(rel),
				Kind:   kind(info),
				Bytes:  size(match, info),
				Reason: spec.Reason,
			})
		}
	}
	findings = unique(findings)

	report := Report{Status: "clean", Findings: findings}
	if len(findings) > 0 {
		report.Status = "dirty"
	}
	if opts.Delete {
		for _, finding := range findings {
			target := filepath.Join(root, filepath.FromSlash(finding.Path))
			if !safeCleanupTarget(root, target) {
				return report, fmt.Errorf("refusing to delete unsafe target %s", finding.Path)
			}
			if err := os.RemoveAll(target); err != nil {
				return report, fmt.Errorf("delete %s: %w", finding.Path, err)
			}
			report.DeletedCount++
			report.DeletedBytes += finding.Bytes
		}
		report.Deleted = true
		if report.DeletedCount > 0 {
			report.Status = "cleaned"
		}
	}
	return report, nil
}

func JSON(report Report) string {
	data, _ := json.MarshalIndent(report, "", "  ")
	return string(data) + "\n"
}

type cleanupSpec struct {
	Pattern  string
	Reason   string
	DirOnly  bool
	FileOnly bool
}

func specs(opts Options) []cleanupSpec {
	out := []cleanupSpec{
		{Pattern: ".DS_Store", Reason: "macOS Finder metadata", FileOnly: true},
		{Pattern: "Thumbs.db", Reason: "Windows Explorer metadata", FileOnly: true},
		{Pattern: ".phpunit.result.cache", Reason: "PHPUnit result cache", FileOnly: true},
		{Pattern: ".php-cs-fixer.cache", Reason: "PHP-CS-Fixer cache", FileOnly: true},
		{Pattern: "*.bak", Reason: "editor backup file", FileOnly: true},
		{Pattern: "*.tmp", Reason: "temporary file", FileOnly: true},
		{Pattern: "*.orig", Reason: "patch merge backup file", FileOnly: true},
		{Pattern: "*.rej", Reason: "patch reject file", FileOnly: true},
		{Pattern: "*~", Reason: "editor backup file", FileOnly: true},
		{Pattern: "npm-debug.log*", Reason: "npm debug log", FileOnly: true},
		{Pattern: "yarn-error.log*", Reason: "package manager debug log", FileOnly: true},
		{Pattern: "pnpm-debug.log*", Reason: "package manager debug log", FileOnly: true},
		{Pattern: "tmp/vigil-gates.*", Reason: "Vigil parallel gate scratch output", DirOnly: true},
		{Pattern: "tmp/vigil-cache", Reason: "Vigil local cache", DirOnly: true},
		{Pattern: "tmp/vigil-stamps", Reason: "Vigil local stamps", DirOnly: true},
		{Pattern: "tmp/vigil-a11y", Reason: "Vigil accessibility scratch output", DirOnly: true},
		{Pattern: "tmp/ci-timings", Reason: "Vigil timing scratch output", DirOnly: true},
		{Pattern: "tmp/test-history", Reason: "Vigil test history scratch output", DirOnly: true},
		{Pattern: "test-results", Reason: "local test result directory", DirOnly: true},
	}
	if opts.IncludeLogs {
		out = append(out, cleanupSpec{Pattern: "logs", Reason: "local log directory", DirOnly: true})
	}
	if opts.IncludeVendor {
		out = append(out,
			cleanupSpec{Pattern: "vendor", Reason: "dependency vendor directory", DirOnly: true},
			cleanupSpec{Pattern: "node_modules", Reason: "npm dependency directory", DirOnly: true},
		)
	}
	return out
}

func walkTrash(repoRoot string) ([]Finding, error) {
	findings := make([]Finding, 0)
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil || rel == "." {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "node_modules", "vendor", "logs", "storage", ".vigil":
				return filepath.SkipDir
			}
			return nil
		}
		reason := trashFileReason(entry.Name())
		if reason == "" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		findings = append(findings, Finding{
			Path:   rel,
			Kind:   "file",
			Bytes:  info.Size(),
			Reason: reason,
		})
		return nil
	})
	return findings, err
}

func trashFileReason(name string) string {
	switch name {
	case ".DS_Store":
		return "macOS Finder metadata"
	case "Thumbs.db":
		return "Windows Explorer metadata"
	case ".phpunit.result.cache":
		return "PHPUnit result cache"
	case ".php-cs-fixer.cache":
		return "PHP-CS-Fixer cache"
	}
	if strings.HasSuffix(name, ".bak") || strings.HasSuffix(name, ".tmp") || strings.HasSuffix(name, ".orig") || strings.HasSuffix(name, ".rej") || strings.HasSuffix(name, "~") {
		return "editor, temporary, or patch backup file"
	}
	if strings.HasPrefix(name, "npm-debug.log") || strings.HasPrefix(name, "yarn-error.log") || strings.HasPrefix(name, "pnpm-debug.log") {
		return "package manager debug log"
	}
	return ""
}

func safeCleanupTarget(repoRoot, target string) bool {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return false
	}
	candidate, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	rel = filepath.ToSlash(rel)
	return rel != ".git" && !strings.HasPrefix(rel, ".git/")
}

func kind(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}

func size(path string, info os.FileInfo) int64 {
	if !info.IsDir() {
		return info.Size()
	}
	var total int64
	_ = filepath.WalkDir(path, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if fileInfo, err := entry.Info(); err == nil {
			total += fileInfo.Size()
		}
		return nil
	})
	return total
}

func unique(findings []Finding) []Finding {
	seen := map[string]Finding{}
	for _, finding := range findings {
		seen[finding.Path] = finding
	}
	out := make([]Finding, 0, len(seen))
	for _, finding := range seen {
		out = append(out, finding)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
