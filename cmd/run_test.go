/*
Copyright © 2024 Christian Hernandez <christian@chernand.io>

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
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestProfileLongHelp(t *testing.T) {
	helpText := profileLongHelp()

	if len(helpText) == 0 {
		t.Error("Profile long help should not be empty")
	}

	// Check that help contains key information
	expectedPhrases := []string{
		"You can use \"run\" to run the specified profile",
		"~/.bekind/profiles/{{name}}",
		"config.yaml",
		"--view flag",
	}

	for _, phrase := range expectedPhrases {
		if !contains(helpText, phrase) {
			t.Errorf("Help text should contain '%s'", phrase)
		}
	}
}

func TestGetProfileNames(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test profile directories
	testProfiles := []string{"profile1", "profile2", "profile3"}
	for _, profile := range testProfiles {
		err := os.MkdirAll(filepath.Join(tmpDir, profile), 0755)
		if err != nil {
			t.Fatalf("Failed to create test profile directory: %v", err)
		}
	}

	// Save original ProfileDir and restore after test
	originalProfileDir := ProfileDir
	defer func() {
		ProfileDir = originalProfileDir
	}()

	// Set ProfileDir to our test directory
	ProfileDir = tmpDir

	// Test getProfileNames
	profiles, err := getProfileNames()
	if err != nil {
		t.Fatalf("getProfileNames() returned error: %v", err)
	}

	if len(profiles) != len(testProfiles) {
		t.Errorf("Expected %d profiles, got %d", len(testProfiles), len(profiles))
	}

	// Check that all expected profiles are present
	for _, expectedProfile := range testProfiles {
		found := false
		for _, actualProfile := range profiles {
			if actualProfile == expectedProfile {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected profile '%s' not found in results", expectedProfile)
		}
	}
}

func TestGetProfileNamesNonexistentDir(t *testing.T) {
	// Save original ProfileDir and restore after test
	originalProfileDir := ProfileDir
	defer func() {
		ProfileDir = originalProfileDir
	}()

	// Set ProfileDir to a nonexistent directory
	ProfileDir = "/nonexistent/path"

	// Test getProfileNames with nonexistent directory
	profiles, err := getProfileNames()
	if err == nil {
		t.Error("getProfileNames() should return error for nonexistent directory")
	}

	if len(profiles) != 0 {
		t.Errorf("Expected empty profiles slice for nonexistent directory, got %d profiles", len(profiles))
	}
}

func TestProfileValidArgs(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test profile directories
	testProfiles := []string{"test-profile1", "test-profile2"}
	for _, profile := range testProfiles {
		err := os.MkdirAll(filepath.Join(tmpDir, profile), 0755)
		if err != nil {
			t.Fatalf("Failed to create test profile directory: %v", err)
		}
	}

	// Save original ProfileDir and restore after test
	originalProfileDir := ProfileDir
	defer func() {
		ProfileDir = originalProfileDir
	}()

	// Set ProfileDir to our test directory
	ProfileDir = tmpDir

	// Test profileValidArgs with no existing args (should return profiles)
	completions, directive := profileValidArgs(runCmd, []string{}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}

	if len(completions) != len(testProfiles) {
		t.Errorf("Expected %d completions, got %d", len(testProfiles), len(completions))
	}

	// Test profileValidArgs with existing args (should return no completions)
	completions, directive = profileValidArgs(runCmd, []string{"profile1"}, "")

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("Expected ShellCompDirectiveNoFileComp, got %v", directive)
	}

	if len(completions) != 0 {
		t.Errorf("Expected no completions when args already provided, got %d", len(completions))
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
