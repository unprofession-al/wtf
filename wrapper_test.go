package main

import (
	"os"
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	tests := []struct {
		name           string
		template       string
		command        string
		args           []string
		verbose        bool
		expectOriginal bool // if true, expect original command/args returned
		expectError    bool
		errorContains  string
		checkContent   func(t *testing.T, content string)
	}{
		{
			name:           "empty template returns original command",
			template:       "",
			command:        "/usr/bin/terraform",
			args:           []string{"plan", "-out=tfplan"},
			verbose:        false,
			expectOriginal: true,
			expectError:    false,
		},
		{
			name:           "template with TerraformBin variable",
			template:       "#!/bin/bash\nexec {{.TerraformBin}} \"$@\"",
			command:        "/path/to/terraform",
			args:           []string{"apply"},
			verbose:        false,
			expectOriginal: false,
			expectError:    false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "/path/to/terraform") {
					t.Error("expected TerraformBin to be expanded in template")
				}
			},
		},
		{
			name:           "template with Command variable",
			template:       "#!/bin/bash\n{{.Command}}",
			command:        "/bin/terraform",
			args:           []string{"init", "-upgrade"},
			verbose:        false,
			expectOriginal: false,
			expectError:    false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "/bin/terraform init -upgrade") {
					t.Errorf("expected full command in template, got: %s", content)
				}
			},
		},
		{
			name:           "template with Verbose variable true",
			template:       "#!/bin/bash\n{{if .Verbose}}set -x{{end}}\nexec {{.TerraformBin}}",
			command:        "/bin/terraform",
			args:           []string{},
			verbose:        true,
			expectOriginal: false,
			expectError:    false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "set -x") {
					t.Error("expected Verbose condition to render set -x")
				}
			},
		},
		{
			name:           "template with Verbose variable false",
			template:       "#!/bin/bash\n{{if .Verbose}}set -x{{end}}\nexec {{.TerraformBin}}",
			command:        "/bin/terraform",
			args:           []string{},
			verbose:        false,
			expectOriginal: false,
			expectError:    false,
			checkContent: func(t *testing.T, content string) {
				if strings.Contains(content, "set -x") {
					t.Error("expected Verbose condition to NOT render set -x")
				}
			},
		},
		{
			name:           "invalid template syntax returns error",
			template:       "{{.Invalid",
			command:        "/bin/terraform",
			args:           []string{},
			verbose:        false,
			expectOriginal: false,
			expectError:    true,
			errorContains:  "unclosed action",
		},
		{
			name:           "template with undefined variable",
			template:       "{{.UndefinedVar}}",
			command:        "/bin/terraform",
			args:           []string{},
			verbose:        false,
			expectOriginal: false,
			expectError:    true,
		},
		{
			name:           "complex template with all variables",
			template:       "#!/bin/bash\n# Verbose: {{.Verbose}}\n# Binary: {{.TerraformBin}}\n{{.Command}}",
			command:        "/opt/terraform/1.5.0",
			args:           []string{"plan", "-var", "foo=bar"},
			verbose:        true,
			expectOriginal: false,
			expectError:    false,
			checkContent: func(t *testing.T, content string) {
				if !strings.Contains(content, "Verbose: true") {
					t.Error("expected Verbose: true in template")
				}
				if !strings.Contains(content, "Binary: /opt/terraform/1.5.0") {
					t.Error("expected Binary path in template")
				}
				if !strings.Contains(content, "/opt/terraform/1.5.0 plan -var foo=bar") {
					t.Errorf("expected full command in template, got: %s", content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &wrapper{ScriptTemplate: tt.template}
			defer func() { _ = w.Cleanup() }() // ensure cleanup after test

			cmd, args, err := w.Wrap(tt.command, tt.args, tt.verbose)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.expectOriginal {
				if cmd != tt.command {
					t.Errorf("expected original command %q, got %q", tt.command, cmd)
				}
				if len(args) != len(tt.args) {
					t.Errorf("expected %d args, got %d", len(tt.args), len(args))
				}
				for i, arg := range args {
					if arg != tt.args[i] {
						t.Errorf("arg[%d] = %q, want %q", i, arg, tt.args[i])
					}
				}
				return
			}

			// For non-original, we expect a temp file
			if w.tmpfile == nil {
				t.Error("expected tmpfile to be set for template wrapper")
				return
			}

			// Verify file exists and is executable
			info, err := os.Stat(cmd)
			if err != nil {
				t.Errorf("temp file should exist: %v", err)
				return
			}
			if info.Mode().Perm()&0100 == 0 {
				t.Error("temp file should be executable")
			}

			// Verify args are empty when using template
			if len(args) != 0 {
				t.Errorf("expected empty args for template wrapper, got %v", args)
			}

			// Check content if provided
			if tt.checkContent != nil {
				content, err := os.ReadFile(cmd)
				if err != nil {
					t.Errorf("could not read temp file: %v", err)
					return
				}
				tt.checkContent(t, string(content))
			}
		})
	}
}

func TestWrapperCleanup(t *testing.T) {
	tests := []struct {
		name        string
		setupTmpfile bool
		expectError bool
	}{
		{
			name:        "cleanup with no tmpfile",
			setupTmpfile: false,
			expectError: false,
		},
		{
			name:        "cleanup removes tmpfile",
			setupTmpfile: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &wrapper{}

			var tmpPath string
			if tt.setupTmpfile {
				// Create a real temp file via Wrap
				w.ScriptTemplate = "#!/bin/bash\necho test"
				_, _, err := w.Wrap("/bin/test", nil, false)
				if err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				tmpPath = w.tmpfile.Name()

				// Verify file exists before cleanup
				if _, err := os.Stat(tmpPath); err != nil {
					t.Fatalf("tmpfile should exist before cleanup: %v", err)
				}
			}

			err := w.Cleanup()

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Verify file is removed after cleanup
			if tt.setupTmpfile {
				if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
					t.Error("tmpfile should be removed after cleanup")
				}
			}
		})
	}
}
