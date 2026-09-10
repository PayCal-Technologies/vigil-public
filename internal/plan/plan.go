package plan

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PayCal-Technologies/vigil-public/internal/atomicfile"
	"github.com/PayCal-Technologies/vigil-public/internal/config"
)

const (
	SchemaVersion = "1"
	MaxFileBytes  = 8 << 20
)

type Inputs struct {
	BinaryDigest          string `json:"binary_digest"`
	ConfigPath            string `json:"config_path"`
	ConfigDigest          string `json:"config_digest"`
	RepositoryRoot        string `json:"repository_root"`
	RepositoryHead        string `json:"repository_head"`
	WorkspaceDigest       string `json:"workspace_digest"`
	CommandRegistryDigest string `json:"command_registry_digest"`
	PackDigest            string `json:"pack_digest"`
}

type Options struct {
	TagFilter      string `json:"tag_filter,omitempty"`
	DefaultTimeout string `json:"default_timeout"`
	MaxParallel    int    `json:"max_parallel,omitempty"`
}

type Document struct {
	SchemaVersion string        `json:"schema_version"`
	PlanID        string        `json:"plan_id,omitempty"`
	Command       string        `json:"command"`
	CreatedAt     string        `json:"created_at"`
	Inputs        Inputs        `json:"inputs"`
	Options       Options       `json:"options"`
	Gates         []config.Gate `json:"gates"`
	Review        *Review       `json:"review,omitempty"`
}

type Review struct {
	Mode                string             `json:"mode"`
	ReviewedAt          string             `json:"reviewed_at"`
	AuthorizationDigest string             `json:"authorization_digest,omitempty"`
	MutationApprovals   []MutationApproval `json:"mutation_approvals,omitempty"`
}

type MutationApproval struct {
	Gate    string `json:"gate"`
	Command string `json:"command"`
}

type Mismatch struct {
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

var digestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func New(command string, createdAt time.Time, inputs Inputs, options Options, gates []config.Gate) (Document, error) {
	document := Document{
		SchemaVersion: SchemaVersion,
		Command:       strings.TrimSpace(command),
		CreatedAt:     createdAt.UTC().Format(time.RFC3339Nano),
		Inputs:        inputs,
		Options:       options,
		Gates:         cloneGates(gates),
	}
	planID, err := ID(document)
	if err != nil {
		return Document{}, err
	}
	document.PlanID = planID
	if err := Validate(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

func Validate(document Document) error {
	if document.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported plan schema_version %q", document.SchemaVersion)
	}
	if document.Command != "workflow:local" {
		return fmt.Errorf("unsupported planned command %q", document.Command)
	}
	if _, err := time.Parse(time.RFC3339Nano, document.CreatedAt); err != nil {
		return fmt.Errorf("invalid plan created_at: %w", err)
	}
	timeout, err := time.ParseDuration(document.Options.DefaultTimeout)
	if err != nil || timeout <= 0 {
		return fmt.Errorf("invalid plan default_timeout %q", document.Options.DefaultTimeout)
	}
	if document.Options.MaxParallel < 0 || document.Options.MaxParallel > 32 {
		return fmt.Errorf("invalid plan max_parallel %d", document.Options.MaxParallel)
	}
	requiredDigests := map[string]string{
		"binary_digest":           document.Inputs.BinaryDigest,
		"config_digest":           document.Inputs.ConfigDigest,
		"workspace_digest":        document.Inputs.WorkspaceDigest,
		"command_registry_digest": document.Inputs.CommandRegistryDigest,
		"pack_digest":             document.Inputs.PackDigest,
	}
	for field, digest := range requiredDigests {
		if !digestPattern.MatchString(digest) {
			return fmt.Errorf("invalid %s %q", field, digest)
		}
	}
	if strings.TrimSpace(document.Inputs.ConfigPath) == "" {
		return errors.New("plan config_path is required")
	}
	if strings.TrimSpace(document.Inputs.RepositoryRoot) == "" {
		return errors.New("plan repository_root is required")
	}
	if issues := config.GateIssues(document.Gates); len(issues) > 0 {
		return errors.New("invalid plan gates: " + strings.Join(config.IssueMessages(issues), "; "))
	}
	if document.Review != nil {
		if strings.TrimSpace(document.Review.Mode) == "" {
			return errors.New("plan review mode is required")
		}
		if _, err := time.Parse(time.RFC3339Nano, document.Review.ReviewedAt); err != nil {
			return fmt.Errorf("invalid plan review reviewed_at: %w", err)
		}
		gatesByName := map[string]config.Gate{}
		for _, gate := range document.Gates {
			gatesByName[gate.Name] = gate
		}
		seenApprovals := map[string]bool{}
		for _, approval := range document.Review.MutationApprovals {
			name := strings.TrimSpace(approval.Gate)
			if name == "" {
				return errors.New("plan review mutation approval gate is required")
			}
			gate, ok := gatesByName[name]
			if !ok {
				return fmt.Errorf("plan review mutation approval references unknown gate %q", name)
			}
			if gate.ReadOnly {
				return fmt.Errorf("plan review mutation approval references read-only gate %q", name)
			}
			if approval.Command != GateDisplayCommand(gate) {
				return fmt.Errorf("plan review mutation approval command mismatch for gate %q", name)
			}
			if seenApprovals[name] {
				return fmt.Errorf("plan review mutation approval duplicates gate %q", name)
			}
			seenApprovals[name] = true
		}
		if !digestPattern.MatchString(document.Review.AuthorizationDigest) {
			return fmt.Errorf("invalid plan review authorization_digest %q", document.Review.AuthorizationDigest)
		}
		expectedDigest, err := ReviewAuthorizationDigest(document)
		if err != nil {
			return err
		}
		if document.Review.AuthorizationDigest != expectedDigest {
			return fmt.Errorf("plan review authorization_digest mismatch: expected %s", expectedDigest)
		}
	}
	expectedID, err := ID(document)
	if err != nil {
		return err
	}
	if document.PlanID != expectedID {
		return fmt.Errorf("plan_id mismatch: expected %s", expectedID)
	}
	return nil
}

func ID(document Document) (string, error) {
	document.PlanID = ""
	document.Review = nil
	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func AuthorizeReview(document Document) (Document, error) {
	if document.Review == nil {
		return document, errors.New("plan review is required")
	}
	document.Review.AuthorizationDigest = ""
	digest, err := ReviewAuthorizationDigest(document)
	if err != nil {
		return Document{}, err
	}
	document.Review.AuthorizationDigest = digest
	if err := Validate(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

func ReviewAuthorizationDigest(document Document) (string, error) {
	if document.Review == nil {
		return "", errors.New("plan review is required")
	}
	if !digestPattern.MatchString(document.PlanID) {
		return "", fmt.Errorf("invalid plan_id %q", document.PlanID)
	}
	review := *document.Review
	review.AuthorizationDigest = ""
	gatesByName := map[string]config.Gate{}
	for _, gate := range document.Gates {
		gatesByName[gate.Name] = gate
	}
	approvals := make([]reviewAuthorizationApproval, 0, len(review.MutationApprovals))
	for _, approval := range review.MutationApprovals {
		gate, ok := gatesByName[strings.TrimSpace(approval.Gate)]
		if !ok {
			return "", fmt.Errorf("plan review mutation approval references unknown gate %q", strings.TrimSpace(approval.Gate))
		}
		approvals = append(approvals, reviewAuthorizationApproval{
			Gate:       strings.TrimSpace(approval.Gate),
			Command:    approval.Command,
			Structured: gate,
		})
	}
	return DigestJSON(reviewAuthorizationPayload{
		Schema:      "vigil.sentry.authorization.v1",
		PlanID:      document.PlanID,
		Mode:        review.Mode,
		ReviewedAt:  review.ReviewedAt,
		Approvals:   approvals,
		Command:     document.Command,
		CreatedAt:   document.CreatedAt,
		Inputs:      document.Inputs,
		Options:     document.Options,
		GateCount:   len(document.Gates),
		PlanVersion: document.SchemaVersion,
	})
}

type reviewAuthorizationPayload struct {
	Schema      string                        `json:"schema"`
	PlanID      string                        `json:"plan_id"`
	PlanVersion string                        `json:"plan_version"`
	Command     string                        `json:"command"`
	CreatedAt   string                        `json:"created_at"`
	Inputs      Inputs                        `json:"inputs"`
	Options     Options                       `json:"options"`
	GateCount   int                           `json:"gate_count"`
	Mode        string                        `json:"mode"`
	ReviewedAt  string                        `json:"reviewed_at"`
	Approvals   []reviewAuthorizationApproval `json:"approvals"`
}

type reviewAuthorizationApproval struct {
	Gate       string      `json:"gate"`
	Command    string      `json:"command"`
	Structured config.Gate `json:"structured"`
}

func GateDisplayCommand(gate config.Gate) string {
	if gate.Shell {
		return "shell: " + gate.Command
	}
	parts := []string{shellQuote(gate.Command)}
	for _, arg := range gate.Args {
		parts = append(parts, shellQuote(arg))
	}
	return strings.Join(parts, " ")
}

func ApprovedMutationGates(document Document) map[string]bool {
	approved := map[string]bool{}
	if document.Review == nil {
		return approved
	}
	for _, approval := range document.Review.MutationApprovals {
		approved[strings.TrimSpace(approval.Gate)] = true
	}
	return approved
}

func UnapprovedMutationGates(document Document) []string {
	approved := ApprovedMutationGates(document)
	var missing []string
	for _, gate := range document.Gates {
		if gate.ReadOnly || approved[gate.Name] {
			continue
		}
		missing = append(missing, gate.Name)
	}
	sort.Strings(missing)
	return missing
}

func Compare(expected, actual Inputs) []Mismatch {
	fields := []struct {
		name     string
		expected string
		actual   string
	}{
		{"binary_digest", expected.BinaryDigest, actual.BinaryDigest},
		{"config_path", expected.ConfigPath, actual.ConfigPath},
		{"config_digest", expected.ConfigDigest, actual.ConfigDigest},
		{"repository_root", expected.RepositoryRoot, actual.RepositoryRoot},
		{"repository_head", expected.RepositoryHead, actual.RepositoryHead},
		{"workspace_digest", expected.WorkspaceDigest, actual.WorkspaceDigest},
		{"command_registry_digest", expected.CommandRegistryDigest, actual.CommandRegistryDigest},
		{"pack_digest", expected.PackDigest, actual.PackDigest},
	}
	var mismatches []Mismatch
	for _, field := range fields {
		if field.expected != field.actual {
			mismatches = append(mismatches, Mismatch{
				Field:    field.name,
				Expected: field.expected,
				Actual:   field.actual,
			})
		}
	}
	sort.Slice(mismatches, func(i, j int) bool {
		return mismatches[i].Field < mismatches[j].Field
	})
	return mismatches
}

func Write(path string, document Document) error {
	if err := Validate(document); err != nil {
		return err
	}
	data, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = atomicfile.Write(path, data, atomicfile.Options{
		DefaultMode:          0o600,
		PreserveExistingMode: false,
	})
	return err
}

func Read(path string) (Document, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return Document{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return Document{}, fmt.Errorf("plan file must not be a symlink: %s", path)
	}
	if !info.Mode().IsRegular() {
		return Document{}, fmt.Errorf("plan path is not a regular file: %s", path)
	}
	file, err := os.Open(path)
	if err != nil {
		return Document{}, err
	}
	defer file.Close()
	limited := io.LimitReader(file, MaxFileBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return Document{}, err
	}
	if len(data) > MaxFileBytes {
		return Document{}, fmt.Errorf("plan file exceeds %d bytes", MaxFileBytes)
	}
	var document Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return Document{}, err
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Document{}, err
	}
	if err := Validate(document); err != nil {
		return Document{}, err
	}
	return document, nil
}

func DigestFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("digest path is not a regular file: %s", path)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func DigestJSON(value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func DigestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", sum[:])
}

func shellQuote(value string) string {
	if value == "" {
		return "''"
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && !strings.ContainsRune("-_./:", r)
	}) < 0 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func ExecutableDigest() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}
	return DigestFile(resolved)
}

func cloneGates(gates []config.Gate) []config.Gate {
	cloned := make([]config.Gate, len(gates))
	for index, gate := range gates {
		cloned[index] = gate
		cloned[index].Args = append([]string(nil), gate.Args...)
		cloned[index].Tags = append([]string(nil), gate.Tags...)
		cloned[index].DependsOn = append([]string(nil), gate.DependsOn...)
		if gate.Required != nil {
			required := *gate.Required
			cloned[index].Required = &required
		}
		if gate.Retry != nil {
			retry := *gate.Retry
			retry.On = append([]string(nil), gate.Retry.On...)
			cloned[index].Retry = &retry
		}
		if gate.Environment != nil {
			cloned[index].Environment = make(map[string]string, len(gate.Environment))
			for key, value := range gate.Environment {
				cloned[index].Environment[key] = value
			}
		}
		cloned[index].Artifacts = append([]config.GateArtifact(nil), gate.Artifacts...)
		for artifactIndex, artifact := range gate.Artifacts {
			if artifact.Required != nil {
				required := *artifact.Required
				cloned[index].Artifacts[artifactIndex].Required = &required
			}
		}
	}
	return cloned
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("plan file contains multiple JSON values")
	}
	return err
}
