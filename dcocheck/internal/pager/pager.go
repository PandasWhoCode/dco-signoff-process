package pager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

const defaultPageSize = 20

// stdinOverride allows tests to replace os.Stdin with a custom reader.
var stdinOverride io.Reader

// stdinReader returns the active stdin source.
func stdinReader() io.Reader {
	if stdinOverride != nil {
		return stdinOverride
	}
	return os.Stdin
}

// getPageSize returns the terminal height minus 2, or defaultPageSize
func getPageSize() int {
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || height <= 3 {
		return defaultPageSize
	}
	return height - 2
}

// Display shows lines in a pageable format.
// If interactive is false, all lines are printed without paging.
// If interactive is true, lines are displayed page by page.
func Display(lines []string, out io.Writer, interactive bool) error {
	if !interactive || len(lines) == 0 {
		for _, line := range lines {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return err
			}
		}
		return nil
	}

	pageSize := getPageSize()
	total := len(lines)
	reader := bufio.NewReader(stdinReader())

	for i := 0; i < total; i += pageSize {
		end := i + pageSize
		if end > total {
			end = total
		}

		for _, line := range lines[i:end] {
			if _, err := fmt.Fprintln(out, line); err != nil {
				return err
			}
		}

		if end < total {
			totalPages := (total + pageSize - 1) / pageSize
			currentPage := (i / pageSize) + 1
			fmt.Fprintf(os.Stderr, "-- Page %d/%d (showing lines %d-%d of %d) -- [Enter] continue, [q] quit: ",
				currentPage, totalPages, i+1, end, total)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if strings.EqualFold(input, "q") {
				return nil
			}
		}
	}
	return nil
}

// IsTerminal returns true if stdout is a terminal
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
