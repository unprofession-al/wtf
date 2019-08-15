# `wtf`

`wtf` is a wrapper for `terraform`. It allows to have multiple versions of terraform installed
and helps you to run the version required for your project.

## Features

* Install new (or old) versions of terraform using `wtf install [terraform_version]`.
* Run terraform via `wtf exec ...` (using the regular terraform commands and options) to execute 
`terraform` or use `alias terraform="wtf exec"` for convenience.
* Detect if a project has used pre or post v0.12.0 terraform syntax automatically and run the best 
fitting terraform version automatically.
* Force a certain version by putting your version constraints (such as `>= 0.12.0`) in a 
`.terraform-version` file in your project folder.
* If required you can define a wrapper script template in `wtf`'s configuration file. The template
will be rendered to a temp file and then executed rather than terraform itself.

## Install

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
version_contraint_file_name: .terraform-version
binary_store_path: ~/.bin/terraform.versions/
detect_syntax: true
wrapper:
  script_template: |
    #!/bin/bash
    if [ -f $(pwd)/secrets.yml ]; then
      echo "SUMMON PASSWORDS FOR TERRAFORM"
      summon -p ~/.bin/summon-gopass {{.}}
    else
      echo "RUN PLAIN TERRAFORM"
      {{.}}
    fi
```
