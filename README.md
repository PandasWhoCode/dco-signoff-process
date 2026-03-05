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

```
dcocheck [OPTIONS] [PATH TO REPO]

Options:
  -d, --dry-run        Display results but always exit 0 (suppresses CI failure)
  -o, --output FILE    Write results to FILE in addition to stdout
  -h, --help           Show this help message
  -v, --version        Print version and exit

Arguments:
  PATH TO REPO         Path to the git repository to check (default: current directory)
```

**Examples**

```bash
# Check the current directory
dcocheck .

# Check a specific repository
dcocheck /path/to/repo

# Dry run – show output but exit 0 (suitable for informational CI steps)
dcocheck --dry-run /path/to/repo
dcocheck -d /path/to/repo

# Write results to a file as well as stdout
dcocheck --output results.txt /path/to/repo
dcocheck -o results.txt /path/to/repo

# Show version
dcocheck --version
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0`  | No DCO issues found, retroactive sign-off completed, or `--dry-run` mode |
| `1`  | One or more commits are missing DCO sign-off (no action taken) |
| `2`  | An error occurred (invalid repo path, git failure, etc.) |

### Output Sections

When issues are found, `dcocheck` outputs four sections:

1. **Summary of Unique Authors** – deduplicated list of authors with missing sign-off
2. **Summary of Commits** – hash + author for each affected commit
3. **DCO Retroactive Message Format** – ready-to-paste lines for a retroactive sign-off commit message
4. **Full Commit Log History** – pageable `git log -1` output for each affected commit

### Interactive Retroactive Sign-off

When running in an interactive terminal (and not `--dry-run`), `dcocheck` prompts after displaying
the pageable list:

```
Would you like to perform retroactive DCO sign-off for the listed commits? [y/N]:
```

If you answer `y`, it reads your name from `git config user.name` and creates a GPG-signed empty
commit using:

```
git commit -s -S --allow-empty
```

with the following message format (as specified in [PROCESS.md](PROCESS.md)):

```
I, <First Name> <Last Name>, retroactively sign off on these commits:

commit <short-hash> <subject>
commit <short-hash> <subject>
…
```

> **Note:** Interactive retroactive sign-off requires a GPG key configured for commit signing
> (`git config user.signingkey`). See `gpg --list-secret-keys` and
> `git config --global user.signingkey <keyid>`.

See [PROCESS.md](PROCESS.md) for the complete retroactive sign-off workflow.

### Developer Documentation

See [DEVELOPER.md](DEVELOPER.md) for build instructions, testing, and release process.

