// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package fleetpkg

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var update = flag.Bool("update", false, "Update golden files")

func TestRead(t *testing.T) {
	packages, err := filepath.Glob("testdata/packages/*")
	if err != nil {
		t.Fatal(err)
	}

	for _, pkgPath := range packages {
		t.Run(filepath.Base(pkgPath), func(t *testing.T) {
			pkg, err := Read(pkgPath)
			if err != nil {
				t.Fatal(err)
			}

			jsonGolden := filepath.Join("testdata", filepath.Base(pkgPath)+".golden.json")
			yamlGolden := filepath.Join("testdata", filepath.Base(pkgPath)+".golden.yml")

			if *update {
				writeJSON(t, jsonGolden, pkg)
				writeYAML(t, yamlGolden, pkg)
				return
			}

			compareJSON(t, jsonGolden, pkg)
			compareYAML(t, yamlGolden, pkg)
		})
	}
}

func TestAllIntegrations(t *testing.T) {
	integrationsDir := os.Getenv("INTEGRATIONS_DIR")
	if integrationsDir == "" {
		t.Skip("INTEGRATIONS_DIR env var is not set.")
	}

	packages, err := filepath.Glob(filepath.Join(integrationsDir, "packages/*"))
	if err != nil {
		t.Fatal(err)
	}

	if len(packages) == 0 {
		t.Error("No packages were found in INTEGRATIONS_DIR")
	}

	// Process packages in parallel using subtests
	for _, pkgPath := range packages {
		pkgPath := pkgPath // capture loop variable
		t.Run(filepath.Base(pkgPath), func(t *testing.T) {
			t.Parallel() // Enable parallel execution
			_, err := Read(pkgPath)
			if err != nil {
				t.Errorf("failed reading from %v: %v", pkgPath, err)
			}
		})
	}
}

func TestConditions(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
	}{
		{
			name: "flat",
			yaml: `
kibana.version: "^8.6.0"
elastic.subscription: basic
`,
		},
		{
			name: "nested",
			yaml: `
kibana:
  version: "^8.6.0"
elastic:
  subscription: basic
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var c Conditions
			if err := yaml.Unmarshal([]byte(tc.yaml), &c); err != nil {
				t.Fatal(err)
			}

			if c.Elastic.Subscription != "basic" {
				t.Errorf("expected basic, got %q", c.Elastic.Subscription)
			}
			if c.Kibana.Version != "^8.6.0" {
				t.Errorf("expected ^8.6.0, got %q", c.Elastic.Subscription)
			}
		})
	}
}

func writeYAML(t *testing.T, file string, data any) {
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)

	if err := enc.Encode(data); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, file string, data any) {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	if err := enc.Encode(data); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func compareJSON(t *testing.T, file string, data any) {
	// Read file into any.
	goldenData, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	// Write data to JSON.
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(data); err != nil {
		t.Fatal(err)
	}

	// Compare JSON
	assert.JSONEq(t, string(goldenData), buf.String())
}

func compareYAML(t *testing.T, file string, data any) {
	// Read file into any.
	goldenData, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	// Write data to YAML.
	buf := new(bytes.Buffer)
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)

	if err := enc.Encode(data); err != nil {
		t.Fatal(err)
	}

	// Compare YAML
	assert.YAMLEq(t, string(goldenData), buf.String())
}

func TestProcessor_UnmarshalYAML(t *testing.T) {
	testCases := []struct {
		name string
		yaml string
		want *Processor
	}{
		{
			name: "no_on_failure",
			yaml: `
set:
  field: key
  value: value
`,
			want: &Processor{
				Type: "set",
				Attributes: map[string]any{
					"field": "key",
					"value": "value",
				},
				FileMetadata: FileMetadata{
					line:   2,
					column: 1,
				},
			},
		},
		{
			name: "on_failure",
			yaml: `
set:
  field: key
  value: value
  on_failure:
    - set:
        field: on_fail_key
        value: on_fail_value
`,
			want: &Processor{
				Type: "set",
				Attributes: map[string]any{
					"field": "key",
					"value": "value",
				},
				OnFailure: []*Processor{
					{
						Type: "set",
						Attributes: map[string]any{
							"field": "on_fail_key",
							"value": "on_fail_value",
						},
						FileMetadata: FileMetadata{
							line:   6,
							column: 7,
						},
					},
				},
				FileMetadata: FileMetadata{
					line:   2,
					column: 1,
				},
			},
		},
		{
			name: "on_failure_nested",
			yaml: `
set:
  field: key
  value: value
  on_failure:
    - set:
        field: on_fail_key
        value: on_fail_value
        on_failure:
        - set:
            field: on_fail2_key
            value: on_fail2_value
`,
			want: &Processor{
				Type: "set",
				Attributes: map[string]any{
					"field": "key",
					"value": "value",
				},
				OnFailure: []*Processor{
					{
						Type: "set",
						Attributes: map[string]any{
							"field": "on_fail_key",
							"value": "on_fail_value",
						},
						OnFailure: []*Processor{
							{
								Type: "set",
								Attributes: map[string]any{
									"field": "on_fail2_key",
									"value": "on_fail2_value",
								},
								FileMetadata: FileMetadata{
									line:   10,
									column: 11,
								},
							},
						},
						FileMetadata: FileMetadata{
							line:   6,
							column: 7,
						},
					},
				},
				FileMetadata: FileMetadata{
					line:   2,
					column: 1,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p Processor

			err := yaml.Unmarshal([]byte(tc.yaml), &p)
			require.NoError(t, err)
			require.Equal(t, tc.want, &p)
		})
	}
}
