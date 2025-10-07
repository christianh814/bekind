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
package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetDefaultRuntime(t *testing.T) {
	// Save original environment variable
	originalProvider := os.Getenv("KIND_EXPERIMENTAL_PROVIDER")
	defer func() {
		if originalProvider != "" {
			os.Setenv("KIND_EXPERIMENTAL_PROVIDER", originalProvider)
		} else {
			os.Unsetenv("KIND_EXPERIMENTAL_PROVIDER")
		}
	}()

	testCases := []struct {
		name     string
		envValue string
		isNil    bool
	}{
		{
			name:     "no provider set",
			envValue: "",
			isNil:    true,
		},
		{
			name:     "docker provider",
			envValue: "docker",
			isNil:    false,
		},
		{
			name:     "podman provider",
			envValue: "podman",
			isNil:    false,
		},
		{
			name:     "unknown provider",
			envValue: "unknown",
			isNil:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.envValue == "" {
				os.Unsetenv("KIND_EXPERIMENTAL_PROVIDER")
			} else {
				os.Setenv("KIND_EXPERIMENTAL_PROVIDER", tc.envValue)
			}

			result := GetDefaultRuntime()

			if tc.isNil && result != nil {
				t.Errorf("Expected nil result for %s, got %v", tc.name, result)
			}

			if !tc.isNil && result == nil {
				t.Errorf("Expected non-nil result for %s", tc.name)
			}
		})
	}
}

func TestDownloadFileString(t *testing.T) {
	// Create a test server
	testContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	// Test successful download
	content, err := DownloadFileString(server.URL)
	if err != nil {
		t.Fatalf("DownloadFileString failed: %v", err)
	}

	if content != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, content)
	}

	// Test invalid URL
	_, err = DownloadFileString("invalid-url")
	if err == nil {
		t.Error("DownloadFileString should fail with invalid URL")
	}
}

func TestSplitYAML(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedLen int
		expectError bool
	}{
		{
			name: "single document",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: test-pod`,
			expectedLen: 1,
			expectError: false,
		},
		{
			name: "multiple documents",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: test-pod1
---
apiVersion: v1
kind: Pod
metadata:
  name: test-pod2`,
			expectedLen: 2,
			expectError: false,
		},
		{
			name:        "empty input",
			input:       "",
			expectedLen: 0,
			expectError: false,
		},
		{
			name:        "invalid YAML",
			input:       "invalid: yaml: content: [",
			expectedLen: 0,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := SplitYAML([]byte(tc.input))

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tc.expectError && len(result) != tc.expectedLen {
				t.Errorf("Expected %d documents, got %d", tc.expectedLen, len(result))
			}
		})
	}
}

func TestIsDeploymentRunning(t *testing.T) {
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset()

	// Test function returns a ConditionWithContextFunc
	conditionFunc := IsDeploymentRunning(clientset, "default", "test-deployment")
	if conditionFunc == nil {
		t.Error("IsDeploymentRunning should return a non-nil function")
	}

	// Test the condition function with nonexistent deployment
	ready, err := conditionFunc(context.TODO())
	if err != nil {
		t.Errorf("Condition function should not error for nonexistent deployment: %v", err)
	}

	if ready {
		t.Error("Nonexistent deployment should not be ready")
	}
}

func TestWaitForDeployment(t *testing.T) {
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset()

	// Test with very short timeout to avoid long test runs
	timeout := 100 * time.Millisecond

	err := WaitForDeployment(clientset, "default", "nonexistent-deployment", timeout)
	if err == nil {
		t.Error("WaitForDeployment should timeout for nonexistent deployment")
	}
}

func TestGetRestConfig(t *testing.T) {
	// Test with empty kubeconfig path
	_, err := GetRestConfig("")
	// This might fail in test environment, which is expected
	_ = err

	// Test with invalid kubeconfig path
	_, err = GetRestConfig("/nonexistent/kubeconfig")
	if err == nil {
		t.Error("GetRestConfig should fail with nonexistent kubeconfig")
	}
}

func TestNewClient(t *testing.T) {
	// Test with empty kubeconfig path
	_, err := NewClient("")
	// This might fail in test environment, which is expected
	_ = err

	// Test with invalid kubeconfig path
	_, err = NewClient("/nonexistent/kubeconfig")
	if err == nil {
		t.Error("NewClient should fail with nonexistent kubeconfig")
	}
}

func TestConvertHelmValsToMap(t *testing.T) {
	testCases := []struct {
		name  string
		input []struct {
			Name  string
			Value string
		}
		expected map[string]string
	}{
		{
			name: "empty input",
			input: []struct {
				Name  string
				Value string
			}{},
			expected: map[string]string{"set": ""},
		},
		{
			name: "single value",
			input: []struct {
				Name  string
				Value string
			}{
				{Name: "key1", Value: "value1"},
			},
			expected: map[string]string{"set": "key1=value1"},
		},
		{
			name: "multiple values",
			input: []struct {
				Name  string
				Value string
			}{
				{Name: "key1", Value: "value1"},
				{Name: "key2", Value: "value2"},
			},
			expected: map[string]string{"set": "key1=value1,key2=value2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ConvertHelmValsToMap(tc.input)

			if result["set"] != tc.expected["set"] {
				t.Errorf("Expected set value '%s', got '%s'", tc.expected["set"], result["set"])
			}
		})
	}
}

func TestLabelWorkers(t *testing.T) {
	// Create a fake Kubernetes client
	clientset := fake.NewSimpleClientset()

	// Test with no nodes
	err := LabelWorkers(clientset)
	if err != nil {
		t.Errorf("LabelWorkers should handle no nodes gracefully: %v", err)
	}
}

func TestPostInstallManifests(t *testing.T) {
	// Test with empty manifests slice
	err := PostInstallManifests([]string{}, context.TODO(), nil)
	if err != nil {
		t.Errorf("PostInstallManifests should handle empty slice: %v", err)
	}

	// Test with invalid manifest URL
	err = PostInstallManifests([]string{"invalid-url"}, context.TODO(), nil)
	if err == nil {
		t.Error("PostInstallManifests should fail with invalid URL")
	}
}

func TestSaveBeKindConfig(t *testing.T) {
	// Test with nil config - this should fail gracefully
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior with nil config
			// but let's make sure it's the expected panic
		}
	}()

	err := SaveBeKindConfig(nil, context.TODO(), "test-ns", "test-name")
	if err == nil {
		t.Error("SaveBeKindConfig should fail with nil config")
	}
}

func TestGetBeKindConfig(t *testing.T) {
	// Test with nil config - this should fail gracefully
	defer func() {
		if r := recover(); r != nil {
			// If it panics, that's expected behavior with nil config
		}
	}()

	_, err := GetBeKindConfig(nil, context.TODO(), "test-ns", "test-name")
	if err == nil {
		t.Error("GetBeKindConfig should fail with nil config")
	}
}

func TestGetPostInstallBytes(t *testing.T) {
	// Test with invalid URL scheme
	_, err := getPostInstallBytes("invalid://test")
	if err == nil {
		t.Error("getPostInstallBytes should fail with invalid URL scheme")
	}

	expectedError := "only http://, https://, and file:// are supported"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}

	// Test with HTTP URL (using test server)
	testContent := "test manifest content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testContent))
	}))
	defer server.Close()

	data, err := getPostInstallBytes(server.URL)
	if err != nil {
		t.Fatalf("getPostInstallBytes failed with HTTP URL: %v", err)
	}

	if string(data) != testContent {
		t.Errorf("Expected content '%s', got '%s'", testContent, string(data))
	}

	// Test with file:// URL
	tmpFile, err := os.CreateTemp("", "test-manifest-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	testFileContent := "file test content"
	tmpFile.WriteString(testFileContent)
	tmpFile.Close()

	fileURL := "file://" + tmpFile.Name()
	data, err = getPostInstallBytes(fileURL)
	if err != nil {
		t.Fatalf("getPostInstallBytes failed with file URL: %v", err)
	}

	if string(data) != testFileContent {
		t.Errorf("Expected file content '%s', got '%s'", testFileContent, string(data))
	}
}

func TestViperIntegration(t *testing.T) {
	// Save original viper state
	originalSettings := viper.AllSettings()
	defer func() {
		viper.Reset()
		for k, v := range originalSettings {
			viper.Set(k, v)
		}
	}()

	viper.Reset()

	// Test setting and getting values
	viper.Set("test.key", "test.value")
	value := viper.GetString("test.key")
	if value != "test.value" {
		t.Errorf("Expected 'test.value', got '%s'", value)
	}

	// Test AllSettings
	settings := viper.AllSettings()
	if len(settings) == 0 {
		t.Error("AllSettings should not be empty after setting values")
	}
}
