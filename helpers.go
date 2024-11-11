package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	ver "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2/hclsimple"
)

type VersionFile struct {
	Terraform TerraformConfig `hcl:"terraform,block"`
}

type TerraformConfig struct {
	RequiredVersion string `hcl:"required_version"`
}

func readConstraint() (ver.Constraints, error) {
	wd, err := os.Getwd()
	if err != nil {
		return ver.Constraints{}, err
	}

	filename := filepath.Join(wd, "versions.tf")

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return ver.Constraints{}, nil
	} else if err != nil {
		return ver.Constraints{}, err
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return ver.Constraints{}, err
	}

	var versionFile VersionFile
	_ = hclsimple.Decode("c.hcl", data, nil, &versionFile)
	if strings.TrimSpace(versionFile.Terraform.RequiredVersion) != "" {
		return ver.NewConstraint(versionFile.Terraform.RequiredVersion)
	}

	return ver.Constraints{}, err
}

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func expandPath(path string) string {
	dir, _ := os.UserHomeDir()
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}
	return os.ExpandEnv(path)
}
