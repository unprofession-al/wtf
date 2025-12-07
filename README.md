# `wtf`

`wtf` is a wrapper for `terraform`. It allows to have multiple versions of terraform installed
and helps you to run the version required for your project.

## Features

* Install new (or old) versions of terraform using `wtf install [terraform_version]`.
* Run terraform via `wtf exec ...` (using the regular terraform commands and options) to execute
`terraform` or create a symlink from `terraform` to `wtf` for convenience.
* Ensure that the proper terraform version according to your versions.tf file is used.
* If required you can define a wrapper script template in `wtf`'s configuration file. The template
will be rendered to a temp file and then executed rather than terraform itself.

## Install

### Binary Download

Navigate to [Releases](https://github.com/unprofession-al/wtf/releases), grab
the package that matches your operating system and architecture. Unpack the archive
and put the binary file somewhere in your `$PATH`.

### Nix

`wtf` provides a Nix flake for installation. You can run it directly:

```bash
nix run github:unprofession-al/wtf
```

Or install it in your profile:

```bash
nix profile install github:unprofession-al/wtf
```

To use it in your own flake, add it as an input:

```nix
{
  inputs = {
    wtf.url = "github:unprofession-al/wtf";
  };

  outputs = { self, nixpkgs, wtf, ... }: {
    # Use wtf.packages.${system}.default
  };
}
```

A development shell is also available with Go and golangci-lint:

```bash
nix develop github:unprofession-al/wtf
```

### From Source

Make sure you have [Go](https://golang.org/doc/install) installed, then run:

```bash
go install github.com/unprofession-al/wtf@latest
```

## Run

Just run `wtf` to get some basic help:

```
$ wtf
Wrapper for Terraform: Transparently work with multiple terraform versions

Usage:
  wtf [command]

Available Commands:
  completion    Generate the autocompletion script for the specified shell
  exec          Run correct version of terraform
  help          Help about any command
  install       install a version of terraform
  list-versions list versions of terraform
  version       Print version info

Flags:
  -h, --help   help for wtf

Use "wtf [command] --help" for more information about a command.
```

## Configure

Configuration is stored at `$XDG_CONFIG_HOME/wtf/config.yaml` (defaults to `~/.config/wtf/config.yaml`).

Terraform binaries are stored at `$XDG_DATA_HOME/wtf/terraform-versions/` (defaults to `~/.local/share/wtf/terraform-versions/`).

Here's an example configuration:

```yaml
---
binary_store_path: ~/.local/share/wtf/terraform-versions/
wrapper:
  script_template: |
    #!/bin/bash
    verbose={{.Verbose}}
    if [ -f $(pwd)/secrets.yml ]; then
      if [ "$verbose" = true ]; then
        echo "SUMMON PASSWORDS FOR TERRAFORM"
      fi
      summon -p ~/.bin/summon-gopass {{.Command}}
    else
      if [ "$verbose" = true ]; then
        echo "RUN PLAIN TERRAFORM"
      fi
      {{.Command}}
    fi
```

### Wrapper Script Template Variables

The wrapper script template supports the following variables:

* `{{.TerraformBin}}` - Path to the terraform binary
* `{{.Command}}` - The full command to execute
* `{{.Verbose}}` - Whether verbose mode is enabled
