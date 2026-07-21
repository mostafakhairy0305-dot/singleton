# uv Taskfile Public Tasks

## What is this Taskfile?

A cross-platform Taskfile for installing uv, managing upgrades, and running
common Python project operations — virtual environments, package installation,
script execution, and isolated tool management.

uv is installed via the official Astral install script on macOS and Linux, and
via the official PowerShell installer on Windows. Most operations use `UV_LOAD`
to ensure the uv binary is reachable immediately after installation without
requiring a shell restart.

## Usage

### Standalone

```sh
task -t taskfiles/uv/Taskfile.yml install
task -t taskfiles/uv/Taskfile.yml version
task -t taskfiles/uv/Taskfile.yml tool:install TOOL=yamllint
```

### Included

```yaml
includes:
  uv: ./taskfiles/uv/Taskfile.yml
```

Then run:

```sh
task uv:install
task uv:venv
task uv:tool:install TOOL=ruff
```

## Public Tasks

| Task             | Description                                        | Key variables                |
| ---------------- | -------------------------------------------------- | ---------------------------- |
| `install`        | Install uv on the current OS if missing            | `UV_VERSION`                 |
| `install:undo`   | Remove uv from the current OS                      | none                         |
| `upgrade`        | Upgrade uv to the latest release                   | none                         |
| `version`        | Show the installed uv version                      | none                         |
| `python:install` | Install a Python version via uv                    | `PYTHON_VERSION`             |
| `venv`           | Create a virtual environment                       | `VENV`, `EXTRA_ARGS`         |
| `pip:install`    | Install packages from a requirements file          | `REQUIREMENTS`, `EXTRA_ARGS` |
| `run`            | Run a script or command via uv                     | `FILE`, `ARGS`, `EXTRA_ARGS` |
| `tool:install`   | Install a Python tool into an isolated environment | `TOOL`, `EXTRA_ARGS`         |
| `tool:upgrade`   | Upgrade an installed uv tool                       | `TOOL`, `EXTRA_ARGS`         |

## Variables

| Variable                 | Default                                | Description                                              |
| ------------------------ | -------------------------------------- | -------------------------------------------------------- |
| `VENV`                   | `.venv`                                | Virtual environment directory for `venv`                 |
| `REQUIREMENTS`           | `requirements.txt`                     | Requirements file for `pip:install`                      |
| `FILE`                   | _(empty)_                              | Script path; required by `run`                           |
| `ARGS`                   | _(empty)_                              | Positional arguments forwarded to the script in `run`    |
| `EXTRA_ARGS`             | _(empty)_                              | Extra flags forwarded to the underlying uv command       |
| `PYTHON_VERSION`         | _(empty)_                              | Python version to install; required by `python:install`  |
| `TOOL`                   | _(empty)_                              | Tool name; required by `tool:install` and `tool:upgrade` |
| `UV_INSTALL_URL`         | `https://astral.sh/uv/install.sh`      | Unix installer URL (unversioned, latest)                 |
| `UV_INSTALL_URL_WINDOWS` | `https://astral.sh/uv/install.ps1`     | Windows installer URL (unversioned, latest)               |
| `UV_VERSION`             | _(empty)_                              | Pin a specific uv release for `install`; empty installs latest |
| `UV_LOAD`                | `export PATH="$HOME/.local/bin:$PATH"` | Shell snippet that ensures uv is in PATH                 |

## Notes

**After installation** on macOS and Linux, the uv binary is placed at
`~/.local/bin/uv`. Restart your shell for it to be available, or let tasks
load it automatically via `UV_LOAD`.

**`tool:install`** creates an isolated environment for each tool so their
dependencies never conflict with your project. The tool's binary is shimmed
into `~/.local/bin` (Unix) or `%USERPROFILE%\.local\bin` (Windows).

**`install:undo`** uses `uv self uninstall` which cleanly removes the binary
and its data directory.
