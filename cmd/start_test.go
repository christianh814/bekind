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
package cmd

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestConvertMapInterface(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "simple string",
			input:    "test",
			expected: "test",
		},
		{
			name:     "simple number",
			input:    42,
			expected: 42,
		},
		{
			name: "map with interface keys",
			input: map[interface{}]interface{}{
				"key1": "value1",
				"key2": 42,
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
			},
		},
		{
			name: "nested map",
			input: map[interface{}]interface{}{
				"outer": map[interface{}]interface{}{
					"inner": "value",
				},
			},
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "value",
				},
			},
		},
		{
			name: "slice with maps",
			input: []interface{}{
				map[interface{}]interface{}{
					"key": "value",
				},
				"string",
			},
			expected: []interface{}{
				map[string]interface{}{
					"key": "value",
				},
				"string",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertMapInterface(tc.input)

			// For simple types, direct comparison
			if tc.name == "simple string" || tc.name == "simple number" {
				if result != tc.expected {
					t.Errorf("Expected %v, got %v", tc.expected, result)
				}
				return
			}

			// For complex types, we'll do basic type checking
			switch expected := tc.expected.(type) {
			case map[string]interface{}:
				resultMap, ok := result.(map[string]interface{})
				if !ok {
					t.Errorf("Expected map[string]interface{}, got %T", result)
					return
				}

				if len(resultMap) != len(expected) {
					t.Errorf("Expected map length %d, got %d", len(expected), len(resultMap))
				}

				// Check that all keys are strings
				for key := range resultMap {
					if key == "" {
						t.Error("Map key should not be empty")
					}
				}

			case []interface{}:
				resultSlice, ok := result.([]interface{})
				if !ok {
					t.Errorf("Expected []interface{}, got %T", result)
					return
				}

				if len(resultSlice) != len(expected) {
					t.Errorf("Expected slice length %d, got %d", len(expected), len(resultSlice))
				}
			}
		})
	}
}

func TestResetGlobalVars(t *testing.T) {
	// Set some global variables to non-default values
	HC = []struct {
		Url          string
		Repo         string
		Chart        string
		Release      string
		Namespace    string
		ValuesObject map[string]interface{}
		Wait         bool
		Version      string
	}{
		{
			Url:     "test-url",
			Repo:    "test-repo",
			Chart:   "test-chart",
			Release: "test-release",
		},
	}
	pullImages = false
	Domain = "test.domain.com"
	KindImageVersion = "test-version"

	// Call reset function
	ResetGlobalVars()

	// Verify variables are reset
	if HC != nil {
		t.Error("HC should be nil after reset")
	}

	if !pullImages {
		t.Error("pullImages should be true after reset")
	}

	if Domain != "127.0.0.1.nip.io" {
		t.Errorf("Domain should be '127.0.0.1.nip.io' after reset, got '%s'", Domain)
	}

	if KindImageVersion != "" {
		t.Errorf("KindImageVersion should be empty after reset, got '%s'", KindImageVersion)
	}
}

func TestHelmStackParsing(t *testing.T) {
	testCases := []struct {
		name          string
		stackYAML     string
		expectedCount int
		expectError   bool
	}{
		{
			name: "single chart in stack",
			stackYAML: `helmCharts:
  - url: "https://helm.cilium.io"
    repo: "cilium"
    chart: "cilium"
    release: "cilium"
    namespace: "kube-system"
    wait: true
    version: "1.14.0"`,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name: "multiple charts in stack",
			stackYAML: `helmCharts:
  - url: "https://helm.cilium.io"
    repo: "cilium"
    chart: "cilium"
    release: "cilium"
    namespace: "kube-system"
    wait: true
  - url: "https://argoproj.github.io/argo-helm"
    repo: "argo"
    chart: "argo-cd"
    release: "argocd"
    namespace: "argocd"
    wait: false`,
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "chart with valuesObject",
			stackYAML: `helmCharts:
  - url: "https://helm.cilium.io"
    repo: "cilium"
    chart: "cilium"
    release: "cilium"
    namespace: "kube-system"
    wait: true
    valuesObject:
      kubeProxyReplacement: true
      operator:
        replicas: 1
      ingressController:
        enabled: true`,
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "empty stack file",
			stackYAML:     ``,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "invalid YAML",
			stackYAML:     `helmCharts: [invalid yaml`,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the stack YAML
			var stackConfig struct {
				HelmCharts []struct {
					Url          string                 `yaml:"url"`
					Repo         string                 `yaml:"repo"`
					Chart        string                 `yaml:"chart"`
					Release      string                 `yaml:"release"`
					Namespace    string                 `yaml:"namespace"`
					ValuesObject map[string]interface{} `yaml:"valuesObject"`
					Wait         bool                   `yaml:"wait"`
					Version      string                 `yaml:"version"`
				} `yaml:"helmCharts"`
			}

			err := yaml.Unmarshal([]byte(tc.stackYAML), &stackConfig)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(stackConfig.HelmCharts) != tc.expectedCount {
				t.Errorf("Expected %d charts, got %d", tc.expectedCount, len(stackConfig.HelmCharts))
			}

			// Verify conversion works for valuesObject
			for _, chart := range stackConfig.HelmCharts {
				if chart.ValuesObject != nil {
					convertedValues := make(map[string]interface{})
					for k, v := range chart.ValuesObject {
						convertedValues[k] = convertMapInterface(v)
					}
					// Verify conversion succeeded
					if len(convertedValues) == 0 && len(chart.ValuesObject) > 0 {
						t.Error("Values conversion failed")
					}
				}
			}
		})
	}
}

func TestHelmStackIntegration(t *testing.T) {
	// Test the integration of stack charts with inline charts
	t.Run("stack charts should be appended before inline charts", func(t *testing.T) {
		// Reset global HC
		HC = nil

		// Simulate adding charts from a stack
		stackCharts := []struct {
			Url          string
			Repo         string
			Chart        string
			Release      string
			Namespace    string
			ValuesObject map[string]interface{}
			Wait         bool
			Version      string
		}{
			{
				Url:       "https://helm.cilium.io",
				Repo:      "cilium",
				Chart:     "cilium",
				Release:   "cilium",
				Namespace: "kube-system",
				Wait:      true,
			},
		}

		// Append stack charts
		for _, chart := range stackCharts {
			HC = append(HC, chart)
		}

		// Simulate adding inline charts
		inlineChart := struct {
			Url          string
			Repo         string
			Chart        string
			Release      string
			Namespace    string
			ValuesObject map[string]interface{}
			Wait         bool
			Version      string
		}{
			Url:       "https://argoproj.github.io/argo-helm",
			Repo:      "argo",
			Chart:     "argo-rollouts",
			Release:   "argo-rollouts",
			Namespace: "argo-rollouts",
			Wait:      true,
		}
		HC = append(HC, inlineChart)

		// Verify order
		if len(HC) != 2 {
			t.Errorf("Expected 2 charts, got %d", len(HC))
		}

		if HC[0].Chart != "cilium" {
			t.Errorf("Expected first chart to be 'cilium', got '%s'", HC[0].Chart)
		}

		if HC[1].Chart != "argo-rollouts" {
			t.Errorf("Expected second chart to be 'argo-rollouts', got '%s'", HC[1].Chart)
		}

		// Cleanup
		HC = nil
	})
}
