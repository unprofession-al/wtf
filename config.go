package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"gopkg.in/yaml.v3"
)

const configFile = "~/.config/wtf/config.yaml"

type conf struct {
	BinaryStorePath string  `yaml:"binary_store_path"`
	Wrapper         wrapper `yaml:"wrapper"`
}

func NewConfiguration() (*conf, error) {
	c := NewConfigurationDefaults()

	data, err := os.ReadFile(expandPath(configFile))
	if os.IsNotExist(err) {
		fmt.Printf("No config file '%s' found, using defaults\n", configFile)
	} else if err != nil {
		return c, fmt.Errorf("config file '%s' could not be read: %s", configFile, err.Error())
	}
	err = yaml.Unmarshal(data, c)

	return c, err
}

func NewConfigurationDefaults() *conf {
	return &conf{
		BinaryStorePath: "~/.bin/terraform.versions/",
	}
}

func (c *conf) Interactive() error {
	var err error

	prompt := promptui.Prompt{
		Label:   "Binary Store Path ",
		Default: c.BinaryStorePath,
	}
	c.BinaryStorePath, err = prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed %v", err)
	}

	yaml, err := c.ToYAML()
	if err != nil {
		return err
	}
	yaml = append([]byte("---\n"), yaml...)
	fmt.Printf("\nYour configuration is:\n\n%s\n", string(yaml))

	prompt = promptui.Prompt{
		Label:     fmt.Sprintf("Save to %s", expandPath(configFile)),
		IsConfirm: true,
	}
	c.BinaryStorePath, err = prompt.Run()
	if err != nil {
		return fmt.Errorf("prompt failed %v", err)
	}

	// SAVE
	path := expandPath(configFile)
	dir := filepath.Dir(path)
	err = createDir(dir)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	_, err = f.Write(yaml)
	if err != nil {
		return err
	}

	err = f.Sync()
	return err
}

func (c conf) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}
