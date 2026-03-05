# Developer Guide

## Prerequisites

- **Go 1.25+** – [Download](https://go.dev/dl/)
- **Git** – Required for all DCO checking functionality
- **Task** (optional) – [taskfile.dev](https://taskfile.dev/) for convenience task runner

Install Task via Homebrew:

```bash
brew install go-task
```

## Project Structure

```
dco-signoff-process/
  dcocheck/                        # Go module root
    cmd/
      dcocheck/
        main.go                    # CLI entry point (flag parsing, run())
        main_test.go               # Integration tests for CLI
    internal/
      git/
        git.go                     # Git operations (log, validate, parse)
        git_test.go                # Unit tests for git package
      checker/
        checker.go                 # DCO check logic and result formatting
        checker_test.go            # Unit tests for checker package
      pager/
        pager.go                   # Pageable terminal output
        pager_test.go              # Unit tests for pager package
    go.mod
    go.sum
  Formula/
    dcocheck.rb                    # Homebrew formula (auto-updated on release)
  .github/
    workflows/
      release.yml                  # CI/CD: build, test, release on git tag push
  Taskfile.yml                     # Task runner definitions
  DEVELOPER.md                     # This file
  README.md
  PROCESS.md
```

### Architecture Overview

**`internal/git`** handles all interaction with the `git` binary:
- `ValidateRepo` – confirms a path is a valid git repository
- `GetCommitsWithoutDCO` – runs `git log` and filters commits missing DCO sign-off
- `GetTotalCommitCount` – runs `git rev-list --count HEAD`
- `GetCommitDetails` – fetches full `git log -1 <hash>` output
- `parseCommits` / `hasDCOSignoff` – pure parsing and regex matching functions

**`internal/checker`** orchestrates the check and formats output:
- `Check` – validates repo, fetches commits, collects unique authors
- `FormatSummaryOfAuthors` / `FormatSummaryOfCommits` / `FormatRetroactiveMessage` / `FormatFullCommitLog` – output formatting matching the `reveal_dco_issues` script format
- `AllOutput` – combines all sections

**`internal/pager`** provides pageable terminal display:
- `Display` – prints lines, either all at once (non-interactive) or page-by-page
- `IsTerminal` – detects if stdout is a TTY

**`cmd/dcocheck`** is the CLI entry point:
- Parses flags (`-d`/`--dry-run`, `-o`/`--output`, `-h`/`--help`, `-v`/`--version`)
- Calls `checker.Check`, formats results, invokes `pager.Display`
- Returns exit codes: 0 (no issues), 1 (issues found), 2 (error)

## Building

```bash
# Build binary into bin/dcocheck
task build
# or manually:
cd dcocheck && go build -o ../bin/dcocheck ./cmd/dcocheck
```

Build for all platforms:

```bash
task build-all
# Output in dist/: dcocheck-{linux,darwin}-{amd64,arm64}
```

## Usage

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

### Exit Codes

| Code | Meaning |
|------|---------|
| `0`  | No DCO issues found, retroactive sign-off completed, or `--dry-run` mode |
| `1`  | One or more commits are missing DCO sign-off (no action taken) |
| `2`  | An error occurred (invalid repo path, git failure, etc.) |

## Running

```bash
# Check current directory
./bin/dcocheck .

# Check a specific repo
./bin/dcocheck /path/to/repo

# Dry run (show results but exit 0)
./bin/dcocheck --dry-run /path/to/repo
./bin/dcocheck -d /path/to/repo

# Write results to a file
./bin/dcocheck --output results.txt /path/to/repo
./bin/dcocheck -o results.txt /path/to/repo

# Print version
./bin/dcocheck --version
```

### Interactive Retroactive Sign-off

When `dcocheck` finds non-compliant commits in an interactive terminal (and `--dry-run` is not set),
it prompts:

```
Would you like to perform retroactive DCO sign-off for the listed commits? [y/N]:
```

If you answer `y`, it reads your name from `git config user.name` and creates a GPG-signed empty
commit with the message:

```
I, <First Name> <Last Name>, retroactively sign off on these commits:

commit <short-hash> <subject>
…
```

> **Note:** Requires a GPG key configured for signing — see
> `git config user.signingkey`.

## Running Tests

```bash
# Run all tests
task test
# or:
cd dcocheck && go test ./...

# Run with verbose output
cd dcocheck && go test -v ./...
```

## Test Coverage

```bash
task test-coverage
# or:
cd dcocheck && go test -coverprofile=coverage.out ./...
cd dcocheck && go tool cover -func=coverage.out
cd dcocheck && go tool cover -html=coverage.out -o coverage.html
```

The project maintains **100% statement, branch, and decision coverage** across all packages.
Every code path — including all error branches, the `main()` entry point, and injected-function
test paths — is verified by unit tests.

## Linting

```bash
task lint
# or:
cd dcocheck && go vet ./...
```

## Releasing a New Version

Releases are fully automated via [semantic-release](https://semantic-release.gitbook.io/) and
triggered on every push to `main`. **Do not create tags or edit the version constant manually.**

### How it works

1. **Commit to `main`** using [Conventional Commits](https://www.conventionalcommits.org/):
   - `fix: …` → patch release (e.g. `1.0.0` → `1.0.1`)
   - `feat: …` → minor release (e.g. `1.0.0` → `1.1.0`)
   - `feat!: …` or a commit body containing `BREAKING CHANGE:` → major release (e.g. `1.0.0` → `2.0.0`)
   - Other types (`chore:`, `docs:`, `refactor:`, `test:`, etc.) do **not** trigger a release.

2. **The CI pipeline** (`.github/workflows/release.yml`) runs automatically:
   - **`test` job** – runs `task test` to ensure all tests pass
   - **`release` job** – runs `npx semantic-release`, which:
     - Determines the next version from conventional commit messages
     - Creates and pushes a git tag (e.g. `v1.2.0`)
     - Creates a GitHub Release with generated release notes
   - **`build` job** – only runs when a new release was published:
     - Builds `linux/darwin × amd64/arm64` binaries via `task build-release`
     - Uploads `.tar.gz` archives and `checksums.txt` to the GitHub Release
     - Updates `Formula/dcocheck.rb` and pushes it to the `PandasWhoCode/homebrew-tools` tap via `task update-formula` (uses `HOMEBREW_TAP_PAT` secret)

### Conventional Commit Examples

```bash
git commit -s -S -m "fix: handle repos with no commits gracefully"
git commit -s -S -m "feat: add --since flag to limit commit range"
git commit -s -S -m "feat!: rename --dry-run to --check"   # breaking change – major bump
git commit -s -S -m "chore(deps): bump golang.org/x/term"  # no release
```

## Homebrew Formula

`Formula/dcocheck.rb` is updated automatically during the release workflow. After each
release, the `update-formula` Taskfile task clones the
[PandasWhoCode/homebrew-tools](https://github.com/PandasWhoCode/homebrew-tools) tap
repository, patches the version and SHA256 checksums, and pushes the change to its
`main` branch using the `HOMEBREW_TAP_PAT` secret.

To install via Homebrew after a release:

```bash
brew tap PandasWhoCode/tools
brew install dcocheck
```

To test the formula locally after a release:

```bash
brew install --build-from-source Formula/dcocheck.rb
```

## Security Considerations

- All user-provided paths are sanitized with `filepath.Clean()` before use.
- Commit hashes are validated against a hex-only regex (7–64 characters) before
  being passed to `git` commands, preventing command injection.
- All `git` operations use `exec.Command` (not a shell) to avoid shell injection.
- Output files are created with mode `0600` (owner read/write only).
- The `#nosec G204` comments on `exec.Command` calls document intentional usage
  where inputs have been sanitized.
