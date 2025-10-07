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
