# actionlint

A [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter) module for [actionlint](https://github.com/rhysd/actionlint) — a static checker for GitHub Actions workflow files.

## What is this Taskfile?

This module provides tasks to lint, install, and manage [actionlint](https://github.com/rhysd/actionlint). actionlint statically checks GitHub Actions workflow files for syntax errors, type mismatches in expressions, incorrect event payloads, and more.

## Usage

### Standalone

```sh
task -t taskfiles/actionlint/Taskfile.yml lint
task -t taskfiles/actionlint/Taskfile.yml lint ACTIONLINT_TARGETS=.github/workflows/ci.yml
task -t taskfiles/actionlint/Taskfile.yml lint ACTIONLINT_EXTRA_ARGS="-ignore 'label.*'"
```

### Included in your Taskfile

```yaml
includes:
  actionlint:
    taskfile: taskfiles/actionlint/Taskfile.yml
    vars:
      ACTIONLINT_TARGETS_OVERRIDE: "{{.ACTIONLINT_TARGETS}}"
      ACTIONLINT_EXTRA_ARGS_OVERRIDE: "{{.ACTIONLINT_EXTRA_ARGS}}"
```

Then run:

```sh
task actionlint:lint
task actionlint:install
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install actionlint on the current operating system |
| `install:undo` | Remove actionlint from the current operating system |
| `lint` | Lint GitHub Actions workflow files with actionlint |
| `upgrade` | Upgrade actionlint to the latest release |
| `version` | Show the installed actionlint version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `ACTIONLINT_VERSION` | `"1.7.12"` | Pinned version used for Linux binary download |
| `ACTIONLINT_EXTRA_ARGS` | `""` | Additional flags passed to `actionlint` (e.g. `-ignore`, `-format`) |
| `ACTIONLINT_TARGETS` | `""` | Paths to workflow files; empty = auto-discover `.github/workflows` |
| `ACTIONLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- **macOS** installs via Homebrew (`brew install actionlint`). Homebrew must be installed.
- **Linux** downloads a pinned binary from GitHub Releases into `/usr/local/bin`. Requires `curl`, `tar`, and `install`. Only `x86_64` and `aarch64` architectures are supported.
- **Windows** installs via Scoop (`scoop install actionlint`). Scoop must be installed.
- When `TARGETS` is empty, actionlint automatically discovers all files under `.github/workflows/` in the current working directory.
- The `lint` task auto-installs actionlint if it is not already present in `PATH`.
