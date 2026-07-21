# Staticcheck Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing and running
[Staticcheck](https://staticcheck.dev), the Go static analysis tool.

Staticcheck is downloaded from the pinned GitHub release and installed into the
global Go bin directory (`GOBIN`, or `GOPATH/bin` when `GOBIN` is unset). The
`lint` task also ensures the Go toolchain is installed through the local Go
Taskfile before analysis starts.

## Usage

### Standalone

```sh
task -t taskfiles/staticcheck/Taskfile.yml install
task -t taskfiles/staticcheck/Taskfile.yml lint
task -t taskfiles/staticcheck/Taskfile.yml version
```

Pass Staticcheck arguments after `--`:

```sh
task -t taskfiles/staticcheck/Taskfile.yml lint -- ./cmd/... ./internal/...
```

### Included

```yaml
includes:
  staticcheck: ./taskfiles/staticcheck/Taskfile.yml
```

Then run:

```sh
task staticcheck:lint
task staticcheck:version
```

## Public Tasks

| Task           | Description                                           | Key variables                          |
| -------------- | ----------------------------------------------------- | -------------------------------------- |
| `install`      | Install the pinned Staticcheck binary into the Go bin | `STATICCHECK_VERSION`, `GLOBAL_GO_BIN` |
| `install:undo` | Remove the Staticcheck binary from the Go bin         | `GLOBAL_GO_BIN`                        |
| `upgrade`      | Re-download and install the pinned STATICCHECK_VERSION | `STATICCHECK_VERSION`, `GLOBAL_GO_BIN` |
| `lint`    | Run Staticcheck against Go packages                   | none (pass args via `--`)              |
| `version` | Print the installed Staticcheck version               | none                                   |

## Variables

| Variable                       | Default                                                  | Description                                                               |
| ------------------------------ | -------------------------------------------------------- | ------------------------------------------------------------------------- |
| `STATICCHECK_VERSION`          | `2026.1`                                                 | Staticcheck release tag to download and enforce                           |
| `STATICCHECK_RELEASE_BASE_URL` | `https://github.com/dominikh/go-tools/releases/download` | Base URL for Staticcheck release assets                                   |
| `GLOBAL_GO_BIN`                | `$(go env GOBIN)` or `$(go env GOPATH)/bin`              | Go bin directory where the Staticcheck binary is installed                |
| `STATICCHECK_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Staticcheck operates on Go packages. When a pattern matches any `.go` file, the entire containing package is omitted.
