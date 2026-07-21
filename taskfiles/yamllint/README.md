# yamllint Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing yamllint, managing upgrades, linting
YAML files, auto-fixing style issues with yamlfix, and generating a project
configuration.

macOS uses Homebrew. Linux installs via apt or dnf, falling back to pip3 when
neither package manager is present. Windows uses pip.

## Usage

### Standalone

```sh
task -t taskfiles/yamllint/Taskfile.yml install
task -t taskfiles/yamllint/Taskfile.yml config:init
task -t taskfiles/yamllint/Taskfile.yml lint
```

### Included

```yaml
includes:
  yamllint: ./taskfiles/yamllint/Taskfile.yml
```

Then run:

```sh
task yamllint:install
task yamllint:lint
task yamllint:ci
```

## Public Tasks

| Task           | Description                                     | Key variables                     |
| -------------- | ----------------------------------------------- | --------------------------------- |
| `install`      | Install yamllint on the current OS if missing   | `YAMLLINT_VERSION`                |
| `install:undo` | Remove yamllint from the current OS             | none                              |
| `upgrade`      | Upgrade yamllint to the latest release          | none                              |
| `version`      | Show the installed yamllint version             | none                              |
| `lint`         | Lint YAML files                                 | `TARGETS`, `CONFIG`, `EXTRA_ARGS` |
| `lint:fix`     | Auto-fix YAML files with yamlfix                | `TARGETS`, `EXTRA_ARGS`           |
| `ci`           | Strict lint for CI (fails on warnings)          | `TARGETS`, `CONFIG`, `EXTRA_ARGS` |
| `config:init`  | Create a default `.yamllint` configuration file | none                              |

## Variables

| Variable     | Default   | Description                                      |
| ------------ | --------- | ------------------------------------------------ |
| `TARGETS`    | `.`       | Files or directories to lint                     |
| `CONFIG`     | _(empty)_ | Path to a yamllint config file passed via `-c`   |
| `EXTRA_ARGS` | _(empty)_ | Extra flags forwarded to `yamllint` or `yamlfix` |
| `YAMLLINT_VERSION` | _(empty)_ | Pin a specific yamllint release for `install`/`upgrade`; empty installs latest |
| `YAMLFIX_VERSION` | _(empty)_ | Pin a specific yamlfix release for `lint:fix`; empty installs latest |
| `YAMLLINT_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

## Notes

**macOS:** Requires Homebrew. Run `brew install yamllint` or let `install` do it.

**Linux:** Installation priority is apt → dnf → pip3. On systems where only pip3
is available (e.g. Alpine, Arch), pip3 is used automatically as a fallback.

**Windows:** pip is required. Install Python first, then run `task install`.

**`lint:fix`** uses [yamlfix](https://github.com/lyz-code/yamlfix), which is a
separate tool from yamllint. Install it before using `lint:fix`:

```sh
pip install yamlfix
```

**`config:init`** writes a `.yamllint` file in the current directory and is
skipped if the file already exists. To regenerate, delete `.yamllint` first.
