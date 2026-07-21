# adrs Taskfile

## What is this Taskfile?

This Taskfile wraps [adrs](https://github.com/joshrotenberg/adrs), a
command-line tool for managing Architecture Decision Records, with automation
tasks for installing the tool and creating, listing, and publishing ADRs on
macOS, Linux, and Windows. adrs is installed with `cargo install`, and the
Rust toolchain itself is bootstrapped through the cargo module when missing.

## Usage

### Standalone

```bash
task --taskfile taskfiles/adrs/Taskfile.yml list
```

### Included

```yaml
includes:
  adrs:
    taskfile: taskfiles/adrs/Taskfile.yml
```

```bash
task adrs:install
task adrs:init
task adrs:new -- "Use PostgreSQL for persistence"
task adrs:list
task adrs:generate -- toc
task adrs:exec -- search postgres
```

Pass arguments and flags with `EXTRA_ARGS=...` or after `--`. Pin a release
with `ADRS_VERSION` (e.g. `task adrs:install ADRS_VERSION=0.4.0`).

## Public Tasks

| Task | Description | Key variables |
|---|---|---|
| `install` | Install adrs on the current operating system | `ADRS_VERSION` |
| `install:undo` | Remove adrs (alias: `uninstall`) | |
| `upgrade` | Reinstall adrs at the requested version | `ADRS_VERSION` |
| `init` | Initialize an ADR repository | `EXTRA_ARGS` |
| `new` | Create a new ADR | `EXTRA_ARGS` |
| `list` | List all ADRs | `EXTRA_ARGS` |
| `generate` | Generate ADR docs (`toc`, `graph`, or `book`) | `EXTRA_ARGS` |
| `exec` | Run any adrs subcommand | `EXTRA_ARGS` |
| `version` | Show the installed adrs version | |

## Variables

| Variable | Default | Description |
|---|---|---|
| `ADRS_VERSION` | `""` (latest) | Optional crate version passed to `cargo install --version` |
| `EXTRA_ARGS` | `""` | Arguments and flags appended to the adrs subcommand |
| `CARGO_BIN_UNIX` | `$HOME/.cargo/bin` | Fallback cargo bin directory on macOS and Linux |

## Notes

- Auto-install: every run task depends on `install`, and `install` bootstraps
  the Rust toolchain via the cargo module first, so `task adrs:list` works on
  a fresh machine. Installs are idempotent and version-aware — changing
  `ADRS_VERSION` triggers a reinstall.
- Binaries are resolved from PATH first, falling back to `~/.cargo/bin`
  (`%USERPROFILE%\.cargo\bin` on Windows), so a fresh cargo install works
  without restarting the shell.
