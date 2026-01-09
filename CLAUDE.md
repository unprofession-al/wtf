# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Development Commands

```bash
# Build
go build

# Run tests
go test ./...

# Run a single test
go test -run TestFindLatest ./...

# Lint (requires golangci-lint)
golangci-lint run

# Install locally
go install
```

### Nix Development

A Nix flake is available for building and development:

```bash
nix build           # Build the package
nix develop         # Enter dev shell with Go and golangci-lint
```

## Architecture

`wtf` (Wrapper for Terraform) manages multiple terraform versions and transparently runs the correct version based on your project's `versions.tf` file.

### Core Components

- **main.go** - Entry point; detects if invoked as `terraform` symlink (direct passthrough) or as `wtf` (CLI mode)
- **cli.go** - Cobra CLI commands: `exec`, `install`, `list-versions`, `version`
- **terraform.go** - `Terraform` struct managing version discovery, downloading (with SHA256 verification), and execution
- **wrapper.go** - Optional wrapper script templating system that can intercept terraform calls
- **config.go** - Configuration loading from `$XDG_CONFIG_HOME/wtf/config.yaml`
- **helpers.go** - Utilities including `readConstraint()` which parses `versions.tf` HCL files

### Key Flow

1. When run, `wtf` reads `versions.tf` in current directory to get `required_version` constraint
2. Finds the latest installed terraform binary matching that constraint
3. Either runs terraform directly or through a configured wrapper script template

### Version Management

- Terraform binaries stored at `$XDG_DATA_HOME/wtf/terraform-versions/` (default: `~/.local/share/wtf/terraform-versions/`)
- Downloads verified against HashiCorp's official SHA256SUMS

## Testing Guidelines

- Use table-driven tests (see `terraform_test.go`, `helpers_test.go` for examples)
- Tests create temp directories and change working directory for isolation
- Helper functions `mustVersions()` and `mustConstraint()` create test fixtures
