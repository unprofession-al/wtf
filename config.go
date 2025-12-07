package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func getConfigFile() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, _ := os.UserHomeDir()
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, "wtf", "config.yaml")
}

func getDefaultDataDir() string {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, _ := os.UserHomeDir()
		dataHome = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(dataHome, "wtf", "terraform-versions")
}

type conf struct {
	BinaryStorePath string  `yaml:"binary_store_path"`
	Wrapper         wrapper `yaml:"wrapper"`
}

func NewConfiguration() (*conf, error) {
	c := NewConfigurationDefaults()

	configFile := getConfigFile()
	data, err := os.ReadFile(configFile)
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
		BinaryStorePath: getDefaultDataDir(),
	}
}
