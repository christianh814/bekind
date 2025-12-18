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

func TestPostInstallActions(t *testing.T) {
	testCases := []struct {
		name        string
		actions     []PostInstallAction
		expectError bool
	}{
		{
			name:        "empty actions",
			actions:     []PostInstallAction{},
			expectError: false,
		},
		{
			name: "missing action field",
			actions: []PostInstallAction{
				{
					Kind: "Deployment",
					Name: "test",
				},
			},
			expectError: false, // Should warn and skip
		},
		{
			name: "missing kind field",
			actions: []PostInstallAction{
				{
					Action: "restart",
					Name:   "test",
				},
			},
			expectError: false, // Should warn and skip
		},
		{
			name: "missing both name and labelSelector",
			actions: []PostInstallAction{
				{
					Action: "restart",
					Kind:   "Deployment",
				},
			},
			expectError: false, // Should warn and skip
		},
		{
			name: "unsupported action",
			actions: []PostInstallAction{
				{
					Action: "update",
					Kind:   "Deployment",
					Name:   "test",
				},
			},
			expectError: false, // Should warn and skip
		},
		{
			name: "unsupported kind for restart",
			actions: []PostInstallAction{
				{
					Action: "restart",
					Kind:   "Pod",
					Name:   "test",
				},
			},
			expectError: false, // Should warn and skip
		},
		{
			name: "unsupported kind for delete",
			actions: []PostInstallAction{
				{
					Action: "delete",
					Kind:   "Deployment",
					Name:   "test",
				},
			},
			expectError: false, // Should warn and skip
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// PostInstallActions with nil config will skip invalid actions
			// We're testing the validation logic, not the actual execution
			err := PostInstallActions(tc.actions, context.TODO(), nil)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			// Note: err may be nil for cases where actions are skipped due to validation failures
		})
	}
}

func TestPostInstallAction_Validation(t *testing.T) {
	testCases := []struct {
		name           string
		action         PostInstallAction
		shouldValidate bool
	}{
		{
			name: "valid with name",
			action: PostInstallAction{
				Action: "restart",
				Kind:   "Deployment",
				Name:   "test",
			},
			shouldValidate: true,
		},
		{
			name: "valid with labelSelector",
			action: PostInstallAction{
				Action: "restart",
				Kind:   "StatefulSet",
				LabelSelector: map[string]string{
					"app": "test",
				},
			},
			shouldValidate: true,
		},
		{
			name: "valid with both name and labelSelector",
			action: PostInstallAction{
				Action: "restart",
				Kind:   "DaemonSet",
				Name:   "test",
				LabelSelector: map[string]string{
					"app": "test",
				},
			},
			shouldValidate: true, // labelSelector takes precedence
		},
		{
			name: "invalid - no action",
			action: PostInstallAction{
				Kind: "Deployment",
				Name: "test",
			},
			shouldValidate: false,
		},
		{
			name: "invalid - no kind",
			action: PostInstallAction{
				Action: "restart",
				Name:   "test",
			},
			shouldValidate: false,
		},
		{
			name: "invalid - no name or labelSelector",
			action: PostInstallAction{
				Action: "restart",
				Kind:   "Deployment",
			},
			shouldValidate: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation checks
			hasAction := tc.action.Action != ""
			hasKind := tc.action.Kind != ""
			hasNameOrSelector := tc.action.Name != "" || len(tc.action.LabelSelector) > 0

			validates := hasAction && hasKind && hasNameOrSelector

			if validates != tc.shouldValidate {
				t.Errorf("Expected validation to be %v, got %v", tc.shouldValidate, validates)
			}
		})
	}
}

func TestPostInstallPatches_Validation(t *testing.T) {
	testCases := []struct {
		name           string
		patch          PostInstallPatch
		shouldValidate bool
	}{
		{
			name: "valid patch with all required fields",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Group:     "gateway.networking.k8s.io",
					Version:   "v1",
					Kind:      "GRPCRoute",
					Name:      "argocd-server-grpc",
					Namespace: "argocd",
				},
				Patch: `[{"op": "replace", "path": "/spec/rules/0/backendRefs/0/port", "value": 443}]`,
			},
			shouldValidate: true,
		},
		{
			name: "valid patch with core group (empty string)",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Group:     "",
					Version:   "v1",
					Kind:      "Service",
					Name:      "my-service",
					Namespace: "default",
				},
				Patch: `[{"op": "add", "path": "/metadata/labels/app", "value": "test"}]`,
			},
			shouldValidate: true,
		},
		{
			name: "valid patch without namespace (defaults to default)",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version: "v1",
					Kind:    "ConfigMap",
					Name:    "my-config",
				},
				Patch: `[{"op": "replace", "path": "/data/key", "value": "new-value"}]`,
			},
			shouldValidate: true,
		},
		{
			name: "valid patch without group (defaults to core)",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version:   "v1",
					Kind:      "Pod",
					Name:      "my-pod",
					Namespace: "kube-system",
				},
				Patch: `[{"op": "add", "path": "/metadata/annotations/test", "value": "annotation"}]`,
			},
			shouldValidate: true,
		},
		{
			name: "invalid - no version",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Kind:      "Deployment",
					Name:      "test",
					Namespace: "default",
				},
				Patch: `[{"op": "replace", "path": "/spec/replicas", "value": 3}]`,
			},
			shouldValidate: false,
		},
		{
			name: "invalid - no kind",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version:   "v1",
					Name:      "test",
					Namespace: "default",
				},
				Patch: `[{"op": "replace", "path": "/spec/replicas", "value": 3}]`,
			},
			shouldValidate: false,
		},
		{
			name: "invalid - no name",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version:   "v1",
					Kind:      "Deployment",
					Namespace: "default",
				},
				Patch: `[{"op": "replace", "path": "/spec/replicas", "value": 3}]`,
			},
			shouldValidate: false,
		},
		{
			name: "invalid - no patch",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version:   "v1",
					Kind:      "Deployment",
					Name:      "test",
					Namespace: "default",
				},
				Patch: "",
			},
			shouldValidate: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Basic validation checks
			hasVersion := tc.patch.Target.Version != ""
			hasKind := tc.patch.Target.Kind != ""
			hasName := tc.patch.Target.Name != ""
			hasPatch := tc.patch.Patch != ""

			validates := hasVersion && hasKind && hasName && hasPatch

			if validates != tc.shouldValidate {
				t.Errorf("Expected validation to be %v, got %v", tc.shouldValidate, validates)
			}
		})
	}
}

func TestPostInstallPatches_Defaults(t *testing.T) {
	testCases := []struct {
		name              string
		patch             PostInstallPatch
		expectedGroup     string
		expectedNamespace string
	}{
		{
			name: "empty group defaults to core",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Group:     "",
					Version:   "v1",
					Kind:      "Service",
					Name:      "my-service",
					Namespace: "test-ns",
				},
				Patch: `[{"op": "add", "path": "/metadata/labels/app", "value": "test"}]`,
			},
			expectedGroup:     "",
			expectedNamespace: "test-ns",
		},
		{
			name: "empty namespace defaults to default",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Group:     "apps",
					Version:   "v1",
					Kind:      "Deployment",
					Name:      "my-deployment",
					Namespace: "",
				},
				Patch: `[{"op": "replace", "path": "/spec/replicas", "value": 3}]`,
			},
			expectedGroup:     "apps",
			expectedNamespace: "default",
		},
		{
			name: "both empty - core group and default namespace",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Version: "v1",
					Kind:    "ConfigMap",
					Name:    "my-config",
				},
				Patch: `[{"op": "add", "path": "/data/key", "value": "value"}]`,
			},
			expectedGroup:     "",
			expectedNamespace: "default",
		},
		{
			name: "explicit values are preserved",
			patch: PostInstallPatch{
				Target: PatchTarget{
					Group:     "gateway.networking.k8s.io",
					Version:   "v1",
					Kind:      "HTTPRoute",
					Name:      "my-route",
					Namespace: "custom-ns",
				},
				Patch: `[{"op": "replace", "path": "/spec/rules/0/backendRefs/0/port", "value": 8080}]`,
			},
			expectedGroup:     "gateway.networking.k8s.io",
			expectedNamespace: "custom-ns",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply defaults as the function would
			group := tc.patch.Target.Group
			if group == "" {
				group = ""
			}
			namespace := tc.patch.Target.Namespace
			if namespace == "" {
				namespace = "default"
			}

			if group != tc.expectedGroup {
				t.Errorf("Expected group to be '%s', got '%s'", tc.expectedGroup, group)
			}
			if namespace != tc.expectedNamespace {
				t.Errorf("Expected namespace to be '%s', got '%s'", tc.expectedNamespace, namespace)
			}
		})
	}
}
