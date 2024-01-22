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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Field struct {
	Name                  string   `json:"name,omitempty" yaml:"name,omitempty"`
	Type                  string   `json:"type,omitempty" yaml:"type,omitempty"`
	Description           string   `json:"description,omitempty" yaml:"description,omitempty"`
	Value                 string   `json:"value,omitempty" yaml:"value,omitempty"`
	Example               any      `json:"example,omitempty" yaml:"example,omitempty"`
	MetricType            string   `json:"metric_type,omitempty" yaml:"metric_type,omitempty"`
	Unit                  string   `json:"unit,omitempty" yaml:"unit,omitempty"`
	DateFormat            string   `json:"date_format,omitempty" yaml:"date_format,omitempty"`
	Dimension             *bool    `json:"dimension,omitempty" yaml:"dimension,omitempty"`
	Pattern               string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	External              string   `json:"external,omitempty" yaml:"external,omitempty"`
	Fields                []Field  `json:"fields,omitempty" yaml:"fields,omitempty"`
	DocValues             *bool    `json:"doc_values,omitempty" yaml:"doc_values,omitempty"`
	Index                 *bool    `json:"index,omitempty" yaml:"index,omitempty"`
	CopyTo                string   `json:"copy_to,omitempty" yaml:"copy_to,omitempty"`
	Enabled               *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Dynamic               string   `json:"dynamic,omitempty" yaml:"dynamic,omitempty"`
	ScalingFactor         *int     `json:"scaling_factor,omitempty" yaml:"scaling_factor,omitempty"`
	Analyzer              string   `json:"analyzer,omitempty" yaml:"analyzer,omitempty"`
	SearchAnalyzer        string   `json:"search_analyzer,omitempty" yaml:"search_analyzer,omitempty"`
	MultiFields           []Field  `json:"multi_fields,omitempty" yaml:"multi_fields,omitempty"`
	NullValue             string   `json:"null_value,omitempty" yaml:"null_value,omitempty"`
	IgnoreMalformed       *bool    `json:"ignore_malformed,omitempty" yaml:"ignore_malformed,omitempty"`
	IgnoreAbove           int      `json:"ignore_above,omitempty" yaml:"ignore_above,omitempty"`
	ObjectType            string   `json:"object_type,omitempty" yaml:"object_type,omitempty"`
	ObjectTypeMappingType string   `json:"object_type_mapping_type,omitempty" yaml:"object_type_mapping_type,omitempty"`
	AliasTargetPath       string   `json:"path,omitempty" yaml:"path,omitempty"`
	Normalize             []string `json:"normalize,omitempty" yaml:"normalize,omitempty"`
	Normalizer            string   `json:"normalizer,omitempty" yaml:"normalizer,omitempty"`
	IncludeInParent       *bool    `json:"include_in_parent,omitempty" yaml:"include_in_parent,omitempty"`
	DefaultMetric         string   `json:"default_metric,omitempty" yaml:"default_metric,omitempty"`

	// AdditionalProperties contains additional properties that are not
	// explicitly specified in the package-spec and are not used by Fleet.
	AdditionalAttributes map[string]any `json:"_additional_attributes,omitempty" yaml:",inline"`

	FileMetadata `json:"-" yaml:"-"`

	YAMLPath string `json:"-" yaml:"-"` // YAML path
}

func (f *Field) UnmarshalYAML(value *yaml.Node) error {
	// Prevent recursion by creating a new type that does not implement Unmarshaler.
	type notField Field
	x := (*notField)(f)

	if err := value.Decode(&x); err != nil {
		return err
	}
	f.FileMetadata.line = value.Line
	f.FileMetadata.column = value.Column
	return nil
}

// ReadFields reads all files matching the given globs and returns a slice
// containing all fields.
func ReadFields(globs ...string) ([]Field, error) {
	// Expand globs.
	var matches []string
	for _, glob := range globs {
		m, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}
		matches = append(matches, m...)
	}

	var fields []Field
	for _, file := range matches {
		tmpFields, err := readFields(file)
		if err != nil {
			return nil, err
		}
		fields = append(fields, tmpFields...)
	}

	return fields, nil
}

func readFields(path string) ([]Field, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fields []Field
	if err = readYAML(path, &fields, true); err != nil {
		return nil, fmt.Errorf("failed reading from %q: %w", path, err)
	}

	for i := range fields {
		annotateYAMLPath(fmt.Sprintf("$[%d]", i), &fields[i])
	}
	annotateFileMetadata(path, &fields)

	return fields, err
}

// annotateYAMLPath sets the YAML path of each field.
func annotateYAMLPath(yamlPath string, field *Field) {
	field.YAMLPath = yamlPath
	for i := range field.Fields {
		annotateYAMLPath(fmt.Sprintf("%s.fields[%d]", yamlPath, i), &field.Fields[i])
	}
}

// FlattenFields returns a flat representation of the fields. It
// removes all nested fields and creates top-level fields whose
// name is constructed by joining the parts with dots. The returned
// slice is sorted by name.
func FlattenFields(fields []Field) ([]Field, error) {
	var flat []Field
	for _, rootField := range fields {
		tmpFlats, err := flattenField(nil, rootField)
		if err != nil {
			return nil, err
		}
		flat = append(flat, tmpFlats...)
	}

	sort.Slice(flat, func(i, j int) bool {
		return flat[i].Name < flat[j].Name
	})

	return flat, nil
}

func flattenField(key []string, f Field) ([]Field, error) {
	// Leaf node.
	if len(f.Fields) == 0 {
		leafName := splitName(f.Name)

		name := make([]string, len(key)+len(leafName))
		copy(name, key)
		copy(name[len(key):], leafName)

		f.Name = strings.Join(name, ".")
		f.Fields = nil
		return []Field{f}, nil
	}

	parentName := append(key, splitName(f.Name)...)
	var flat []Field
	for _, child := range f.Fields {
		tmpFlats, err := flattenField(parentName, child)
		if err != nil {
			return nil, err
		}

		flat = append(flat, tmpFlats...)
	}

	// I would consider this to be an incorrect definition to
	// have sub-fields in a field not declared as a type=group.
	// This will include those fields in the list in order to not
	// mask bad definitions.
	if f.Type != "" && f.Type != "group" {
		parent := f
		parent.Name = strings.Join(parentName, ".")
		parent.Fields = nil
		flat = append(flat, parent)
	}

	return flat, nil
}

func splitName(name string) []string {
	return strings.Split(name, ".")
}
