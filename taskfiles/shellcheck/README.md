# ShellCheck

A [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter) module for [ShellCheck](https://www.shellcheck.net) — a static analysis tool for shell scripts.

## What is this Taskfile?

This module provides tasks to lint, install, and manage [ShellCheck](https://www.shellcheck.net). ShellCheck finds bugs and style issues in Bash/sh scripts, covering syntax errors, quoting mistakes, deprecated constructs, and portability problems.

## Usage

### Standalone

```sh
task -t taskfiles/shellcheck/Taskfile.yml lint
task -t taskfiles/shellcheck/Taskfile.yml lint TARGETS="scripts/*.sh"
task -t taskfiles/shellcheck/Taskfile.yml lint EXTRA_ARGS="--shell=bash --severity=warning"
```

### Included in your Taskfile

```yaml
includes:
  shellcheck:
    taskfile: taskfiles/shellcheck/Taskfile.yml
    vars:
      TARGETS_OVERRIDE: "{{.TARGETS}}"
      EXTRA_ARGS_OVERRIDE: "{{.EXTRA_ARGS}}"
```

Then run:

```sh
task shellcheck:lint
task shellcheck:install
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install ShellCheck on the current operating system |
| `install:undo` | Remove ShellCheck from the current operating system |
| `lint` | Lint shell scripts with ShellCheck (TARGETS=glob) |
| `upgrade` | Upgrade ShellCheck to the latest release |
| `version` | Show the installed ShellCheck version |

## Variables

| Variable | Default | Description |
|---|---|---|
| `EXTRA_ARGS` | `""` | Additional flags passed to `shellcheck` (e.g. `--shell`, `--severity`) |
| `TARGETS` | `""` | Paths or globs of scripts to check; empty = discover all `*.sh` recursively |
| `VERSION` | `""` | Pin a specific shellcheck release for `install`; empty installs latest. Exact availability depends on the platform's package manager/repository. |
| `SHELLCHECK_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

- **macOS** installs via Homebrew (`brew install shellcheck`). Homebrew must be installed.
- **Linux** installs via `apt-get` (Debian/Ubuntu) or `dnf` (Fedora/RHEL). The task dispatches to whichever package manager is present.
- **Windows** installs via Scoop (`scoop install shellcheck`). Scoop must be installed.
- When `TARGETS` is empty, all `*.sh` and `*.bash` files under the working tree are discovered recursively (excluding `.git`).
- Pass explicit paths or globs (e.g. `TARGETS="scripts/*.sh"`) to limit the scope.
- The `lint` task auto-installs ShellCheck if it is not already present in `PATH`.
