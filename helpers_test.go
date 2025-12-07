package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get user home dir: %v", err)
	}

	// Set a test environment variable
	os.Setenv("TEST_WTF_VAR", "testvalue")
	defer os.Unsetenv("TEST_WTF_VAR")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "tilde alone expands to home directory",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "tilde with path expands correctly",
			input:    "~/some/path",
			expected: filepath.Join(homeDir, "some/path"),
		},
		{
			name:     "tilde with single segment",
			input:    "~/dir",
			expected: filepath.Join(homeDir, "dir"),
		},
		{
			name:     "absolute path unchanged",
			input:    "/absolute/path",
			expected: "/absolute/path",
		},
		{
			name:     "relative path unchanged",
			input:    "relative/path",
			expected: "relative/path",
		},
		{
			name:     "environment variable expanded",
			input:    "$TEST_WTF_VAR/path",
			expected: "testvalue/path",
		},
		{
			name:     "tilde and env var combined",
			input:    "~/$TEST_WTF_VAR",
			expected: filepath.Join(homeDir, "testvalue"),
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "tilde in middle not expanded",
			input:    "/path/~/other",
			expected: "/path/~/other",
		},
		{
			name:     "undefined env var becomes empty",
			input:    "$UNDEFINED_WTF_VAR/path",
			expected: "/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandPath(tt.input)
			if err != nil {
				t.Errorf("expandPath(%q) returned unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestReadConstraint(t *testing.T) {
	tests := []struct {
		name              string
		versionsContent   string
		createFile        bool
		expectedEmpty     bool
		expectedConstraint string
		expectError       bool
		errorContains     string
	}{
		{
			name:              "valid constraint parses correctly",
			versionsContent:   `terraform { required_version = ">= 1.0.0" }`,
			createFile:        true,
			expectedEmpty:     false,
			expectedConstraint: ">= 1.0.0",
			expectError:       false,
		},
		{
			name:              "pessimistic constraint",
			versionsContent:   `terraform { required_version = "~> 1.5.0" }`,
			createFile:        true,
			expectedEmpty:     false,
			expectedConstraint: "~> 1.5.0",
			expectError:       false,
		},
		{
			name:              "exact version constraint",
			versionsContent:   `terraform { required_version = "= 1.2.3" }`,
			createFile:        true,
			expectedEmpty:     false,
			expectedConstraint: "= 1.2.3",
			expectError:       false,
		},
		{
			name:              "combined constraints",
			versionsContent:   `terraform { required_version = ">= 1.0.0, < 2.0.0" }`,
			createFile:        true,
			expectedEmpty:     false,
			expectedConstraint: ">= 1.0.0, < 2.0.0",
			expectError:       false,
		},
		{
			name:            "no versions.tf file returns empty constraint",
			createFile:      false,
			expectedEmpty:   true,
			expectError:     false,
		},
		{
			name:              "empty required_version returns empty constraint",
			versionsContent:   `terraform { required_version = "" }`,
			createFile:        true,
			expectedEmpty:     true,
			expectError:       false,
		},
		{
			name:              "whitespace-only required_version returns empty constraint",
			versionsContent:   `terraform { required_version = "   " }`,
			createFile:        true,
			expectedEmpty:     true,
			expectError:       false,
		},
		{
			name:            "invalid HCL syntax returns error",
			versionsContent: `terraform { required_version = }`,
			createFile:      true,
			expectError:     true,
			errorContains:   "failed to parse",
		},
		{
			name:            "invalid constraint value returns error",
			versionsContent: `terraform { required_version = "not-a-constraint" }`,
			createFile:      true,
			expectError:     true,
		},
		{
			name: "versions.tf with only required_version",
			versionsContent: `
terraform {
  required_version = ">= 1.3.0"
}
`,
			createFile:         true,
			expectedEmpty:      false,
			expectedConstraint: ">= 1.3.0",
			expectError:        false,
		},
		{
			// NOTE: This documents a limitation - hclsimple doesn't support required_providers block
			name: "versions.tf with required_providers fails",
			versionsContent: `
terraform {
  required_version = ">= 1.3.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}
`,
			createFile:    true,
			expectError:   true,
			errorContains: "Unsupported block type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temp directory
			tmpDir, err := os.MkdirTemp("", "wtf-test-*")
			if err != nil {
				t.Fatalf("could not create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Save current working directory and restore after test
			originalWd, err := os.Getwd()
			if err != nil {
				t.Fatalf("could not get working directory: %v", err)
			}
			defer func() { _ = os.Chdir(originalWd) }()

			// Change to temp directory
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatalf("could not change to temp dir: %v", err)
			}

			// Create versions.tf if needed
			if tt.createFile {
				err := os.WriteFile("versions.tf", []byte(tt.versionsContent), 0644)
				if err != nil {
					t.Fatalf("could not write versions.tf: %v", err)
				}
			}

			// Call readConstraint
			constraint, err := readConstraint()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !containsSubstring(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectedEmpty {
				if len(constraint) != 0 {
					t.Errorf("expected empty constraint, got %v", constraint)
				}
				return
			}

			if constraint.String() != tt.expectedConstraint {
				t.Errorf("readConstraint() = %q, want %q", constraint.String(), tt.expectedConstraint)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
