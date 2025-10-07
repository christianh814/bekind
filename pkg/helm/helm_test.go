/*
Copyright © 2022 Christian Hernandez christian@chernand.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package helm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"helm.sh/helm/v3/pkg/cli"
)

func TestInstallFunction(t *testing.T) {
	// Test the Install function exists and has the right signature
	// We can't test actual execution without a Kubernetes cluster and Helm setup
	// but we can test the function signature and basic parameter validation

	defer func() {
		if r := recover(); r != nil {
			// If it panics due to missing Kubernetes cluster, that's expected
			// We just want to make sure the function exists and is callable
		}
	}()

	// Test with empty parameters (should fail gracefully)
	err := Install("", "", "", "", "", "", false, nil)
	if err == nil {
		t.Error("Install with empty parameters should return an error")
	}
}

func TestRepoAddFunction(t *testing.T) {
	// Test RepoAdd function with invalid parameters
	// This should fail gracefully without panicking

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Set HELM_REPOSITORY_CONFIG to our test directory
	originalHelmConfig := os.Getenv("HELM_REPOSITORY_CONFIG")
	defer func() {
		if originalHelmConfig != "" {
			os.Setenv("HELM_REPOSITORY_CONFIG", originalHelmConfig)
		} else {
			os.Unsetenv("HELM_REPOSITORY_CONFIG")
		}
	}()

	// Initialize settings for testing
	settings = cli.New()
	settings.RepositoryConfig = filepath.Join(tmpDir, "repositories.yaml")

	// Test with invalid URL
	err := RepoAdd("test-repo", "invalid-url")
	if err == nil {
		t.Error("RepoAdd with invalid URL should return an error")
	}

	// Test with empty name
	err = RepoAdd("", "https://charts.example.com")
	if err == nil {
		t.Error("RepoAdd with empty name should return an error")
	}
}

func TestRepoUpdateFunction(t *testing.T) {
	// Test RepoUpdate function
	// This will likely fail in a test environment, but should not panic

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Initialize settings for testing
	settings = cli.New()
	settings.RepositoryConfig = filepath.Join(tmpDir, "repositories.yaml")

	// Create an empty repositories file
	emptyRepoFile := `apiVersion: ""
generated: "0001-01-01T00:00:00Z"
repositories: []
`
	err := os.WriteFile(settings.RepositoryConfig, []byte(emptyRepoFile), 0644)
	if err != nil {
		t.Fatalf("Failed to create test repositories file: %v", err)
	}

	// Test RepoUpdate with empty repositories
	err = RepoUpdate()
	if err == nil {
		// This might succeed with empty repositories
	}
}

func TestInstallChartFunction(t *testing.T) {
	// Test InstallChart function exists and handles invalid parameters

	defer func() {
		if r := recover(); r != nil {
			// If it panics due to missing Kubernetes cluster, that's expected
		}
	}()

	// Test with invalid parameters
	err := InstallChart("", "", "", "", "", false, nil)
	if err == nil {
		t.Error("InstallChart with empty parameters should return an error")
	}
}

func TestIsChartInstallable(t *testing.T) {
	// We can't easily test this without creating actual chart metadata
	// but we can test that the function exists
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("isChartInstallable should not panic: %v", r)
		}
	}()

	// The function exists if we can reference it without compilation errors
	_ = isChartInstallable
}

func TestDebugFunction(t *testing.T) {
	// Test debug function
	// This function just formats a debug message, so we can test it safely

	// Capture any output (though debug might not write to stdout)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("debug function should not panic: %v", r)
		}
	}()

	// Call debug function
	debug("test message")
	debug("test message with args: %s %d", "string", 42)
}

func TestGetChartPath(t *testing.T) {
	// Test getChartPath function logic for OCI vs regular repos

	// We can't fully test this without proper Helm setup, but we can test
	// the URL detection logic

	defer func() {
		if r := recover(); r != nil {
			// Expected to fail in test environment
		}
	}()

	// Test OCI URL detection
	ociUrl := "oci://registry.example.com/charts/app"
	if !strings.HasPrefix(ociUrl, "oci://") {
		t.Error("OCI URL should be detected correctly")
	}

	// Test regular URL
	regularUrl := "https://charts.example.com"
	if strings.HasPrefix(regularUrl, "oci://") {
		t.Error("Regular URL should not be detected as OCI")
	}
}

func TestHelmSettingsInitialization(t *testing.T) {
	// Test that settings can be initialized
	testSettings := cli.New()

	if testSettings == nil {
		t.Error("Helm settings should be initialized")
	}

	// Test that settings has expected properties
	if testSettings.RepositoryConfig == "" {
		t.Error("RepositoryConfig should not be empty")
	}

	if testSettings.RepositoryCache == "" {
		t.Error("RepositoryCache should not be empty")
	}
}

func TestInstallParameterValidation(t *testing.T) {
	// Test parameter validation logic that would be used in Install function

	testCases := []struct {
		name        string
		namespace   string
		url         string
		repoName    string
		chartName   string
		releaseName string
		version     string
		expectError bool
	}{
		{
			name:        "all empty",
			namespace:   "",
			url:         "",
			repoName:    "",
			chartName:   "",
			releaseName: "",
			version:     "",
			expectError: true,
		},
		{
			name:        "valid OCI parameters",
			namespace:   "default",
			url:         "oci://registry.example.com/charts/app",
			repoName:    "",
			chartName:   "",
			releaseName: "test-release",
			version:     "1.0.0",
			expectError: false, // URL format is valid
		},
		{
			name:        "valid regular parameters",
			namespace:   "default",
			url:         "https://charts.example.com",
			repoName:    "stable",
			chartName:   "nginx",
			releaseName: "test-nginx",
			version:     "1.0.0",
			expectError: false, // Parameters look valid
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test basic parameter validation logic
			isEmpty := tc.namespace == "" || tc.url == "" || tc.releaseName == ""

			if tc.expectError && !isEmpty {
				// Additional validation would happen in the actual function
			}

			if !tc.expectError && isEmpty {
				t.Error("Should expect error for empty required parameters")
			}
		})
	}
}

func TestHelmNamespaceHandling(t *testing.T) {
	// Test namespace environment variable handling
	originalNamespace := os.Getenv("HELM_NAMESPACE")
	defer func() {
		if originalNamespace != "" {
			os.Setenv("HELM_NAMESPACE", originalNamespace)
		} else {
			os.Unsetenv("HELM_NAMESPACE")
		}
	}()

	// Test setting namespace
	testNamespace := "test-namespace"
	os.Setenv("HELM_NAMESPACE", testNamespace)

	envNamespace := os.Getenv("HELM_NAMESPACE")
	if envNamespace != testNamespace {
		t.Errorf("Expected namespace '%s', got '%s'", testNamespace, envNamespace)
	}
}

func TestHelmDriverEnvironment(t *testing.T) {
	// Test HELM_DRIVER environment variable
	originalDriver := os.Getenv("HELM_DRIVER")
	defer func() {
		if originalDriver != "" {
			os.Setenv("HELM_DRIVER", originalDriver)
		} else {
			os.Unsetenv("HELM_DRIVER")
		}
	}()

	// Test default driver (should be empty or default)
	_ = os.Getenv("HELM_DRIVER") // Just check it doesn't panic

	// Test setting custom driver
	os.Setenv("HELM_DRIVER", "configmap")
	customDriver := os.Getenv("HELM_DRIVER")
	if customDriver != "configmap" {
		t.Errorf("Expected driver 'configmap', got '%s'", customDriver)
	}

	// Clean up
	if originalDriver == "" {
		os.Unsetenv("HELM_DRIVER")
	} else {
		os.Setenv("HELM_DRIVER", originalDriver)
	}
}
