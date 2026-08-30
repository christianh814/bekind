package vars

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		expected  map[string]string
		expectErr bool
	}{
		{
			name:     "no vars section",
			input:    "kindConfig: |\n  kind: Cluster\n",
			expected: map[string]string{},
		},
		{
			name: "simple vars",
			input: `
vars:
  - name: myimage
    value: quay.io/christianh814/simple-go:latest
  - name: domainName
    value: 7f000001.nip.io
`,
			expected: map[string]string{
				"myimage":    "quay.io/christianh814/simple-go:latest",
				"domainName": "7f000001.nip.io",
			},
		},
		{
			name: "var referencing earlier var",
			input: `
vars:
  - name: ip
    value: 7f000001
  - name: domainName
    value: ${{ .vars.ip }}.nip.io
`,
			expected: map[string]string{
				"ip":         "7f000001",
				"domainName": "7f000001.nip.io",
			},
		},
		{
			name: "var referencing later var fails",
			input: `
vars:
  - name: domainName
    value: ${{ .vars.ip }}.nip.io
  - name: ip
    value: 7f000001
`,
			expectErr: true,
		},
		{
			name: "reserved name loadDockerImages",
			input: `
vars:
  - name: loadDockerImages
    value: something
`,
			expectErr: true,
		},
		{
			name: "reserved name kindConfig",
			input: `
vars:
  - name: kindConfig
    value: something
`,
			expectErr: true,
		},
		{
			name: "invalid name with dash",
			input: `
vars:
  - name: my-image
    value: something
`,
			expectErr: true,
		},
		{
			name: "invalid name starting with digit",
			input: `
vars:
  - name: 1image
    value: something
`,
			expectErr: true,
		},
		{
			name: "empty name",
			input: `
vars:
  - name: ""
    value: something
`,
			expectErr: true,
		},
		{
			name: "duplicate name",
			input: `
vars:
  - name: myimage
    value: one
  - name: myimage
    value: two
`,
			expectErr: true,
		},
		{
			name:      "invalid yaml",
			input:     "vars: [::",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse([]byte(tc.input))
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got none (result: %v)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestExpand(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		vars      map[string]string
		expected  string
		expectErr bool
	}{
		{
			name:     "no references passes through unchanged",
			input:    "loadDockerImages:\n  images:\n    - nginx:latest\n",
			vars:     map[string]string{"myimage": "unused"},
			expected: "loadDockerImages:\n  images:\n    - nginx:latest\n",
		},
		{
			name:  "expand in list item",
			input: "loadDockerImages:\n  images:\n    - ${{ .vars.myimage }}\n",
			vars:  map[string]string{"myimage": "quay.io/christianh814/simple-go:latest"},
			expected: "loadDockerImages:\n  images:\n" +
				"    - quay.io/christianh814/simple-go:latest\n",
		},
		{
			name: "expand in nested helm valuesObject",
			input: `helmCharts:
  - chart: test
    valuesObject:
      controller:
        domains:
          - ${{ .vars.domainName }}
`,
			vars: map[string]string{"domainName": "7f000001.nip.io"},
			expected: `helmCharts:
  - chart: test
    valuesObject:
      controller:
        domains:
          - 7f000001.nip.io
`,
		},
		{
			name:     "multiple references in one line",
			input:    "url: https://${{ .vars.host }}:${{ .vars.port }}\n",
			vars:     map[string]string{"host": "example.com", "port": "8443"},
			expected: "url: https://example.com:8443\n",
		},
		{
			name:      "undefined variable errors",
			input:     "image: ${{ .vars.nope }}\n",
			vars:      map[string]string{},
			expectErr: true,
		},
		{
			name:      "unclosed delimiter errors",
			input:     "image: ${{ .vars.myimage\n",
			vars:      map[string]string{"myimage": "nginx"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Expand([]byte(tc.input), tc.vars)
			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error, got none (result: %s)", result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(result) != tc.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tc.expected, result)
			}
		})
	}
}
