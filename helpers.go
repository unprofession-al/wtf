package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

const versionFile = ".terraform-version"

func readConstraint(detectConstraint bool) (version.Constraints, error) {
	wd, err := os.Getwd()
	if err != nil {
		return version.Constraints{}, err
	}

	filename := fmt.Sprintf("%s/%s", wd, versionFile)
	data, err := ioutil.ReadFile(filename)
	if err != nil && detectConstraint {
		return detectSyntax()
	} else if err != nil {
		return version.NewConstraint(">= 0.0.0")
	}

	return version.NewConstraint(string(data))
}

func detectSyntax() (version.Constraints, error) {
	wd, err := os.Getwd()
	if err != nil {
		return version.Constraints{}, err
	}

	matches, err := filepath.Glob(fmt.Sprintf("%s/*.tf", wd))
	if err != nil {
		return version.Constraints{}, err
	}

	var all []byte
	for _, file := range matches {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return version.Constraints{}, err
		}
		all = append(all, data...)
	}

	matched, err := regexp.Match(`"\${.*}"`, all)
	if err != nil {
		return version.Constraints{}, err
	}

	if matched {
		return version.NewConstraint("< 0.12.0")
	}

	return version.NewConstraint(">= 0.12.0")
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
