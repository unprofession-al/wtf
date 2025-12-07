package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home dir: %v", err)
	}

	tests := []struct {
		name          string
		xdgConfigHome string
		expected      string
	}{
		{
			name:          "uses XDG_CONFIG_HOME when set",
			xdgConfigHome: "/custom/config",
			expected:      "/custom/config/wtf/config.yaml",
		},
		{
			name:          "falls back to ~/.config when XDG_CONFIG_HOME is empty",
			xdgConfigHome: "",
			expected:      filepath.Join(homeDir, ".config", "wtf", "config.yaml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env and restore after test
			originalXDG := os.Getenv("XDG_CONFIG_HOME")
			defer os.Setenv("XDG_CONFIG_HOME", originalXDG)

			os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)

			result := getConfigFile()
			if result != tt.expected {
				t.Errorf("getConfigFile() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetDefaultDataDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home dir: %v", err)
	}

	tests := []struct {
		name        string
		xdgDataHome string
		expected    string
	}{
		{
			name:        "uses XDG_DATA_HOME when set",
			xdgDataHome: "/custom/data",
			expected:    "/custom/data/wtf/terraform-versions",
		},
		{
			name:        "falls back to ~/.local/share when XDG_DATA_HOME is empty",
			xdgDataHome: "",
			expected:    filepath.Join(homeDir, ".local", "share", "wtf", "terraform-versions"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env and restore after test
			originalXDG := os.Getenv("XDG_DATA_HOME")
			defer os.Setenv("XDG_DATA_HOME", originalXDG)

			os.Setenv("XDG_DATA_HOME", tt.xdgDataHome)

			result := getDefaultDataDir()
			if result != tt.expected {
				t.Errorf("getDefaultDataDir() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNewConfigurationDefaults(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home dir: %v", err)
	}

	// Save original env and restore after test
	originalXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", originalXDG)
	os.Setenv("XDG_DATA_HOME", "")

	config := NewConfigurationDefaults()

	expectedPath := filepath.Join(homeDir, ".local", "share", "wtf", "terraform-versions")
	if config.BinaryStorePath != expectedPath {
		t.Errorf("BinaryStorePath = %q, want %q", config.BinaryStorePath, expectedPath)
	}

	if config.Wrapper.ScriptTemplate != "" {
		t.Errorf("Wrapper.ScriptTemplate should be empty by default, got %q", config.Wrapper.ScriptTemplate)
	}
}

func TestNewConfiguration(t *testing.T) {
	tests := []struct {
		name               string
		configContent      string
		createFile         bool
		expectedStorePath  string
		expectedWrapper    string
		expectError        bool
		expectDefaultPath  bool
	}{
		{
			name:              "uses defaults when config file does not exist",
			createFile:        false,
			expectDefaultPath: true,
			expectError:       false,
		},
		{
			name: "reads binary_store_path from config",
			configContent: `
binary_store_path: /custom/terraform/versions
`,
			createFile:        true,
			expectedStorePath: "/custom/terraform/versions",
			expectError:       false,
		},
		{
			name: "reads wrapper template from config",
			configContent: `
wrapper:
  script_template: |
    #!/bin/bash
    exec {{.TerraformBin}}
`,
			createFile:        true,
			expectDefaultPath: true,
			expectedWrapper:   "#!/bin/bash\nexec {{.TerraformBin}}\n",
			expectError:       false,
		},
		{
			name: "reads full config",
			configContent: `
binary_store_path: /my/terraform
wrapper:
  script_template: "exec {{.Command}}"
`,
			createFile:        true,
			expectedStorePath: "/my/terraform",
			expectedWrapper:   "exec {{.Command}}",
			expectError:       false,
		},
		{
			name:          "invalid YAML returns error",
			configContent: `binary_store_path: [invalid`,
			createFile:    true,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp directory for config
			tmpDir, err := os.MkdirTemp("", "wtf-config-test-*")
			if err != nil {
				t.Fatalf("could not create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Save original env and restore after test
			originalXDGConfig := os.Getenv("XDG_CONFIG_HOME")
			originalXDGData := os.Getenv("XDG_DATA_HOME")
			defer func() {
				os.Setenv("XDG_CONFIG_HOME", originalXDGConfig)
				os.Setenv("XDG_DATA_HOME", originalXDGData)
			}()

			// Set XDG_CONFIG_HOME to our temp dir
			os.Setenv("XDG_CONFIG_HOME", tmpDir)
			os.Setenv("XDG_DATA_HOME", tmpDir)

			// Create config file if needed
			if tt.createFile {
				configDir := filepath.Join(tmpDir, "wtf")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("could not create config dir: %v", err)
				}
				configFile := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configFile, []byte(tt.configContent), 0644); err != nil {
					t.Fatalf("could not write config file: %v", err)
				}
			}

			config, err := NewConfiguration()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectDefaultPath {
				expectedDefault := filepath.Join(tmpDir, "wtf", "terraform-versions")
				if config.BinaryStorePath != expectedDefault {
					t.Errorf("BinaryStorePath = %q, want default %q", config.BinaryStorePath, expectedDefault)
				}
			} else if config.BinaryStorePath != tt.expectedStorePath {
				t.Errorf("BinaryStorePath = %q, want %q", config.BinaryStorePath, tt.expectedStorePath)
			}

			if tt.expectedWrapper != "" {
				if !strings.Contains(config.Wrapper.ScriptTemplate, strings.TrimSpace(tt.expectedWrapper)) &&
					config.Wrapper.ScriptTemplate != tt.expectedWrapper {
					t.Errorf("Wrapper.ScriptTemplate = %q, want %q", config.Wrapper.ScriptTemplate, tt.expectedWrapper)
				}
			}
		})
	}
}
