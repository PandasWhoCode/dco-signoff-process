package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initRepo creates a temporary git repo for testing.
func initRepo(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "dcocheck-main-test-*")
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

func addCommit(t *testing.T, dir, filename, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filename), []byte("content"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("add", filename)
	run("commit", "-m", message)
}

func captureFiles(t *testing.T) (*os.File, *os.File, *os.File, *os.File) {
	t.Helper()
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	return outR, outW, errR, errW
}

func readPipe(f *os.File) string {
	buf, _ := io.ReadAll(f)
	return string(buf)
}

// ---------------------------------------------------------------------------
// run() tests
// ---------------------------------------------------------------------------

func TestRun_Help(t *testing.T) {
	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"--help"}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "dcocheck - DCO Sign-off Checker") {
		t.Errorf("expected help text, got: %s", output)
	}
}

func TestRun_ShortHelp(t *testing.T) {
	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"-h"}, outW, errW, strings.NewReader(""))
	outW.Close()
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_Version(t *testing.T) {
	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"--version"}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "dcocheck version 1.0.0") {
		t.Errorf("expected version string, got: %s", output)
	}
}

func TestRun_ShortVersion(t *testing.T) {
	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"-v"}, outW, errW, strings.NewReader(""))
	outW.Close()
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_InvalidRepo(t *testing.T) {
	outR, outW, errR, errW := captureFiles(t)
	defer outR.Close()
	defer errR.Close()

	code := run([]string{"/tmp/no-such-repo-dcocheck-xyz"}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	errOut := readPipe(errR)
	if !strings.Contains(errOut, "Error") {
		t.Errorf("expected error message, got: %s", errOut)
	}
}

func TestRun_AllCommitsHaveDCO(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "fix: signed\n\nSigned-off-by: Test User <test@example.com>")

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "No issues found") {
		t.Errorf("expected 'No issues found', got: %s", output)
	}
}

func TestRun_CommitsWithoutDCO_NonInteractive(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	// Default isInteractiveFn returns false in test environment
	code := run([]string{dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "without DCO sign-off") {
		t.Errorf("expected DCO issue message, got: %s", output)
	}
}

// TestRun_CommitsWithoutDCO is an alias kept for backward compat of naming.
func TestRun_CommitsWithoutDCO(t *testing.T) {
	TestRun_CommitsWithoutDCO_NonInteractive(t)
}

func TestRun_DryRun(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"--dry-run", dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0 for dry-run, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected dry-run message, got: %s", output)
	}
}

func TestRun_ShortDryRun(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"-d", dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0 for -d, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_OutputFile(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	tmpFile, err := os.CreateTemp("", "dcocheck-output-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"--output", tmpFile.Name(), dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "Results written to") {
		t.Errorf("expected 'Results written to', got: %s", output)
	}

	// Check file was written
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty output file")
	}
}

func TestRun_ShortOutputFile(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	tmpFile, err := os.CreateTemp("", "dcocheck-output-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"-o", tmpFile.Name(), dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_OutputFileError(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	outR, outW, errR, errW := captureFiles(t)
	defer outR.Close()
	defer errR.Close()

	// Use an invalid path
	code := run([]string{"--output", "/no/such/dir/file.txt", dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	errOut := readPipe(errR)
	if !strings.Contains(errOut, "Error writing to file") {
		t.Errorf("expected file write error, got: %s", errOut)
	}
}

func TestRun_InvalidFlag(t *testing.T) {
	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{"--unknown-flag"}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRun_DefaultPath(t *testing.T) {
	// Run with no arguments - will use "." which may or may not be a git repo
	// We just check it doesn't panic
	outR, outW, errR, errW := captureFiles(t)
	defer outR.Close()
	defer errR.Close()

	code := run([]string{}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	_ = code // 0, 1, or 2 depending on environment
	_ = readPipe(outR)
}

// ---------------------------------------------------------------------------
// Interactive retroactive sign-off via promptRetroactiveSignoff
// ---------------------------------------------------------------------------

// withInteractive sets isInteractiveFn to always return true for the duration
// of the test, then restores the original.
func withInteractive(t *testing.T) {
	t.Helper()
	orig := isInteractiveFn
	isInteractiveFn = func() bool { return true }
	t.Cleanup(func() { isInteractiveFn = orig })
}

// withMockCreateCommit replaces createRetroactiveSignoffCommitFn with a
// function that returns the supplied error (nil = success).
func withMockCreateCommit(t *testing.T, retErr error) {
	t.Helper()
	orig := createRetroactiveSignoffCommitFn
	createRetroactiveSignoffCommitFn = func(_, _ string) error { return retErr }
	t.Cleanup(func() { createRetroactiveSignoffCommitFn = orig })
}

// withMockGetUserName replaces getGitUserNameFn with a function that returns
// the supplied name/error pair.
func withMockGetUserName(t *testing.T, name string, retErr error) {
	t.Helper()
	orig := getGitUserNameFn
	getGitUserNameFn = func(_ string) (string, error) { return name, retErr }
	t.Cleanup(func() { getGitUserNameFn = orig })
}

func TestRun_SignoffYes(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)
	withMockGetUserName(t, "Test User", nil)
	withMockCreateCommit(t, nil)

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader("y\n"))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0 for successful sign-off, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "Retroactive DCO sign-off commit created successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
}

func TestRun_SignoffUpperY(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)
	withMockGetUserName(t, "Test User", nil)
	withMockCreateCommit(t, nil)

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader("Y\n"))
	outW.Close()
	errW.Close()

	if code != 0 {
		t.Errorf("expected exit code 0 for uppercase Y, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_SignoffNo(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader("n\n"))
	outW.Close()
	errW.Close()

	if code != 1 {
		t.Errorf("expected exit code 1 for declined sign-off, got %d", code)
	}
	output := readPipe(outR)
	if !strings.Contains(output, "Skipping retroactive sign-off") {
		t.Errorf("expected skip message, got: %s", output)
	}
}

func TestRun_SignoffEmptyInput(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	// Empty input (just Enter) should be treated as "no"
	code := run([]string{dir}, outW, errW, strings.NewReader("\n"))
	outW.Close()
	errW.Close()

	if code != 1 {
		t.Errorf("expected exit code 1 for empty input, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_SignoffEOF(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)

	outR, outW, _, errW := captureFiles(t)
	defer outR.Close()
	defer errW.Close()

	// Closed reader simulates EOF on stdin
	code := run([]string{dir}, outW, errW, strings.NewReader(""))
	outW.Close()
	errW.Close()

	if code != 1 {
		t.Errorf("expected exit code 1 on stdin EOF, got %d", code)
	}
	_ = readPipe(outR)
}

func TestRun_SignoffGetUserNameError(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)
	withMockGetUserName(t, "", errors.New("git user.name not set"))

	outR, outW, errR, errW := captureFiles(t)
	defer outR.Close()
	defer errR.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader("y\n"))
	outW.Close()
	errW.Close()

	if code != 2 {
		t.Errorf("expected exit code 2 when user name lookup fails, got %d", code)
	}
	errOut := readPipe(errR)
	if !strings.Contains(errOut, "failed to get git user name") {
		t.Errorf("expected user name error, got: %s", errOut)
	}
}

func TestRun_SignoffCommitError(t *testing.T) {
	dir := initRepo(t)
	addCommit(t, dir, "a.txt", "feat: no signoff")

	withInteractive(t)
	withMockGetUserName(t, "Test User", nil)
	withMockCreateCommit(t, errors.New("gpg: no secret key"))

	outR, outW, errR, errW := captureFiles(t)
	defer outR.Close()
	defer errR.Close()

	code := run([]string{dir}, outW, errW, strings.NewReader("y\n"))
	outW.Close()
	errW.Close()

	if code != 2 {
		t.Errorf("expected exit code 2 when commit creation fails, got %d", code)
	}
	errOut := readPipe(errR)
	if !strings.Contains(errOut, "Error") {
		t.Errorf("expected error message, got: %s", errOut)
	}
}

// ---------------------------------------------------------------------------
// writeToFile
// ---------------------------------------------------------------------------

func TestWriteToFile_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "dcocheck-write-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	lines := []string{"line1", "line2", "line3"}
	if err := writeToFile(tmpFile.Name(), lines); err != nil {
		t.Fatalf("writeToFile: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, l := range lines {
		if !strings.Contains(content, l) {
			t.Errorf("expected %q in file, got: %s", l, content)
		}
	}
}

func TestWriteToFile_InvalidPath(t *testing.T) {
	err := writeToFile("/no/such/directory/file.txt", []string{"line"})
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}
