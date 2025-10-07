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
	"testing"
)

func TestPurgeFunction(t *testing.T) {
	// Test that the purge function exists and can be called
	// We can't test actual execution without KIND clusters
	// but we can verify the function exists

	defer func() {
		if r := recover(); r != nil {
			// If it panics due to no KIND being available, that's expected
			// We just want to make sure it doesn't panic due to the function not existing
		}
	}()

	// The purge function exists if we can reference it without compilation errors
	_ = purge
}
