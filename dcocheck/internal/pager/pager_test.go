package pager

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

// pipeForTest creates an os.Pipe and returns the read and write ends.
func pipeForTest(t *testing.T) (*os.File, *os.File, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	t.Cleanup(func() { r.Close() })
	return r, w, nil
}

func TestIsTerminal(t *testing.T) {
	// In a test environment, stdout is not a terminal
	got := IsTerminal()
	if got {
		t.Error("expected IsTerminal() to return false in test environment")
	}
}

// ---------------------------------------------------------------------------
// getPageSize
// ---------------------------------------------------------------------------

func TestGetPageSize(t *testing.T) {
	// In test environment, stdout is not a terminal, so it returns defaultPageSize
	size := getPageSize()
	if size <= 0 {
		t.Errorf("expected positive page size, got %d", size)
	}
}

// ---------------------------------------------------------------------------
// Display - non-interactive
// ---------------------------------------------------------------------------

func TestDisplay_NonInteractive_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := Display([]string{}, &buf, false)
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

func TestDisplay_NonInteractive_Lines(t *testing.T) {
	lines := []string{"line1", "line2", "line3"}
	var buf bytes.Buffer
	err := Display(lines, &buf, false)
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	output := buf.String()
	for _, l := range lines {
		if !strings.Contains(output, l) {
			t.Errorf("expected %q in output, got: %s", l, output)
		}
	}
}

func TestDisplay_NonInteractive_Interactive_EmptyLines(t *testing.T) {
	// interactive=true but len(lines)==0 falls through to non-interactive path
	var buf bytes.Buffer
	err := Display([]string{}, &buf, true)
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output, got %q", buf.String())
	}
}

// ---------------------------------------------------------------------------
// Display - write error
// ---------------------------------------------------------------------------

type errWriter struct{}

func (e *errWriter) Write(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}

func TestDisplay_WriteError(t *testing.T) {
	lines := []string{"line1"}
	err := Display(lines, &errWriter{}, false)
	if err == nil {
		t.Error("expected error from failing writer, got nil")
	}
}

// ---------------------------------------------------------------------------
// Display - interactive mode with mocked stdin
// ---------------------------------------------------------------------------

func TestDisplay_Interactive_FitsOnePage(t *testing.T) {
	// Lines fit in one page (defaultPageSize=20, we use 3 lines)
	// No prompt should appear since all lines fit
	lines := []string{"a", "b", "c"}
	var buf bytes.Buffer
	// Even with interactive=true, if lines <= pageSize, no prompt is shown
	err := Display(lines, &buf, true)
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") || !strings.Contains(output, "c") {
		t.Errorf("unexpected output: %s", output)
	}
}

func TestDisplay_Interactive_MultiPage_Continue(t *testing.T) {
	// Replace stdin with a pipe that simulates user pressing Enter (continue)
	oldStdin := stdinOverride
	r, w, err := pipeForTest(t)
	if err != nil {
		t.Fatal(err)
	}
	stdinOverride = r
	defer func() {
		stdinOverride = oldStdin
		r.Close()
	}()

	// Write "enter" for each page break
	go func() {
		defer w.Close()
		// We'll have enough lines to require 2 pages at defaultPageSize=20
		// So one page prompt, user presses enter
		w.WriteString("\n")
	}()

	lines := make([]string, defaultPageSize+5)
	for i := range lines {
		lines[i] = "line"
	}
	var buf bytes.Buffer
	if err := Display(lines, &buf, true); err != nil {
		t.Fatalf("Display interactive: %v", err)
	}
}

func TestDisplay_Interactive_MultiPage_Quit(t *testing.T) {
	oldStdin := stdinOverride
	r, w, err := pipeForTest(t)
	if err != nil {
		t.Fatal(err)
	}
	stdinOverride = r
	defer func() {
		stdinOverride = oldStdin
		r.Close()
	}()

	go func() {
		defer w.Close()
		w.WriteString("q\n")
	}()

	lines := make([]string, defaultPageSize+5)
	for i := range lines {
		lines[i] = "line"
	}
	var buf bytes.Buffer
	err = Display(lines, &buf, true)
	if err != nil {
		t.Fatalf("Display interactive quit: %v", err)
	}
	// Should have only printed first page
	lineCount := strings.Count(buf.String(), "\n")
	if lineCount > defaultPageSize+1 {
		t.Errorf("expected at most %d lines printed, got %d", defaultPageSize, lineCount)
	}
}
