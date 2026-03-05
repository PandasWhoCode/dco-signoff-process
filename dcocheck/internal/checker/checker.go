// Package checker provides the DCO sign-off checking logic. It uses the git
// package to retrieve commits from a repository, filters those that are missing
// a "Signed-off-by:" trailer, and produces formatted output sections that mirror
// the reveal_dco_issues bash script described in PROCESS.md. It also builds the
// retroactive sign-off commit message used when the user elects to perform an
// in-place retroactive DCO sign-off.
package checker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/PandasWhoCode/dco-signoff-process/dcocheck/internal/git"
)

// Result holds the results of a DCO check
type Result struct {
	RepoPath          string
	TotalCommits      int
	CommitsWithoutDCO []git.Commit
	UniqueAuthors     []string
}

// Options configures the checker
type Options struct {
	DryRun bool
}

// Check performs a DCO check on the given repository
func Check(repoPath string, opts Options) (*Result, error) {
	if err := git.ValidateRepo(repoPath); err != nil {
		return nil, err
	}

	commits, err := git.GetCommitsWithoutDCO(repoPath)
	if err != nil {
		return nil, err
	}

	// Get total commit count separately
	totalCommits, err := git.GetTotalCommitCount(repoPath)
	if err != nil {
		return nil, err
	}

	// Collect unique authors
	authorSet := make(map[string]struct{})
	for _, c := range commits {
		key := fmt.Sprintf("%s <%s>", c.Author, c.Email)
		authorSet[key] = struct{}{}
	}
	authors := make([]string, 0, len(authorSet))
	for a := range authorSet {
		authors = append(authors, a)
	}
	sort.Strings(authors)

	return &Result{
		RepoPath:          repoPath,
		TotalCommits:      totalCommits,
		CommitsWithoutDCO: commits,
		UniqueAuthors:     authors,
	}, nil
}

// FormatSummaryOfAuthors returns a formatted string of unique authors
func (r *Result) FormatSummaryOfAuthors() []string {
	var lines []string
	lines = append(lines, strings.Repeat("<", 80))
	lines = append(lines, "Summary of Unique Authors")
	lines = append(lines, strings.Repeat(">", 80))
	lines = append(lines, "")
	lines = append(lines, strings.Repeat("-", 80))
	lines = append(lines, r.UniqueAuthors...)
	lines = append(lines, strings.Repeat("-", 80))
	return lines
}

// FormatSummaryOfCommits returns a formatted string of commits without DCO
func (r *Result) FormatSummaryOfCommits() []string {
	var lines []string
	lines = append(lines, strings.Repeat("<", 80))
	lines = append(lines, "Summary of Commits")
	lines = append(lines, strings.Repeat(">", 80))
	lines = append(lines, "")
	lines = append(lines, strings.Repeat("-", 80))
	for _, c := range r.CommitsWithoutDCO {
		lines = append(lines, fmt.Sprintf("%s: %s <%s>", c.Hash, c.Author, c.Email))
	}
	lines = append(lines, strings.Repeat("-", 80))
	return lines
}

// FormatRetroactiveMessage returns the DCO retroactive message format
func (r *Result) FormatRetroactiveMessage() []string {
	var lines []string
	lines = append(lines, strings.Repeat("<", 80))
	lines = append(lines, "DCO Retroactive Message Format")
	lines = append(lines, strings.Repeat(">", 80))
	lines = append(lines, "")
	lines = append(lines, strings.Repeat("-", 80))
	for _, c := range r.CommitsWithoutDCO {
		shortHash := c.Hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}
		lines = append(lines, fmt.Sprintf("commit %s %s", shortHash, c.Subject))
	}
	lines = append(lines, strings.Repeat("-", 80))
	return lines
}

// FormatFullCommitLog returns lines for the full commit log section
func (r *Result) FormatFullCommitLog(repoPath string) []string {
	var lines []string
	lines = append(lines, strings.Repeat("<", 80))
	lines = append(lines, "Full Commit Log History")
	lines = append(lines, strings.Repeat(">", 80))
	lines = append(lines, "")
	for _, c := range r.CommitsWithoutDCO {
		lines = append(lines, strings.Repeat("-", 80))
		detail, err := git.GetCommitDetails(repoPath, c.Hash)
		if err != nil {
			lines = append(lines, fmt.Sprintf("commit %s (error fetching details: %v)", c.Hash, err))
		} else {
			lines = append(lines, detail)
		}
	}
	lines = append(lines, strings.Repeat("-", 80))
	return lines
}

// BuildRetroactiveCommitMessage builds the commit message body for a retroactive
// DCO sign-off empty commit. gitUserName should be the value of `git config user.name`.
func (r *Result) BuildRetroactiveCommitMessage(gitUserName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("I, %s, retroactively sign off on these commits:\n\n", gitUserName))
	for _, c := range r.CommitsWithoutDCO {
		shortHash := c.Hash
		if len(shortHash) > 8 {
			shortHash = shortHash[:8]
		}
		sb.WriteString(fmt.Sprintf("commit %s %s\n", shortHash, c.Subject))
	}
	return sb.String()
}

// AllOutput returns all output sections combined
func (r *Result) AllOutput(repoPath string) []string {
	var all []string
	all = append(all, r.FormatSummaryOfAuthors()...)
	all = append(all, "")
	all = append(all, r.FormatSummaryOfCommits()...)
	all = append(all, "")
	all = append(all, r.FormatRetroactiveMessage()...)
	all = append(all, "")
	all = append(all, r.FormatFullCommitLog(repoPath)...)
	return all
}
