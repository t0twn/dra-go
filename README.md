# dra-go

> A command line tool to download release assets from GitHub repositories.

A Go port of [devmatteini/dra](https://github.com/devmatteini/dra), optimized for Alpine Docker environments. Produces a fully static binary with no dynamic linking.

## Features

- **No authentication required** for public repositories (optional token support for higher rate limits)
- **Automatic OS/arch detection** — auto-selects the matching asset for your platform
- **Pattern-based selection** — literal names, `{tag}` placeholders, and `*`/`?` wildcards
- **Interactive mode** — pick assets from a numbered list
- **Installation support** — extract and install executables from tar.gz, zip, deb, rpm, and compressed files
- **Shell completions** — bash, zsh, fish, powershell (via Cobra)
- **Alpine-friendly** — `CGO_ENABLED=0` static binary (~8.7 MB)

## Installation

### From source

```bash
# Build
make build

# Install to /usr/local/bin
make install
```

### Cross-compile

```bash
make build-linux-amd64
make build-linux-arm64
make build-linux-arm
make build-macos-amd64
make build-macos-arm64
make build-all              # all platforms
```

### Docker

```bash
make docker
docker run --rm dra:latest download --automatic junegunn/fzf
```

## Usage

### Download (interactive)

```bash
dra download devmatteini/dra-tests
```

### Download (automatic — based on OS/arch)

```bash
dra download --automatic junegunn/fzf
```

### Download by pattern

```bash
# Literal name
dra download --select "myapp-v1.0.0-linux-amd64.tar.gz" owner/repo

# {tag} placeholder (version-free)
dra download --select "myapp-{tag}-linux-amd64.tar.gz" owner/repo

# Wildcard
dra download --select "*linux*amd64*.tar.gz" owner/repo
```

### Download and install

```bash
# Auto-detect executable from archive
dra download --install junegunn/fzf

# Install to custom directory
dra download --install junegunn/fzf -o /usr/local/bin/

# Select specific executables from archive
dra download -s "helloworld-*.tar.gz" -I helloworld -I helper owner/repo
```

### Fetch specific release tag

```bash
dra download --tag v1.2.3 --automatic owner/repo
```

### Custom output path

```bash
dra download --select "myapp.tar.gz" owner/repo -o /tmp/myapp.tar.gz
dra download --select "myapp.tar.gz" owner/repo -o /opt/tools/
```

### Generate untagged pattern

```bash
dra untag owner/repo
# Output: myapp_{tag}-linux-amd64.tar.gz
```

### Shell completions

```bash
eval "$(dra completion bash)"
eval "$(dra completion zsh)"
dra completion fish > ~/.config/fish/completions/dra.fish
```

## Authentication

GitHub tokens are detected automatically from the following environment variables (checked in order):

1. `DRA_GITHUB_TOKEN`
2. `GITHUB_TOKEN`
3. `GH_TOKEN`
4. `gh auth token` (GitHub CLI fallback)

To disable authentication entirely:

```bash
export DRA_DISABLE_GITHUB_AUTHENTICATION=true
```

## Supported install formats

| Format | Method |
|---|---|
| `.tar.gz` / `.tgz` | Extract → find executables → copy |
| `.tar.xz` / `.txz` | Extract (via `ulikunitz/xz`) → find executables → copy |
| `.tar.bz2` / `.tbz` | Extract (via `bzip2` CLI) → find executables → copy |
| `.zip` | Extract → find executables → copy |
| `.gz` / `.xz` / `.bz2` | Decompress single file → set +x |
| `.deb` | `dpkg --install` |
| `.rpm` | `rpm --install --replacepkgs` |
| ELF / Mach-O / `.AppImage` / `.exe` / extensionless | Copy → set +x |

## Architecture

```
internal/
├── github/        # GitHub API client, release/asset types, repository parsing
├── system/        # OS/arch detection, asset matching with priority sorting
├── installer/     # File type detection, archive extraction, executable discovery
├── cli/           # Cobra commands, interactive selection
├── progress/      # Download progress bar (pure Go, ANSI-based)
└── wildcard/      # Custom wildcard matching (* and ?)
```

## Project structure

| Directory | Purpose |
|---|---|
| `internal/github/` | GitHub API client, authentication, release/asset models |
| `internal/system/` | Runtime OS/arch detection, asset name matching, priority ranking |
| `internal/installer/` | File type detection, archive extraction, executable installation |
| `internal/cli/` | Cobra root command, download/untag handlers, interactive prompts |
| `internal/progress/` | Terminal progress bar using ANSI escape codes |
| `internal/wildcard/` | Wildcard pattern matching (avoids `{}` conflicts with `{tag}` syntax) |

## Differences from original Rust version

- **No 7z support** — requires external `7z` binary; rare in Alpine environments
- **Pure Go progress bar** — no dependency on heavy terminal libraries
- **Custom wildcard matcher** — avoids `doublestar`'s `{}` conflict with `{tag}` placeholder
- **Simpler interactive prompt** — uses numbered list instead of arrow-key navigation
- **Bzip2 via CLI** — delegates to `bzip2`/`bzcat` when available; graceful error if missing
- **Same CLI interface** — all flags, subcommands, and behaviors are preserved

## License

MIT — see [LICENSE](./LICENSE). Original project by [Cosimo Matteini](https://github.com/devmatteini/dra).
