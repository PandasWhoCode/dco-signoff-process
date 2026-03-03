package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/PandasWhoCode/dco-signoff-process/dcocheck/internal/checker"
	"github.com/PandasWhoCode/dco-signoff-process/dcocheck/internal/pager"
)

const version = "1.0.0"

const helpText = `dcocheck - DCO Sign-off Checker

USAGE:
    dcocheck [OPTIONS] [PATH]

ARGUMENTS:
    PATH    Path to the git repository to check (default: current directory)

OPTIONS:
    -d, --dry-run    Show results without prompting for commit creation
    -o, --output     Write results to the specified file
    -h, --help       Show this help message
    -v, --version    Show version information

DESCRIPTION:
    dcocheck scans a git repository for commits missing Developer Certificate
    of Origin (DCO) sign-off (Signed-off-by: trailer). It displays a pageable
    list of commits requiring retroactive sign-off.

    Results include:
      - Summary of unique authors with missing sign-off
      - Summary of commits missing sign-off  
      - DCO retroactive message format (for use in git commit message)
      - Full commit log history for affected commits

EXAMPLES:
    dcocheck .
    dcocheck /path/to/repo
    dcocheck --dry-run /path/to/repo
    dcocheck --output results.txt /path/to/repo

EXIT CODES:
    0    No DCO issues found (or --dry-run mode)
    1    DCO issues found
    2    Error occurred
`

func run(args []string, stdout, stderr *os.File) int {
	fs := flag.NewFlagSet("dcocheck", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		dryRun     bool
		outputFile string
		showHelp   bool
		showVer    bool
	)

	// Support both short and long flags
	fs.BoolVar(&dryRun, "d", false, "dry run")
	fs.BoolVar(&dryRun, "dry-run", false, "dry run")
	fs.StringVar(&outputFile, "o", "", "output file")
	fs.StringVar(&outputFile, "output", "", "output file")
	fs.BoolVar(&showHelp, "h", false, "show help")
	fs.BoolVar(&showHelp, "help", false, "show help")
	fs.BoolVar(&showVer, "v", false, "show version")
	fs.BoolVar(&showVer, "version", false, "show version")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if showHelp {
		fmt.Fprint(stdout, helpText)
		return 0
	}

	if showVer {
		fmt.Fprintf(stdout, "dcocheck version %s\n", version)
		return 0
	}

	// Determine repo path
	repoPath := "."
	if fs.NArg() > 0 {
		repoPath = fs.Arg(0)
	}
	repoPath = filepath.Clean(repoPath)

	// Run the check
	opts := checker.Options{DryRun: dryRun}
	result, err := checker.Check(repoPath, opts)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 2
	}

	if len(result.CommitsWithoutDCO) == 0 {
		fmt.Fprintf(stdout, "✓ All %d commits have DCO sign-off. No issues found.\n", result.TotalCommits)
		return 0
	}

	// Print summary header
	fmt.Fprintf(stdout, "Found %d commit(s) without DCO sign-off (out of %d total commits)\n\n",
		len(result.CommitsWithoutDCO), result.TotalCommits)

	// Get all output lines
	allLines := result.AllOutput(repoPath)

	// Write to file if requested
	if outputFile != "" {
		if err := writeToFile(outputFile, allLines); err != nil {
			fmt.Fprintf(stderr, "Error writing to file: %v\n", err)
			return 2
		}
		fmt.Fprintf(stdout, "Results written to: %s\n\n", outputFile)
	}

	// Display in terminal (pageable if interactive)
	interactive := pager.IsTerminal() && !dryRun
	if err := pager.Display(allLines, stdout, interactive); err != nil {
		fmt.Fprintf(stderr, "Error displaying output: %v\n", err)
		return 2
	}

	if dryRun {
		fmt.Fprintf(stdout, "\n[dry-run] Exiting without creating retroactive commit.\n")
		return 0
	}

	return 1
}

func writeToFile(path string, lines []string) error {
	// #nosec G304 - path is user-provided output file, intended
	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open output file %s: %w", path, err)
	}
	defer f.Close()

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
