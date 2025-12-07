package main

import (
	"errors"
	"fmt"
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
	if err := hclsimple.Decode("c.hcl", data, nil, &versionFile); err != nil {
		return ver.Constraints{}, fmt.Errorf("failed to parse versions.tf: %w", err)
	}
	if strings.TrimSpace(versionFile.Terraform.RequiredVersion) != "" {
		return ver.NewConstraint(versionFile.Terraform.RequiredVersion)
	}

	return ver.Constraints{}, nil
}

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func expandPath(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		if path == "~" {
			path = dir
		} else {
			path = filepath.Join(dir, path[2:])
		}
	}
	return os.ExpandEnv(path), nil
}
