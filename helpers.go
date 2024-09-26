package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	ver "github.com/hashicorp/go-version"
)

const versionFile = ".terraform-version"

func readConstraint(detectConstraint bool) (ver.Constraints, error) {
	wd, err := os.Getwd()
	if err != nil {
		return ver.Constraints{}, err
	}

	filename := fmt.Sprintf("%s/%s", wd, versionFile)
	data, err := os.ReadFile(filename)
	if err == nil {
		return ver.NewConstraint(">= 0.0.0")
	}

	return ver.NewConstraint(string(data))
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
