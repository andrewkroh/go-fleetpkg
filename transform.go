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

// Transform represents an Elasticsearch transform configuration within an
// integration package. It contains the transform definition, manifest settings,
// field mappings, and source directory path.
type Transform struct {
	Transform *ElasticsearchTransform `json:"transform,omitempty" yaml:"transform,omitempty"`
	Manifest  *TransformManifest      `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Fields    []Field                 `json:"fields,omitempty" yaml:"fields,omitempty"`

	sourceDir string
}

// Path returns the source directory path where the transform configuration is
// located.
func (t *Transform) Path() string {
	return t.sourceDir
}

type ElasticsearchTransform struct {
	// Source defines the source configuration for the transform
	Source *TransformSource `json:"source,omitempty" yaml:"source,omitempty"`

	// Dest defines the destination configuration for the transform
	Dest *TransformDest `json:"dest,omitempty" yaml:"dest,omitempty"`

	// Pivot defines the pivot configuration for aggregating data
	Pivot *TransformPivot `json:"pivot,omitempty" yaml:"pivot,omitempty"`

	// Latest defines the latest configuration for getting the most recent documents
	Latest *TransformLatest `json:"latest,omitempty" yaml:"latest,omitempty"`

	// Description provides a human-readable description of the transform
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`

	// Frequency defines how often the transform should run
	Frequency *string `json:"frequency,omitempty" yaml:"frequency,omitempty"`

	// ID is the unique identifier for the transform
	ID *string `json:"id,omitempty" yaml:"id,omitempty"`

	// Settings contains various transform settings
	Settings *TransformSettings `json:"settings,omitempty" yaml:"settings,omitempty"`

	// Meta contains arbitrary metadata for the transform
	Meta map[string]interface{} `json:"_meta,omitempty" yaml:"_meta,omitempty"`

	// RetentionPolicy defines the retention policy for the transform
	RetentionPolicy *TransformRetentionPolicy `json:"retention_policy,omitempty" yaml:"retention_policy,omitempty"`

	// Sync defines the sync configuration for continuous transforms
	Sync *TransformSync `json:"sync,omitempty" yaml:"sync,omitempty"`
}

// TransformSource represents the source configuration for a transform
type TransformSource struct {
	// Index specifies the source index or indices
	Index interface{} `json:"index,omitempty" yaml:"index,omitempty"` // Can be string or []string

	// Query defines the query to filter source documents
	Query map[string]interface{} `json:"query,omitempty" yaml:"query,omitempty"`

	// RuntimeMappings defines runtime field mappings
	RuntimeMappings map[string]interface{} `json:"runtime_mappings,omitempty" yaml:"runtime_mappings,omitempty"`
}

// TransformDest represents the destination configuration for a transform
type TransformDest struct {
	// Index is the destination index name
	Index *string `json:"index,omitempty" yaml:"index,omitempty"`

	// Pipeline is the ingest pipeline to use for the destination
	Pipeline *string `json:"pipeline,omitempty" yaml:"pipeline,omitempty"`

	// Aliases defines the aliases for the destination index
	Aliases []TransformDestAlias `json:"aliases,omitempty" yaml:"aliases,omitempty"`
}

// TransformDestAlias represents an alias configuration for the destination index
type TransformDestAlias struct {
	// Alias is the name of the alias
	Alias *string `json:"alias,omitempty" yaml:"alias,omitempty"`

	// MoveOnCreation indicates whether the destination index should be the only index in this alias
	MoveOnCreation *bool `json:"move_on_creation,omitempty" yaml:"move_on_creation,omitempty"`
}

// TransformPivot represents the pivot configuration for a transform
type TransformPivot struct {
	// GroupBy defines the grouping configuration
	GroupBy map[string]interface{} `json:"group_by,omitempty" yaml:"group_by,omitempty"`

	// Aggregations defines the aggregations to perform
	Aggregations map[string]interface{} `json:"aggregations,omitempty" yaml:"aggregations,omitempty"`

	// Aggs is an alternative name for aggregations
	Aggs map[string]interface{} `json:"aggs,omitempty" yaml:"aggs,omitempty"`
}

// TransformLatest represents the latest configuration for a transform
type TransformLatest struct {
	// Sort defines the sort field for determining the latest documents
	Sort *string `json:"sort,omitempty" yaml:"sort,omitempty"`

	// UniqueKey defines the unique key fields
	UniqueKey []string `json:"unique_key,omitempty" yaml:"unique_key,omitempty"`
}

// TransformSettings represents the settings configuration for a transform
type TransformSettings struct {
	// DatesAsEpochMillis indicates whether dates should be stored as epoch milliseconds
	DatesAsEpochMillis *bool `json:"dates_as_epoch_millis,omitempty" yaml:"dates_as_epoch_millis,omitempty"`

	// DocsPerSecond limits the number of documents processed per second
	DocsPerSecond *float64 `json:"docs_per_second,omitempty" yaml:"docs_per_second,omitempty"`

	// AlignCheckpoints indicates whether checkpoints should be aligned
	AlignCheckpoints *bool `json:"align_checkpoints,omitempty" yaml:"align_checkpoints,omitempty"`

	// MaxPageSearchSize defines the maximum page size for search requests
	MaxPageSearchSize *int `json:"max_page_search_size,omitempty" yaml:"max_page_search_size,omitempty"`

	// UsePointInTime indicates whether to use point-in-time for search requests
	UsePointInTime *bool `json:"use_point_in_time,omitempty" yaml:"use_point_in_time,omitempty"`

	// DeduceMappings indicates whether to deduce mappings automatically
	DeduceMappings *bool `json:"deduce_mappings,omitempty" yaml:"deduce_mappings,omitempty"`

	// Unattended indicates whether the transform runs in unattended mode
	Unattended *bool `json:"unattended,omitempty" yaml:"unattended,omitempty"`
}

// TransformRetentionPolicy represents the retention policy configuration for a transform
type TransformRetentionPolicy struct {
	// Time defines the time-based retention policy
	Time *TransformTimeRetentionPolicy `json:"time,omitempty" yaml:"time,omitempty"`
}

// TransformTimeRetentionPolicy represents the time-based retention policy configuration
type TransformTimeRetentionPolicy struct {
	// Field is the field used for time-based retention
	Field *string `json:"field,omitempty" yaml:"field,omitempty"`

	// MaxAge defines the maximum age for retaining documents
	MaxAge *string `json:"max_age,omitempty" yaml:"max_age,omitempty"`
}

// TransformSync represents the sync configuration for continuous transforms
type TransformSync struct {
	// Time defines the time-based sync configuration
	Time *TransformSyncTime `json:"time,omitempty" yaml:"time,omitempty"`
}

// TransformSyncTime represents the time-based sync configuration
type TransformSyncTime struct {
	// Field is the field used for time-based synchronization
	Field *string `json:"field,omitempty" yaml:"field,omitempty"`

	// Delay defines the delay for synchronization
	Delay *string `json:"delay,omitempty" yaml:"delay,omitempty"`
}

type TransformManifest struct {
	DestinationIndexTemplate *IndexTemplate `json:"destination_index_template,omitempty" yaml:"destination_index_template,omitempty"`
	Start                    *bool          `json:"start,omitempty" yaml:"start,omitempty"`
}

// IndexTemplate contains a subset of the index template format that can be
// customized by an integration.
type IndexTemplate struct {
	Mappings *Mappings `json:"mappings,omitempty" yaml:"mappings,omitempty"`
	Settings *Settings `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// Mappings are Elasticsearch mappings.
// https://www.elastic.co/guide/en/elasticsearch/reference/current/mapping.html
type Mappings struct {
	Properties map[string]any `json:"properties,omitempty" yaml:"properties,omitempty"`
}

// Settings are Elasticsearch index settings.
// https://www.elastic.co/guide/en/elasticsearch/reference/current/index-modules.html#index-modules-settings
type Settings map[string]any
