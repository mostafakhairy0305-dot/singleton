# cargo

A [TaskOtter](https://github.com/mostafakhairy0305-dot/TaskOtter) module for the
[Rust](https://www.rust-lang.org/) toolchain and its [Cargo](https://doc.rust-lang.org/cargo/)
build tool.

## What is this Taskfile?

This module installs the Rust toolchain with the official
[rustup](https://rustup.rs/) installer — providing `cargo`, `rustc`, `rustfmt`,
and `clippy` — and runs the common Cargo workflow (build, test, check, format,
and lint). macOS and Linux use the `sh.rustup.rs` script; Windows uses
`rustup-init.exe`. Because it installs and manages the toolchain, other modules
can depend on `cargo:install` before running `cargo install <tool>`.

## Usage

### Standalone

```sh
task -t taskfiles/cargo/Taskfile.yml install
task -t taskfiles/cargo/Taskfile.yml build
task -t taskfiles/cargo/Taskfile.yml test
task -t taskfiles/cargo/Taskfile.yml lint
task -t taskfiles/cargo/Taskfile.yml version
```

### Included in your Taskfile

```yaml
includes:
  cargo:
    taskfile: taskfiles/cargo/Taskfile.yml
```

Then run:

```sh
task cargo:install
task cargo:build
task cargo:fmt
task cargo:lint
task cargo:version
```

## Building and testing

Cargo tasks run in the directory where you invoke `task`, so run them from a
crate root (where `Cargo.toml` lives). Pass extra Cargo flags with `EXTRA_ARGS`:

```sh
task cargo:build EXTRA_ARGS=--release
task cargo:test EXTRA_ARGS="-- --nocapture"
task cargo:check
```

## Formatting and linting

```sh
task cargo:fmt              # cargo fmt
task cargo:fmt:check        # cargo fmt --check (reports without writing)
task cargo:lint             # cargo clippy
task cargo:lint:fix         # cargo clippy --fix
task cargo:lint EXTRA_ARGS="-- -D warnings"
```

## Toolchains

Leave `RUST_TOOLCHAIN` empty to install and use the `stable` toolchain. Set it to
a channel or version to pin the toolchain; `install` passes it to rustup as
`--default-toolchain`, and workflow tasks invoke Cargo as `cargo +<toolchain>`:

```sh
task cargo:install RUST_TOOLCHAIN=nightly
task cargo:build RUST_TOOLCHAIN=1.79.0 EXTRA_ARGS=--release
```

## Public Tasks

| Task | Description |
|---|---|
| `install` | Install the Rust toolchain via rustup |
| `install:undo` | Remove the Rust toolchain installed by rustup |
| `upgrade` | Upgrade the Rust toolchain with `rustup update` |
| `build` | Build the crate with `cargo build` |
| `check` | Type-check the crate with `cargo check` |
| `test` | Run the crate test suite with `cargo test` |
| `fmt` | Format Rust code with `cargo fmt` |
| `fmt:check` | Check Rust formatting with `cargo fmt --check` |
| `lint` | Lint Rust code with `cargo clippy` |
| `lint:fix` | Auto-fix Rust lint issues with `cargo clippy --fix` |
| `version` | Show the installed cargo version |
| `which` | Show the path to the cargo binary |
| `verify` | Print cargo and rustc versions and the rustup toolchain overview |

## Variables

| Variable | Default | Description |
|---|---|---|
| `RUSTUP_INSTALL_URL` | `https://sh.rustup.rs` | macOS/Linux rustup installer script URL |
| `RUSTUP_INSTALL_URL_WINDOWS` | `https://win.rustup.rs/x86_64` | Base URL for the Windows `rustup-init.exe` |
| `RUST_TOOLCHAIN` | empty (`stable`) | Optional toolchain channel or version, such as `nightly` or `1.79.0` |
| `CARGO_BIN_UNIX` | `$HOME/.cargo/bin` | Directory containing the rustup-installed binaries, used as a PATH fallback |
| `EXTRA_ARGS` | empty | Extra flags appended to Cargo subcommands |
| `CARGO_LINT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by lint checks and fixes |
| `CARGO_FMT_SKIP_PATTERN` | _(empty)_ | Forward-slash path glob for files skipped by formatting checks and fixes |

Skip patterns support `*` within one path segment, `**` across directories, and `?` for one character. Paths are matched relative to the task working directory; for example, `**/generated/**`.

Cargo cannot exclude individual Rust source files. When a pattern matches any `.rs` file, both lint and format tasks skip its entire containing package.

## Notes

- macOS, Linux, and other Unix-like systems run rustup's documented
  `curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh` installer.
- Windows downloads and runs `rustup-init.exe`.
- rustup adds `~/.cargo/bin` to your shell profile during install. Restart the
  shell or terminal if `cargo` is not immediately available in `PATH`; tasks
  fall back to `CARGO_BIN_UNIX` in the meantime.
- `rustfmt` and `clippy` ship with rustup's default profile, so `fmt` and `lint`
  work out of the box after `install`.
