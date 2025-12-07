package main

import (
	"testing"

	ver "github.com/hashicorp/go-version"
)

// helper to create version collection from strings
func mustVersions(t *testing.T, versions ...string) ver.Collection {
	t.Helper()
	var result ver.Collection
	for _, v := range versions {
		version, err := ver.NewVersion(v)
		if err != nil {
			t.Fatalf("invalid test version %q: %v", v, err)
		}
		result = append(result, version)
	}
	return result
}

// helper to create constraint from string
func mustConstraint(t *testing.T, constraint string) ver.Constraints {
	t.Helper()
	c, err := ver.NewConstraint(constraint)
	if err != nil {
		t.Fatalf("invalid test constraint %q: %v", constraint, err)
	}
	return c
}

func TestFindLatest(t *testing.T) {
	tests := []struct {
		name           string
		versions       []string
		constraint     string
		expectedResult string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "finds exact version",
			versions:       []string{"1.0.0", "1.1.0", "1.2.0"},
			constraint:     "= 1.1.0",
			expectedResult: "1.1.0",
			expectError:    false,
		},
		{
			name:           "finds latest matching constraint",
			versions:       []string{"1.0.0", "1.1.0", "1.2.0", "2.0.0"},
			constraint:     "~> 1.0",
			expectedResult: "1.2.0",
			expectError:    false,
		},
		{
			name:           "pessimistic constraint with patch version",
			versions:       []string{"1.0.0", "1.0.5", "1.1.0", "1.2.0"},
			constraint:     "~> 1.0.0",
			expectedResult: "1.0.5",
			expectError:    false,
		},
		{
			name:           "greater than constraint",
			versions:       []string{"0.9.0", "1.0.0", "1.5.0", "2.0.0"},
			constraint:     "> 1.0.0",
			expectedResult: "2.0.0",
			expectError:    false,
		},
		{
			name:           "greater than or equal constraint",
			versions:       []string{"0.9.0", "1.0.0", "1.5.0"},
			constraint:     ">= 1.0.0",
			expectedResult: "1.5.0",
			expectError:    false,
		},
		{
			name:           "less than constraint",
			versions:       []string{"1.0.0", "1.5.0", "2.0.0"},
			constraint:     "< 2.0.0",
			expectedResult: "1.5.0",
			expectError:    false,
		},
		{
			name:           "combined constraints",
			versions:       []string{"0.9.0", "1.0.0", "1.5.0", "2.0.0", "2.5.0"},
			constraint:     ">= 1.0.0, < 2.0.0",
			expectedResult: "1.5.0",
			expectError:    false,
		},
		{
			name:           "no matching version returns error",
			versions:       []string{"1.0.0", "1.1.0"},
			constraint:     ">= 2.0.0",
			expectError:    true,
			errorContains:  "no matching version",
		},
		{
			name:          "empty versions returns error",
			versions:      []string{},
			constraint:    ">= 1.0.0",
			expectError:   true,
			errorContains: "no binaries available",
		},
		{
			name:           "prerelease versions are considered",
			versions:       []string{"1.0.0", "1.1.0-alpha", "1.1.0-beta", "1.1.0"},
			constraint:     "~> 1.0",
			expectedResult: "1.1.0",
			expectError:    false,
		},
		{
			name:           "prerelease not matched by release constraint",
			versions:       []string{"1.0.0", "2.0.0-alpha"},
			constraint:     ">= 1.0.0",
			expectedResult: "1.0.0",
			expectError:    false,
		},
		{
			name:           "prerelease matched when explicitly targeted",
			versions:       []string{"1.0.0", "2.0.0-alpha", "2.0.0-beta"},
			constraint:     ">= 2.0.0-alpha",
			expectedResult: "2.0.0-beta",
			expectError:    false,
		},
		{
			name:           "versions not in order still finds latest",
			versions:       []string{"1.5.0", "1.0.0", "1.3.0", "1.2.0"},
			constraint:     "~> 1.0",
			expectedResult: "1.5.0",
			expectError:    false,
		},
		{
			name:           "single version matching",
			versions:       []string{"1.0.0"},
			constraint:     ">= 1.0.0",
			expectedResult: "1.0.0",
			expectError:    false,
		},
		{
			name:           "constraint with != exclusion",
			versions:       []string{"1.0.0", "1.1.0", "1.2.0"},
			constraint:     ">= 1.0.0, != 1.2.0",
			expectedResult: "1.1.0",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &Terraform{
				location: "/tmp/test",
				versions: mustVersions(t, tt.versions...),
			}

			constraint := mustConstraint(t, tt.constraint)
			result, err := tf.FindLatest(constraint)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errorContains)
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.String() != tt.expectedResult {
				t.Errorf("FindLatest() = %v, want %v", result.String(), tt.expectedResult)
			}
		})
	}
}

func TestTerraformString(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		expected string
	}{
		{
			name:     "empty versions",
			versions: []string{},
			expected: "",
		},
		{
			name:     "single version",
			versions: []string{"1.0.0"},
			expected: "1.0.0",
		},
		{
			name:     "multiple versions",
			versions: []string{"1.0.0", "1.1.0", "1.2.0"},
			expected: "1.0.0\n1.1.0\n1.2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := &Terraform{
				versions: mustVersions(t, tt.versions...),
			}
			result := tf.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestListInstalled(t *testing.T) {
	versions := []string{"1.0.0", "1.1.0", "1.2.0"}
	tf := &Terraform{
		versions: mustVersions(t, versions...),
	}

	result := tf.ListInstalled()

	if len(result) != len(versions) {
		t.Errorf("ListInstalled() returned %d versions, want %d", len(result), len(versions))
	}

	for i, v := range result {
		if v.String() != versions[i] {
			t.Errorf("ListInstalled()[%d] = %v, want %v", i, v.String(), versions[i])
		}
	}
}

// helper function to check substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
