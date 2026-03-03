# Developer Guide

## Prerequisites

- **Go 1.21+** – [Download](https://go.dev/dl/)
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

## Running

```bash
# Check current directory
./bin/dcocheck .

# Check a specific repo
./bin/dcocheck /path/to/repo

# Dry run (show results but exit 0)
./bin/dcocheck --dry-run /path/to/repo

# Write results to a file
./bin/dcocheck --output results.txt /path/to/repo
```

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

Target coverage is ≥95% overall. The `main()` function itself is excluded from coverage
(it calls `os.Exit` and is not unit-testable), but `run()` is fully tested.

## Linting

```bash
task lint
# or:
cd dcocheck && go vet ./...
```

## Releasing a New Version

1. Update the `version` constant in `dcocheck/cmd/dcocheck/main.go`:

   ```go
   const version = "1.2.0"
   ```

2. Commit and push the change.

3. Create and push a git tag:

   ```bash
   git tag -s v1.2.0 -m "Release v1.2.0"
   git push origin v1.2.0
   ```

4. The `.github/workflows/release.yml` workflow will automatically:
   - Run all tests
   - Build binaries for linux/darwin × amd64/arm64
   - Create a GitHub Release with the tag
   - Upload `.tar.gz` archives and `checksums.txt`
   - Update `Formula/dcocheck.rb` with the new version and SHA256 checksums
   - Push the updated formula back to the repository

## Homebrew Formula

`Formula/dcocheck.rb` is updated automatically during the release workflow. The
`PLACEHOLDER_*_SHA256` values are replaced with actual SHA256 checksums of the
release archives.

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
