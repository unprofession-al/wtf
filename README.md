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
the package that matches your operating system and achitecture. Unpack the archive
and put the binary file somewhere in your `$PATH`

### From Source

Make sure you have [go](https://golang.org/doc/install) installed, then run: 

```
# go get -u https://github.com/unprofession-al/wtf
```

## Run

Just run `wtf` to get some basic help:

```
# wtf 
Wrapper for Terraform: tansparently work with multiple terrafrom versions

Usage:
  wtf [command]

Available Commands:
  config      print configuration
  configure   configure interactively
  exec        run correct version of terraform
  help        Help about any command
  install     install a version of terraform

Flags:
  -h, --help   help for wtf

Use "wtf [command] --help" for more information about a command
```

## Configure

Basic options can be configured in the configuration found at `cat ~/.config/wtf/config.yaml`.
Heres an example:

```
---
binary_store_path: ~/.bin/terraform.versions/
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
