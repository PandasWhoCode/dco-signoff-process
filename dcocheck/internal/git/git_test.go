package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temporary git repo with configured user and returns its path.
func initRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dcocheck-git-test-*")
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
	// suppress advice/hint noise
	run("config", "advice.detachedHead", "false")
	return dir
}

// addCommit creates a file and commits it with the given message.
func addCommit(t *testing.T, dir, filename, message string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte("content"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	run := func(args ...string) string {
		t.Helper()
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
// ValidateRepo
// ---------------------------------------------------------------------------

func TestValidateRepo_Valid(t *testing.T) {
	dir := initRepo(t)
	if err := ValidateRepo(dir); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestValidateRepo_Invalid(t *testing.T) {
	dir, err := os.MkdirTemp("", "dcocheck-notrepo-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	if err := ValidateRepo(dir); err == nil {
		t.Error("expected error for non-repo dir, got nil")
	}
}

func TestValidateRepo_NonExistent(t *testing.T) {
	if err := ValidateRepo("/tmp/this-path-should-not-exist-dcocheck-xyz"); err == nil {
		t.Error("expected error for non-existent path, got nil")
	}
}

// ---------------------------------------------------------------------------
// isValidHash
// ---------------------------------------------------------------------------

func TestIsValidHash(t *testing.T) {
	cases := []struct {
		hash  string
		valid bool
	}{
		{"abc1234", true},                                                          // min length 7
		{"abc1234def5678901234567890123456789012345", true},                        // 41 chars
		{"abc1234def567890123456789012345678901234567890123456789012345678", true},  // 64 chars
		{"abc1234def5678901234567890123456789012345678901234567890123456789", false}, // 65 chars (too long)
		{"abc123", false},   // too short (6)
		{"abc123g", false},  // non-hex char 'g'
		{"ABC1234", true},   // uppercase hex
		{"", false},         // empty
		{"abcdefg", false},  // 'g' not hex
		{"1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", true}, // 64 chars
	}
	for _, tc := range cases {
		got := isValidHash(tc.hash)
		if got != tc.valid {
			t.Errorf("isValidHash(%q) = %v, want %v", tc.hash, got, tc.valid)
		}
	}
}

// ---------------------------------------------------------------------------
// hasDCOSignoff
// ---------------------------------------------------------------------------

func TestHasDCOSignoff(t *testing.T) {
	cases := []struct {
		subject string
		body    string
		want    bool
		name    string
	}{
		{
			name:    "signoff in body",
			subject: "fix: something",
			body:    "Details here.\n\nSigned-off-by: Alice <alice@example.com>",
			want:    true,
		},
		{
			name:    "signoff in subject",
			subject: "Signed-off-by: Bob <bob@example.com>",
			body:    "",
			want:    true,
		},
		{
			name:    "no signoff",
			subject: "fix: something",
			body:    "just a body without signoff",
			want:    false,
		},
		{
			name:    "malformed signoff - missing email",
			subject: "fix: something",
			body:    "Signed-off-by: Alice",
			want:    false,
		},
		{
			name:    "malformed signoff - no angle brackets",
			subject: "fix: something",
			body:    "Signed-off-by: Alice alice@example.com",
			want:    false,
		},
		{
			name:    "empty subject and body",
			subject: "",
			body:    "",
			want:    false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasDCOSignoff(tc.subject, tc.body); got != tc.want {
				t.Errorf("hasDCOSignoff() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// parseCommits
// ---------------------------------------------------------------------------

func TestParseCommits_Empty(t *testing.T) {
	result := parseCommits("", "---SEP---")
	if len(result) != 0 {
		t.Errorf("expected 0 commits, got %d", len(result))
	}
}

func TestParseCommits_SingleWithoutDCO(t *testing.T) {
	sep := "---SEP---"
	output := sep + "\nabc1234567890\nAlice\nalice@example.com\nfix: something\njust a body\n"
	result := parseCommits(output, sep)
	if len(result) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(result))
	}
	if result[0].HasDCO {
		t.Error("expected HasDCO=false")
	}
}

func TestParseCommits_SingleWithDCO(t *testing.T) {
	sep := "---SEP---"
	output := sep + "\nabc1234567890\nAlice\nalice@example.com\nfix: something\nSigned-off-by: Alice <alice@example.com>\n"
	result := parseCommits(output, sep)
	if len(result) != 0 {
		t.Errorf("expected 0 commits (filtered out), got %d", len(result))
	}
}

func TestParseCommits_MultipleCommits(t *testing.T) {
	sep := "---SEP---"
	// Two commits: one without DCO, one with
	output := sep + "\nabc1234567890\nAlice\nalice@example.com\nfix: a\nbody without signoff\n" +
		sep + "\ndef1234567890\nBob\nbob@example.com\nfeat: b\nSigned-off-by: Bob <bob@example.com>\n"
	result := parseCommits(output, sep)
	if len(result) != 1 {
		t.Fatalf("expected 1 commit without DCO, got %d", len(result))
	}
	if result[0].Hash != "abc1234567890" {
		t.Errorf("unexpected hash: %s", result[0].Hash)
	}
}

func TestParseCommits_BlockTooFewLines(t *testing.T) {
	sep := "---SEP---"
	// Block with only 3 lines (needs at least 4)
	output := sep + "\nabc1234567890\nAlice\n"
	result := parseCommits(output, sep)
	if len(result) != 0 {
		t.Errorf("expected 0 commits from malformed block, got %d", len(result))
	}
}

func TestParseCommits_EmptyHash(t *testing.T) {
	sep := "---SEP---"
	// Block where hash line is empty
	output := sep + "\n\nAlice\nalice@example.com\nsubject\nbody\n"
	result := parseCommits(output, sep)
	if len(result) != 0 {
		t.Errorf("expected 0 commits for empty hash, got %d", len(result))
	}
}

func TestParseCommits_NoBody(t *testing.T) {
	sep := "---SEP---"
	// Only 4 lines (no body)
	output := sep + "\nabc1234567890\nAlice\nalice@example.com\nfix: something\n"
	result := parseCommits(output, sep)
	if len(result) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(result))
	}
	if result[0].Body != "" {
		t.Errorf("expected empty body, got %q", result[0].Body)
	}
}

// ---------------------------------------------------------------------------
// GetCommitsWithoutDCO
// ---------------------------------------------------------------------------

func TestGetCommitsWithoutDCO_WithCommits(t *testing.T) {
	dir := initRepo(t)

	// Commit without DCO
	addCommit(t, dir, "file1.txt", "feat: add feature\n\nNo signoff here")
	// Commit with DCO
	addCommit(t, dir, "file2.txt", "fix: bug\n\nSigned-off-by: Test User <test@example.com>")

	commits, err := GetCommitsWithoutDCO(dir)
	if err != nil {
		t.Fatalf("GetCommitsWithoutDCO: %v", err)
	}
	if len(commits) != 1 {
		t.Errorf("expected 1 commit without DCO, got %d", len(commits))
	}
	if commits[0].Subject != "feat: add feature" {
		t.Errorf("unexpected subject: %q", commits[0].Subject)
	}
}

func TestGetCommitsWithoutDCO_EmptyRepo(t *testing.T) {
	dir := initRepo(t)
	// No commits yet - git log returns empty
	commits, err := GetCommitsWithoutDCO(dir)
	if err != nil {
		t.Fatalf("unexpected error on empty repo: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected 0 commits, got %d", len(commits))
	}
}

func TestGetCommitsWithoutDCO_AllHaveDCO(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "file1.txt", "fix: something\n\nSigned-off-by: Test User <test@example.com>")
	commits, err := GetCommitsWithoutDCO(dir)
	if err != nil {
		t.Fatalf("GetCommitsWithoutDCO: %v", err)
	}
	if len(commits) != 0 {
		t.Errorf("expected 0 commits, got %d", len(commits))
	}
}

func TestGetCommitsWithoutDCO_InvalidRepo(t *testing.T) {
	_, err := GetCommitsWithoutDCO("/tmp/not-a-repo-dcocheck-xyz")
	if err == nil {
		t.Error("expected error for invalid repo, got nil")
	}
}

// ---------------------------------------------------------------------------
// GetTotalCommitCount
// ---------------------------------------------------------------------------

func TestGetTotalCommitCount_WithCommits(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "first")
	addCommit(t, dir, "b.txt", "second")

	count, err := GetTotalCommitCount(dir)
	if err != nil {
		t.Fatalf("GetTotalCommitCount: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestGetTotalCommitCount_EmptyRepo(t *testing.T) {
	dir := initRepo(t)
	count, err := GetTotalCommitCount(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// GetCommitDetails
// ---------------------------------------------------------------------------

func TestGetCommitDetails_ValidHash(t *testing.T) {
	dir := initRepo(t)
	hash := addCommit(t, dir, "file.txt", "test commit")

	details, err := GetCommitDetails(dir, hash)
	if err != nil {
		t.Fatalf("GetCommitDetails: %v", err)
	}
	if !strings.Contains(details, "test commit") {
		t.Errorf("expected commit message in details, got: %q", details)
	}
}

func TestGetCommitDetails_InvalidHashFormat(t *testing.T) {
	dir := initRepo(t)
	_, err := GetCommitDetails(dir, "not-a-valid-hash!!!")
	if err == nil {
		t.Error("expected error for invalid hash format, got nil")
	}
}

func TestGetCommitDetails_TooShortHash(t *testing.T) {
	dir := initRepo(t)
	_, err := GetCommitDetails(dir, "abc12")
	if err == nil {
		t.Error("expected error for too-short hash, got nil")
	}
}

func TestGetCommitDetails_ValidHashNotInRepo(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "file.txt", "init")
	// Valid 40-char hex hash that does not exist in this repository.
	_, err := GetCommitDetails(dir, "abcdef1234567890abcdef1234567890abcdef12")
	if err == nil {
		t.Error("expected error for valid-format hash not present in repo, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get commit details") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GetTotalCommitCount – Sscanf error path
// ---------------------------------------------------------------------------

func TestGetTotalCommitCount_NonNumericOutput(t *testing.T) {
	orig := gitRevListCountOutputFn
	gitRevListCountOutputFn = func(_ string) ([]byte, error) {
		return []byte("not-a-number\n"), nil
	}
	t.Cleanup(func() { gitRevListCountOutputFn = orig })

	count, err := GetTotalCommitCount(".")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 for non-numeric git output, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// GetGitUserName
// ---------------------------------------------------------------------------

func TestGetGitUserName_Valid(t *testing.T) {
	dir := initRepo(t) // sets user.name = "Test User"
	name, err := GetGitUserName(dir)
	if err != nil {
		t.Fatalf("GetGitUserName: %v", err)
	}
	if name != "Test User" {
		t.Errorf("expected 'Test User', got %q", name)
	}
}

func TestGetGitUserName_InvalidRepo(t *testing.T) {
	_, err := GetGitUserName("/tmp/not-a-repo-dcocheck-xyz-getusername")
	if err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get git user.name") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ---------------------------------------------------------------------------
// CreateRetroactiveSignoffCommit
// ---------------------------------------------------------------------------

func TestCreateRetroactiveSignoffCommit_InvalidRepo(t *testing.T) {
	err := CreateRetroactiveSignoffCommit("/tmp/not-a-repo-dcocheck-xyz-signoff", "test message")
	if err == nil {
		t.Error("expected error for invalid repo path, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create retroactive signoff commit") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCreateRetroactiveSignoffCommit_Success(t *testing.T) {
	// Replace the commit command with `true` so no real GPG key is required.
	orig := commitCommandFn
	commitCommandFn = func(_, _ string) *exec.Cmd {
		return exec.Command("true")
	}
	t.Cleanup(func() { commitCommandFn = orig })

	if err := CreateRetroactiveSignoffCommit(".", "test message"); err != nil {
		t.Errorf("unexpected error from mocked commit: %v", err)
	}
}
