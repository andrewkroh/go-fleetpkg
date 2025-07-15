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

// Field represents an Elasticsearch field definition in the Elastic package
// specification. It contains all the properties that can be used to define how a
// field is mapped, indexed, and used in Elasticsearch.
//
// Field definitions can be nested, with parent fields containing child fields.
// Each field must have a name, and most fields will have a type that determines
// how the field is indexed and stored in Elasticsearch.
type Field struct {
	Name                  string   `json:"name,omitempty" yaml:"name,omitempty"`                                         // Name of the field
	Type                  string   `json:"type,omitempty" yaml:"type,omitempty"`                                         // Type of the field as used in Elasticsearch
	Description           string   `json:"description,omitempty" yaml:"description,omitempty"`                           // Description of the field
	Value                 string   `json:"value,omitempty" yaml:"value,omitempty"`                                       // Constant value assigned to this field
	Example               any      `json:"example,omitempty" yaml:"example,omitempty"`                                   // Example value for this field, used in documentation
	MetricType            string   `json:"metric_type,omitempty" yaml:"metric_type,omitempty"`                           // For metric fields, defines what kind of metric this is
	Unit                  string   `json:"unit,omitempty" yaml:"unit,omitempty"`                                         // Unit this field is measured in
	DateFormat            string   `json:"date_format,omitempty" yaml:"date_format,omitempty"`                           // Format of the date string in this field
	Dimension             *bool    `json:"dimension,omitempty" yaml:"dimension,omitempty"`                               // Identifies a field to be a dimension for grouping when a metric is associated with multiple dimensions
	Pattern               string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`                                   // Regular expression pattern for validating field values
	External              string   `json:"external,omitempty" yaml:"external,omitempty"`                                 // Identifier for the type of metric when referencing an external schema
	Fields                []Field  `json:"fields,omitempty" yaml:"fields,omitempty"`                                     // Nested fields within this field
	DocValues             *bool    `json:"doc_values,omitempty" yaml:"doc_values,omitempty"`                             // Controls whether the field is indexed in a column-stride fashion for sorting and aggregations
	Index                 *bool    `json:"index,omitempty" yaml:"index,omitempty"`                                       // Controls whether the field will be indexed for full-text search
	CopyTo                string   `json:"copy_to,omitempty" yaml:"copy_to,omitempty"`                                   // Target field to copy this field's values to
	Enabled               *bool    `json:"enabled,omitempty" yaml:"enabled,omitempty"`                                   // Whether mappings are created for this field's children
	Dynamic               string   `json:"dynamic,omitempty" yaml:"dynamic,omitempty"`                                   // Controls whether new fields are added dynamically or ignored if not defined
	ScalingFactor         *int     `json:"scaling_factor,omitempty" yaml:"scaling_factor,omitempty"`                     // Scaling factor to use for scaled_float type
	Analyzer              string   `json:"analyzer,omitempty" yaml:"analyzer,omitempty"`                                 // Analyzer used for full-text search
	SearchAnalyzer        string   `json:"search_analyzer,omitempty" yaml:"search_analyzer,omitempty"`                   // Analyzer to use at search time
	MultiFields           []Field  `json:"multi_fields,omitempty" yaml:"multi_fields,omitempty"`                         // Sub-fields to index the same value in different ways
	NullValue             string   `json:"null_value,omitempty" yaml:"null_value,omitempty"`                             // Value to replace null with when indexing
	IgnoreMalformed       *bool    `json:"ignore_malformed,omitempty" yaml:"ignore_malformed,omitempty"`                 // Whether to ignore malformed values in the field
	IgnoreAbove           int      `json:"ignore_above,omitempty" yaml:"ignore_above,omitempty"`                         // String values longer than this will not be indexed or stored
	ObjectType            string   `json:"object_type,omitempty" yaml:"object_type,omitempty"`                           // Type of the object field
	ObjectTypeMappingType string   `json:"object_type_mapping_type,omitempty" yaml:"object_type_mapping_type,omitempty"` // Mapping type for the object field
	AliasTargetPath       string   `json:"path,omitempty" yaml:"path,omitempty"`                                         // For alias type fields this is the path to the target field, including parent objects
	Normalize             []string `json:"normalize,omitempty" yaml:"normalize,omitempty"`                               // Specifies the expected normalizations for a field, such as 'array' normalization
	Normalizer            string   `json:"normalizer,omitempty" yaml:"normalizer,omitempty"`                             // Specifies the name of a normalizer to apply to keyword fields
	IncludeInParent       *bool    `json:"include_in_parent,omitempty" yaml:"include_in_parent,omitempty"`               // For nested field types, specifies if fields in the nested object are also added to the parent document
	DefaultMetric         string   `json:"default_metric,omitempty" yaml:"default_metric,omitempty"`                     // For aggregate_metric_double fields, specifies the default metric aggregation

	// Specifies if field names containing dots should be expanded into
	// subobjects. For example, if this is set to `true`, a field named `foo.bar`
	// will be expanded into an object with a field named `bar` inside an object
	// named `foo`. Defaults to true.
	Subobjects *bool `json:"subobjects,omitempty" yaml:"subobjects,omitempty"`

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

	// I would consider this to be an incorrect definition to have sub-fields in
	// a field not declared as a type=group. This will include those fields in
	// the list to not mask bad definitions.
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
