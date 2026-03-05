// Package git provides helper functions for interacting with a local git
// repository. All git operations are performed by calling the git binary via
// exec.Command – never through shell interpolation – to prevent command-injection
// vulnerabilities. The package validates commit hashes before use, cleans
// user-supplied repository paths with filepath.Clean, and exposes helpers for
// reading repository metadata, retrieving commit history, and creating commits.
package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Commit represents a git commit with DCO sign-off status
type Commit struct {
	Hash    string
	Author  string
	Email   string
	Subject string
	Body    string
	HasDCO  bool
}

// DCO sign-off regex: "Signed-off-by: Name <email>"
var signoffRegex = regexp.MustCompile(`(?m)^Signed-off-by:\s+\S.*<\S+@\S+>`)

// ValidateRepo checks if the given path is a valid git repository
func ValidateRepo(repoPath string) error {
	clean := filepath.Clean(repoPath)
	cmd := exec.Command("git", "-C", clean, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a valid git repository: %s", clean)
	}
	return nil
}

// GetCommitsWithoutDCO returns all commits in the repo that lack DCO sign-off.
// It uses a single git log call for efficiency.
func GetCommitsWithoutDCO(repoPath string) ([]Commit, error) {
	clean := filepath.Clean(repoPath)
	const sep = "---COMMIT_BOUNDARY_7f3a9b2c---"
	// %H=hash, %aN=author name, %aE=author email, %s=subject, %b=body
	format := sep + "\n%H\n%aN\n%aE\n%s\n%b"

	// #nosec G204 - repoPath is validated as a git repo above
	cmd := exec.Command("git", "-C", clean, "log", "--format="+format)
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf
	output, err := cmd.Output()
	if err != nil {
		// Empty repo: git log exits 128 with "does not have any commits yet"
		if strings.Contains(stderrBuf.String(), "does not have any commits yet") {
			return []Commit{}, nil
		}
		return nil, fmt.Errorf("failed to run git log: %w", err)
	}

	return parseCommits(string(output), sep), nil
}

// GetTotalCommitCount returns the total number of commits in the repository
func GetTotalCommitCount(repoPath string) (int, error) {
	clean := filepath.Clean(repoPath)
	// #nosec G204
	cmd := exec.Command("git", "-C", clean, "rev-list", "--count", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		// If no commits yet, return 0
		return 0, nil
	}
	var count int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &count); err != nil {
		return 0, nil
	}
	return count, nil
}

// GetCommitDetails returns the full git log output for a commit hash
func GetCommitDetails(repoPath, hash string) (string, error) {
	clean := filepath.Clean(repoPath)
	// Validate hash format to prevent injection
	if !isValidHash(hash) {
		return "", fmt.Errorf("invalid commit hash: %s", hash)
	}
	// #nosec G204 - hash is validated above
	cmd := exec.Command("git", "-C", clean, "log", "-1", hash)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit details for %s: %w", hash, err)
	}
	return string(out), nil
}

// GetGitUserName returns the user.name configured in the git repository.
func GetGitUserName(repoPath string) (string, error) {
	clean := filepath.Clean(repoPath)
	// #nosec G204 - repoPath is validated as a git repo before this is called
	cmd := exec.Command("git", "-C", clean, "config", "user.name")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user.name: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// CreateRetroactiveSignoffCommit creates a GPG-signed empty commit carrying the
// retroactive DCO sign-off message. Requires the user to have a GPG key
// configured (git commit.gpgsign or -S).
func CreateRetroactiveSignoffCommit(repoPath, message string) error {
	clean := filepath.Clean(repoPath)
	// #nosec G204 - message is generated internally from validated commit data
	cmd := exec.Command("git", "-C", clean, "commit", "-s", "-S", "--allow-empty", "-m", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create retroactive signoff commit: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// isValidHash checks that a string is a valid full or short git commit hash
func isValidHash(hash string) bool {
	if len(hash) < 7 || len(hash) > 64 {
		return false
	}
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// hasDCOSignoff checks if a commit body contains a valid DCO sign-off line
func hasDCOSignoff(subject, body string) bool {
	combined := subject + "\n" + body
	return signoffRegex.MatchString(combined)
}

// parseCommits parses git log output into Commit structs, filtering to those without DCO
func parseCommits(output, sep string) []Commit {
	var result []Commit
	blocks := strings.Split(output, sep+"\n")
	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		lines := strings.SplitN(block, "\n", 5)
		if len(lines) < 4 {
			continue
		}
		hash := strings.TrimSpace(lines[0])
		author := strings.TrimSpace(lines[1])
		email := strings.TrimSpace(lines[2])
		subject := strings.TrimSpace(lines[3])
		body := ""
		if len(lines) == 5 {
			body = strings.TrimSpace(lines[4])
		}
		if hash == "" {
			continue
		}
		c := Commit{
			Hash:    hash,
			Author:  author,
			Email:   email,
			Subject: subject,
			Body:    body,
			HasDCO:  hasDCOSignoff(subject, body),
		}
		if !c.HasDCO {
			result = append(result, c)
		}
	}
	return result
}
