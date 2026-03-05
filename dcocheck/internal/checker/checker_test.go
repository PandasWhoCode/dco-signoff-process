package checker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PandasWhoCode/dco-signoff-process/dcocheck/internal/git"
)

// initRepo creates a temporary git repo for testing.
func initRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dcocheck-checker-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")
	return dir
}

func addCommit(t *testing.T, dir, filename, message string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte("content"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	run := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("add", filename)
	run("commit", "-m", message)
	return run("rev-parse", "HEAD")
}

// ---------------------------------------------------------------------------
// Check
// ---------------------------------------------------------------------------

func TestCheck_InvalidRepo(t *testing.T) {
	_, err := Check("/tmp/not-a-repo-dcocheck-xyz", Options{})
	if err == nil {
		t.Error("expected error for invalid repo, got nil")
	}
}

func TestCheck_GetCommitsError(t *testing.T) {
	dir := initRepo(t)
	orig := getCommitsWithoutDCOFn
	getCommitsWithoutDCOFn = func(_ string) ([]git.Commit, error) {
		return nil, fmt.Errorf("injected git log error")
	}
	t.Cleanup(func() { getCommitsWithoutDCOFn = orig })

	_, err := Check(dir, Options{})
	if err == nil {
		t.Error("expected error from injected GetCommitsWithoutDCO, got nil")
	}
	if !strings.Contains(err.Error(), "injected git log error") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheck_GetTotalCountError(t *testing.T) {
	dir := initRepo(t)
	orig := getTotalCommitCountFn
	getTotalCommitCountFn = func(_ string) (int, error) {
		return 0, fmt.Errorf("injected count error")
	}
	t.Cleanup(func() { getTotalCommitCountFn = orig })

	_, err := Check(dir, Options{})
	if err == nil {
		t.Error("expected error from injected GetTotalCommitCount, got nil")
	}
	if !strings.Contains(err.Error(), "injected count error") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheck_EmptyRepo(t *testing.T) {
	dir := initRepo(t)
	result, err := Check(dir, Options{})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.TotalCommits != 0 {
		t.Errorf("expected 0 total commits, got %d", result.TotalCommits)
	}
	if len(result.CommitsWithoutDCO) != 0 {
		t.Errorf("expected 0 commits without DCO, got %d", len(result.CommitsWithoutDCO))
	}
	if len(result.UniqueAuthors) != 0 {
		t.Errorf("expected 0 unique authors, got %d", len(result.UniqueAuthors))
	}
}

func TestCheck_CommitsWithoutDCO(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")
	addCommit(t, dir, "b.txt", "fix: also no signoff")
	addCommit(t, dir, "c.txt", "chore: has signoff\n\nSigned-off-by: Test User <test@example.com>")

	result, err := Check(dir, Options{})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if result.TotalCommits != 3 {
		t.Errorf("expected 3 total commits, got %d", result.TotalCommits)
	}
	if len(result.CommitsWithoutDCO) != 2 {
		t.Errorf("expected 2 commits without DCO, got %d", len(result.CommitsWithoutDCO))
	}
	if len(result.UniqueAuthors) != 1 {
		t.Errorf("expected 1 unique author, got %d", len(result.UniqueAuthors))
	}
}

func TestCheck_AllWithDCO(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "fix: signed\n\nSigned-off-by: Test User <test@example.com>")

	result, err := Check(dir, Options{DryRun: true})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.CommitsWithoutDCO) != 0 {
		t.Errorf("expected 0 commits without DCO, got %d", len(result.CommitsWithoutDCO))
	}
}

func TestCheck_MultipleAuthors(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: by alice")

	// Change git user for second commit
	cmd := exec.Command("git", "config", "user.name", "Bob")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.email", "bob@example.com")
	cmd.Dir = dir
	cmd.Run()

	addCommit(t, dir, "b.txt", "feat: by bob")

	result, err := Check(dir, Options{})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(result.UniqueAuthors) != 2 {
		t.Errorf("expected 2 unique authors, got %d: %v", len(result.UniqueAuthors), result.UniqueAuthors)
	}
}

// ---------------------------------------------------------------------------
// FormatSummaryOfAuthors
// ---------------------------------------------------------------------------

func TestFormatSummaryOfAuthors_Empty(t *testing.T) {
	r := &Result{}
	lines := r.FormatSummaryOfAuthors()
	if len(lines) == 0 {
		t.Error("expected non-empty output")
	}
	found := false
	for _, l := range lines {
		if l == "Summary of Unique Authors" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Summary of Unique Authors' header")
	}
}

func TestFormatSummaryOfAuthors_WithAuthors(t *testing.T) {
	r := &Result{UniqueAuthors: []string{"Alice <a@b.com>", "Bob <b@c.com>"}}
	lines := r.FormatSummaryOfAuthors()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "Alice <a@b.com>") {
		t.Error("expected Alice in output")
	}
	if !strings.Contains(content, "Bob <b@c.com>") {
		t.Error("expected Bob in output")
	}
}

// ---------------------------------------------------------------------------
// FormatSummaryOfCommits
// ---------------------------------------------------------------------------

func TestFormatSummaryOfCommits_Empty(t *testing.T) {
	r := &Result{}
	lines := r.FormatSummaryOfCommits()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "Summary of Commits") {
		t.Error("expected 'Summary of Commits' header")
	}
}

func TestFormatSummaryOfCommits_WithCommits(t *testing.T) {
	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: "abc1234567890", Author: "Alice", Email: "a@b.com", Subject: "fix: x"},
		},
	}
	lines := r.FormatSummaryOfCommits()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "abc1234567890: Alice <a@b.com>") {
		t.Errorf("unexpected output: %s", content)
	}
}

// ---------------------------------------------------------------------------
// FormatRetroactiveMessage
// ---------------------------------------------------------------------------

func TestFormatRetroactiveMessage_Empty(t *testing.T) {
	r := &Result{}
	lines := r.FormatRetroactiveMessage()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "DCO Retroactive Message Format") {
		t.Error("expected 'DCO Retroactive Message Format' header")
	}
}

func TestFormatRetroactiveMessage_WithCommits(t *testing.T) {
	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: "abcdef1234567890", Subject: "feat: new thing"},
		},
	}
	lines := r.FormatRetroactiveMessage()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "commit abcdef12 feat: new thing") {
		t.Errorf("unexpected output: %s", content)
	}
}

func TestFormatRetroactiveMessage_ShortHash(t *testing.T) {
	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: "abc1234", Subject: "feat: short"},
		},
	}
	lines := r.FormatRetroactiveMessage()
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "commit abc1234 feat: short") {
		t.Errorf("unexpected output: %s", content)
	}
}

// ---------------------------------------------------------------------------
// FormatFullCommitLog
// ---------------------------------------------------------------------------

func TestFormatFullCommitLog_Empty(t *testing.T) {
	r := &Result{}
	lines := r.FormatFullCommitLog(".")
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "Full Commit Log History") {
		t.Error("expected 'Full Commit Log History' header")
	}
}

func TestFormatFullCommitLog_WithRealCommit(t *testing.T) {
	dir := initRepo(t)
	hash := addCommit(t, dir, "f.txt", "test commit message")

	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: hash, Author: "Test User", Email: "test@example.com", Subject: "test commit message"},
		},
	}
	lines := r.FormatFullCommitLog(dir)
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "test commit message") {
		t.Errorf("expected commit message in full log, got: %s", content)
	}
}

func TestFormatFullCommitLog_ErrorFetchingDetails(t *testing.T) {
	// Use an invalid hash to trigger error path
	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: "abcdef1234567890abcdef1234567890abcdef12", Author: "X", Email: "x@y.com"},
		},
	}
	dir := initRepo(t)
	addCommit(t, dir, "f.txt", "init")
	lines := r.FormatFullCommitLog(dir)
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "error fetching details") {
		t.Errorf("expected error message in output, got: %s", content)
	}
}

// ---------------------------------------------------------------------------
// AllOutput
// ---------------------------------------------------------------------------

func TestAllOutput_Empty(t *testing.T) {
	r := &Result{}
	lines := r.AllOutput(".")
	content := strings.Join(lines, "\n")
	if !strings.Contains(content, "Summary of Unique Authors") {
		t.Error("expected authors section")
	}
	if !strings.Contains(content, "Summary of Commits") {
		t.Error("expected commits section")
	}
	if !strings.Contains(content, "DCO Retroactive Message Format") {
		t.Error("expected retroactive section")
	}
	if !strings.Contains(content, "Full Commit Log History") {
		t.Error("expected full log section")
	}
}

func TestAllOutput_WithData(t *testing.T) {
	dir := initRepo(t)
	hash := addCommit(t, dir, "f.txt", "feat: something")

	r := &Result{
		RepoPath:     dir,
		TotalCommits: 1,
		CommitsWithoutDCO: []git.Commit{
			{Hash: hash, Author: "Test User", Email: "test@example.com", Subject: "feat: something"},
		},
		UniqueAuthors: []string{"Test User <test@example.com>"},
	}
	lines := r.AllOutput(dir)
	if len(lines) == 0 {
		t.Error("expected non-empty output")
	}
}

// ---------------------------------------------------------------------------
// BuildRetroactiveCommitMessage
// ---------------------------------------------------------------------------

func TestBuildRetroactiveCommitMessage_NoCommits(t *testing.T) {
	r := &Result{}
	msg := r.BuildRetroactiveCommitMessage("Test User")
	if !strings.Contains(msg, "I, Test User, retroactively sign off on these commits:") {
		t.Errorf("expected header line, got: %s", msg)
	}
}

func TestBuildRetroactiveCommitMessage_WithCommits(t *testing.T) {
	r := &Result{
		CommitsWithoutDCO: []git.Commit{
			{Hash: "abcdef1234567890", Subject: "feat: new thing"},
			{Hash: "abc1234", Subject: "fix: bug"},
		},
	}
	msg := r.BuildRetroactiveCommitMessage("Alice Smith")
	if !strings.Contains(msg, "I, Alice Smith, retroactively sign off on these commits:") {
		t.Errorf("expected header, got: %s", msg)
	}
	// Long hash: first 8 chars
	if !strings.Contains(msg, "commit abcdef12 feat: new thing") {
		t.Errorf("expected truncated hash line, got: %s", msg)
	}
	// Short hash (≤8 chars): kept as-is
	if !strings.Contains(msg, "commit abc1234 fix: bug") {
		t.Errorf("expected short hash line, got: %s", msg)
	}
}
