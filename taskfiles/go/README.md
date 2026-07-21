# Go Taskfile

## What is this Taskfile?

A cross-platform Taskfile for installing the Go toolchain and running Go
formatting, lint, vulnerability, and security checks.

macOS uses Homebrew. Linux uses the official tarball from go.dev and installs it
under `/usr/local/go` by default. Windows uses the official MSI installer from
go.dev. Development tools are installed into `GOBIN`, falling back to
`GOPATH/bin`.

## Usage

### Standalone

```sh
task -t taskfiles/go/Taskfile.yml install
task -t taskfiles/go/Taskfile.yml fmt
task -t taskfiles/go/Taskfile.yml lint
task -t taskfiles/go/Taskfile.yml lint:fix
task -t taskfiles/go/Taskfile.yml version
task -t taskfiles/go/Taskfile.yml verify
```

### Included

```yaml
includes:
  go: ./taskfiles/go/Taskfile.yml
```

Then run:

```sh
task go:install
task go:fmt
task go:lint
task go:lint:fix
task go:version
task go:verify
```

## Linting

Run every configured check:

```sh
task -t taskfiles/go/Taskfile.yml lint
```

When this Taskfile is included under the `go` namespace:

```sh
task go:lint
```

The aggregate task runs:

- `golangci-lint:lint`: runs `golangci-lint run ./...`
- `golangci-lint:fmt:check`: checks Go formatting with `gci`, `gofmt`,
  `gofumpt`, `goimports`, `golines`, and `swaggo`
- `govulncheck:lint`: runs `govulncheck ./...`
- `gosec:lint`: runs `gosec ./...`

Each check depends on its matching `install:<tool>` task, so missing tools are
installed automatically. The formatter check prints diffs and exits with a
nonzero status when files require changes.

Run an individual check or override its default arguments with `--`:

```sh
task -t taskfiles/go/Taskfile.yml gosec:lint
task -t taskfiles/go/Taskfile.yml golangci-lint:lint -- ./internal/...
task go:govulncheck:lint -- -test ./...
```

Set `GO_LINT_SKIP_PATTERN` to exclude matching file paths from golangci-lint,
govulncheck, and gosec analysis. It applies to both `lint` and `lint:fix` and
uses the same shell-style path glob syntax as `GO_FMT_SKIP_PATTERN`; quote the
value so your shell passes it through unchanged:

```sh
task go:lint GO_LINT_SKIP_PATTERN="**/generated/**"
task go:lint:fix GO_LINT_SKIP_PATTERN="**/mocks/*.go"
```

The skip pattern is added to a temporary overlay of the existing
golangci-lint YAML or JSON configuration, so project-specific settings remain
active. govulncheck and gosec operate on packages rather than individual
files, so any package containing a matching file is omitted from those checks.
Use `GO_FMT_SKIP_PATTERN` as well when the same files should be excluded from
formatting.

Auto-fix lint issues that supported tools can rewrite:

```sh
task -t taskfiles/go/Taskfile.yml lint:fix
task go:lint:fix -- ./internal/...
```

`lint:fix` runs `golangci-lint:lint:fix`, then `fmt` so any generated edits are
normalized with the same golangci-lint formatter set. `golangci-lint:lint:fix`
is also available directly when you only want `golangci-lint run --fix`.

## Formatting

Format Go files in place:

```sh
task -t taskfiles/go/Taskfile.yml fmt
task go:fmt
```

The aggregate formatter runs `golangci-lint fmt` with `gci`, `gofmt`,
`gofumpt`, `goimports`, `golines`, and `swaggo` enabled. The formatter defaults
to `.` and accepts CLI arguments after `--`:

```sh
task go:fmt -- ./internal/...
task go:golangci-lint:fmt -- ./taskfiles/go
task go:golangci-lint:fmt:check -- ./taskfiles/go
```

Set `GO_FMT_SKIP_PATTERN` to exclude matching Go file paths from both `fmt` and
`fmt:check`. The value is a shell-style path glob matched against paths with
forward slashes; quote it so your shell does not expand it before Task receives
it:

```sh
task go:fmt GO_FMT_SKIP_PATTERN="**/generated/**"
task go:fmt:check GO_FMT_SKIP_PATTERN="**/mocks/*.go"
```

An empty pattern keeps the default behavior. When a pattern is set, formatter
targets are expanded to `.go` files first and matching paths are omitted.

## Testing

Run unit tests, benchmarks, fuzz targets, and coverage against the current
project:

```sh
task go:test
task go:bench
task go:coverage
task go:fuzz -- -fuzz FuzzName ./internal/parser
```

Each task defaults to the `./...` package pattern (except `fuzz`, which fuzzes
one target in a single package). Pass extra `go test` flags or a narrower target
after `--`:

```sh
task go:test -- -race -run TestName ./internal/...
task go:bench -- -bench BenchmarkName ./internal/parser
```

`coverage` writes a profile to `GO_COVER_PROFILE` (default `coverage.out`) and
prints only packages containing executable statements, ordered from lowest to
highest coverage. Packages with statements and zero coverage are included.
`fuzz` runs a single target for `GO_FUZZTIME` (default `30s`); Go fuzzes one
target in one package per run, so supply the `-fuzz` pattern and package after
`--`:

```sh
task go:coverage GO_COVER_PROFILE=cover.out
task go:fuzz GO_FUZZTIME=60s -- -fuzz FuzzName ./internal/parser
```

## Versions

Use `GO_VERSION` to install a specific Go toolchain release. It must use the
official release name, including the `go` prefix:

```sh
task -t taskfiles/go/Taskfile.yml install GO_VERSION=go1.26.2
task go:install GO_VERSION=go1.26.2
```

When `GO_VERSION` is empty, `install` uses the latest stable Go release. On
macOS, latest uses Homebrew while an explicit version uses the official Go
package. Linux and Windows use official Go downloads for both modes.

Each development tool has its own optional version variable:

```sh
task go:install:golangci-lint GOLANGCI_LINT_VERSION=v2.1.6
task go:install:govulncheck GOVULNCHECK_VERSION=v1.1.4
task go:install:gosec GOSEC_VERSION=v2.22.7
```

An empty tool version defaults to `latest`. Supplying a tool version forces its
installer to run even when the executable already exists.

## Public Tasks

| Task                        | Description                                           | Key variables      |
| --------------------------- | ----------------------------------------------------- | ------------------ |
| `fmt`                       | Format Go files with golangci-lint formatters         | `GO_FMT_SKIP_PATTERN` |
| `fmt:check`                 | Check Go file formatting with golangci-lint formatters | `GO_FMT_SKIP_PATTERN` |
| `install`                   | Install Go on the current operating system if missing | `INSTALL_DIR_UNIX`, `GO_VERSION` |
| `install:undo`              | Remove Go from the current operating system            | `INSTALL_DIR_UNIX` |
| `install:golangci-lint`     | Install golangci-lint into the global Go bin          | `GLOBAL_GO_BIN`, `GOLANGCI_LINT_VERSION` |
| `install:govulncheck`       | Install govulncheck into the global Go bin             | `GLOBAL_GO_BIN`, `GOVULNCHECK_VERSION` |
| `install:gosec`             | Install gosec into the global Go bin                   | `GLOBAL_GO_BIN`, `GOSEC_VERSION` |
| `lint`                      | Run all Go lint and security checks                    | `GO_LINT_SKIP_PATTERN` |
| `lint:fix`                  | Auto-fix Go lint and formatting issues                 | `GO_LINT_SKIP_PATTERN`, `GO_FMT_SKIP_PATTERN` |
| `golangci-lint:lint`        | Lint all Go packages with golangci-lint                | `GO_LINT_SKIP_PATTERN` |
| `golangci-lint:lint:fix`    | Auto-fix Go lint issues with golangci-lint             | `GO_LINT_SKIP_PATTERN` |
| `golangci-lint:fmt`         | Format Go files with golangci-lint formatters         | `GO_FMT_SKIP_PATTERN` |
| `golangci-lint:fmt:check`   | Check Go formatting with golangci-lint formatters      | `GO_FMT_SKIP_PATTERN` |
| `govulncheck:lint`          | Scan Go packages for known vulnerabilities             | `GO_LINT_SKIP_PATTERN` |
| `gosec:lint`                | Scan Go packages for security issues                   | `GO_LINT_SKIP_PATTERN` |
| `upgrade`                   | Upgrade Go to the selected or latest stable release    | `INSTALL_DIR_UNIX`, `GO_VERSION` |
| `version`                   | Show the installed Go version                          | none               |
| `which`                     | Show the path to the Go binary                         | none               |
| `verify`                    | Print Go version, GOROOT, and GOPATH                   | none               |
| `test`                      | Run Go unit tests                                      | none               |
| `bench`                     | Run Go benchmarks                                      | none               |
| `fuzz`                      | Run a Go fuzz target                                   | `GO_FUZZTIME`      |
| `coverage`                  | Run Go tests and report coverage                       | `GO_COVER_PROFILE` |

## Variables

| Variable               | Default                         | Description                                                           |
| ---------------------- | ------------------------------- | --------------------------------------------------------------------- |
| `INSTALL_DIR_UNIX`     | `/usr/local`                    | Parent directory for the Linux tarball install                        |
| `GO_ROOT_UNIX`         | `{{.INSTALL_DIR_UNIX}}/go`      | Linux Go root directory                                               |
| `GO_BIN_UNIX`          | `{{.GO_ROOT_UNIX}}/bin`         | Linux Go binary directory added to shell profiles                     |
| `GO_CMD_UNIX`          | `{{.GO_BIN_UNIX}}/go`           | Linux Go binary path used as a fallback before the shell reloads PATH |
| `GO_VERSION_URL`       | `https://go.dev/VERSION?m=text` | Endpoint used to resolve the latest stable Go version                 |
| `GO_DOWNLOAD_BASE_URL` | `https://go.dev/dl`             | Base URL for official Go downloads                                    |
| `GO_VERSION`           | empty (latest stable)           | Optional official Go release name, such as `go1.26.2`                 |
| `GOLANGCI_LINT_VERSION` | empty (`latest`)               | Optional golangci-lint module version                                 |
| `GOVULNCHECK_VERSION`  | empty (`latest`)                | Optional govulncheck module version                                   |
| `GOSEC_VERSION`        | empty (`latest`)                | Optional gosec module version                                         |
| `GO_FMT_SKIP_PATTERN`  | empty                           | Shell-style path glob for Go files skipped by `fmt` and `fmt:check`   |
| `GO_LINT_SKIP_PATTERN` | empty                           | Shell-style path glob for Go files skipped by `lint` and `lint:fix`   |
| `GO_COVER_PROFILE`     | empty (`coverage.out`)          | Output path for the `coverage` profile file                          |
| `GO_FUZZTIME`          | empty (`30s`)                   | Duration a single `fuzz` target runs before stopping                 |
| `GLOBAL_GO_BIN`        | `GOBIN` or `GOPATH/bin`         | Destination and lookup directory for installed Go development tools   |

## Notes

Linux installs replace `INSTALL_DIR_UNIX/go`. The task uses `sudo` when it is
not already running as root, then adds `GO_BIN_UNIX` to the current user's shell
profile if Go is not already available on PATH.

Downloaded Go archives are checked against the official `.sha256` published
alongside each release, and the new toolchain is extracted and smoke-tested in a
temporary directory before it replaces `INSTALL_DIR_UNIX/go`. A failed download,
a checksum mismatch, or a bad archive therefore leaves the existing installation
untouched.

macOS requires Homebrew to already be installed.
