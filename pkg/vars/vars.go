package vars

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"gopkg.in/yaml.v2"
)

// Var is a single user-defined variable from the top-level "vars" config section
type Var struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

// reservedNames are top-level bekind config fields that cannot be used as variable names
var reservedNames = map[string]bool{
	"vars":                 true,
	"name":                 true,
	"domain":               true,
	"kindConfig":           true,
	"kindImageVersion":     true,
	"loadDockerImages":     true,
	"preHelmManifests":     true,
	"postInstallManifests": true,
	"postInstallActions":   true,
	"postInstallPatches":   true,
	"helmCharts":           true,
	"helmStack":            true,
	"nodes":                true,
}

// validName matches variable names that are safe to reference in a template
var validName = regexp.MustCompile(`^[a-zA-Z_]\w*$`)

// Parse extracts the top-level "vars" section from YAML data and returns the
// evaluated variables. A variable's value may reference previously defined
// variables (e.g. ${{ .vars.previous }}).
func Parse(data []byte) (map[string]string, error) {
	var doc struct {
		Vars []Var `yaml:"vars"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing vars: %w", err)
	}

	vars := make(map[string]string, len(doc.Vars))
	for _, v := range doc.Vars {
		if !validName.MatchString(v.Name) {
			return nil, fmt.Errorf("invalid variable name %q: must match %s", v.Name, validName.String())
		}
		if reservedNames[v.Name] {
			return nil, fmt.Errorf("variable name %q is reserved: it matches a bekind config field", v.Name)
		}
		if _, ok := vars[v.Name]; ok {
			return nil, fmt.Errorf("duplicate variable name %q", v.Name)
		}
		val, err := render(v.Value, vars)
		if err != nil {
			return nil, fmt.Errorf("evaluating variable %q: %w", v.Name, err)
		}
		vars[v.Name] = val
	}
	return vars, nil
}

// Expand renders all ${{ .vars.<name> }} references in data. Referencing an
// undefined variable is an error.
func Expand(data []byte, vars map[string]string) ([]byte, error) {
	out, err := render(string(data), vars)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}

// render executes s as a Go template with ${{ }} delimiters and the given vars
func render(s string, vars map[string]string) (string, error) {
	tmpl, err := template.New("bekind").Delims("${{", "}}").Option("missingkey=error").Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]interface{}{"vars": vars}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
