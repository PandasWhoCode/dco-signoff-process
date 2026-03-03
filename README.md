# DCO Signoff Process and Script. Provided by LF.

## dcocheck

`dcocheck` is a Go CLI tool that scans a git repository for commits missing
[Developer Certificate of Origin (DCO)](https://developercertificate.org/) sign-off
(`Signed-off-by:` trailer). It displays results in a pageable CLI list and produces
output compatible with the retroactive DCO sign-off process described in [PROCESS.md](PROCESS.md).

### Installation

**Homebrew (macOS / Linux)**

```bash
brew install --formula https://raw.githubusercontent.com/PandasWhoCode/dco-signoff-process/main/Formula/dcocheck.rb
```

**go install**

```bash
go install github.com/PandasWhoCode/dco-signoff-process/dcocheck/cmd/dcocheck@latest
```

**Binary download**

Download the latest release from the [Releases page](https://github.com/PandasWhoCode/dco-signoff-process/releases)
and place the binary in your `PATH`.

### Usage

```bash
# Check current directory
dcocheck .

# Check a specific repository
dcocheck /path/to/repo

# Dry run – show output but exit 0 (no issues reported to CI)
dcocheck --dry-run /path/to/repo

# Write results to a file
dcocheck --output results.txt /path/to/repo

# Show help
dcocheck --help

# Show version
dcocheck --version
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0`  | No DCO issues found (or `--dry-run` mode) |
| `1`  | One or more commits are missing DCO sign-off |
| `2`  | An error occurred (invalid repo path, etc.) |

### Output Sections

When issues are found, `dcocheck` outputs four sections:

1. **Summary of Unique Authors** – deduplicated list of authors with missing sign-off
2. **Summary of Commits** – hash + author for each affected commit
3. **DCO Retroactive Message Format** – ready-to-paste lines for a retroactive sign-off commit message
4. **Full Commit Log History** – full `git log -1` output for each affected commit

See [PROCESS.md](PROCESS.md) for the complete retroactive sign-off workflow.

### Developer Documentation

See [DEVELOPER.md](DEVELOPER.md) for build instructions, testing, and release process.

