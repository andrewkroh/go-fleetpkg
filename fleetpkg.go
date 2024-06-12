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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Integration struct {
	Build       *BuildManifest         `json:"build,omitempty" yaml:"build,omitempty"`
	Manifest    Manifest               `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Input       *DataStream            `json:"input,omitempty" yaml:"input,omitempty"`
	DataStreams map[string]*DataStream `json:"data_streams,omitempty" yaml:"data_streams,omitempty"`
	Changelog   Changelog              `json:"changelog,omitempty" yaml:"changelog,omitempty"`

	sourceFile string
}

// Path returns the path to the integration dir.
func (i Integration) Path() string {
	return i.sourceFile
}

type BuildManifest struct {
	Dependencies struct {
		ECS struct {
			// Source reference (pattern '^git@.+').
			Reference string `json:"reference,omitempty" yaml:"reference,omitempty"`
			// Whether to import common used dynamic templates and properties into the package.
			ImportMappings *bool `json:"import_mappings,omitempty" yaml:"import_mappings,omitempty"`
		} `json:"ecs,omitempty" yaml:"ecs,omitempty"`
	} `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	sourceFile string
}

// Path returns the path to the build.yml file.
func (m BuildManifest) Path() string {
	return m.sourceFile
}

type DataStream struct {
	Manifest    DataStreamManifest        `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	Pipelines   map[string]IngestPipeline `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	SampleEvent *SampleEvent              `json:"sample_event,omitempty" yaml:"sample_event,omitempty"`
	Fields      map[string]FieldsFile     `json:"fields,omitempty" yaml:"fields,omitempty"`

	sourceDir string
}

// Path returns the path to the data stream dir.
func (ds DataStream) Path() string {
	return ds.sourceDir
}

// AllFields returns a slice containing all fields declared in the DataStream.
func (ds DataStream) AllFields() []Field {
	var count int
	for _, ff := range ds.Fields {
		count += len(ff.Fields)
	}
	if count == 0 {
		return nil
	}

	out := make([]Field, 0, count)
	for _, ff := range ds.Fields {
		out = append(out, ff.Fields...)
	}
	return out
}

type FieldsFile struct {
	Fields []Field `json:"fields" yaml:"fields"`

	sourceFile string
}

// Path returns the path to the fields file.
func (f FieldsFile) Path() string {
	return f.sourceFile
}

type Manifest struct {
	Name            string                     `json:"name,omitempty" yaml:"name,omitempty"`
	Title           string                     `json:"title,omitempty" yaml:"title,omitempty"`
	Version         string                     `json:"version,omitempty" yaml:"version,omitempty"`
	Release         string                     `json:"release,omitempty" yaml:"release,omitempty"`
	Description     string                     `json:"description,omitempty" yaml:"description,omitempty"`
	Type            string                     `json:"type,omitempty" yaml:"type,omitempty"`
	Icons           []Icons                    `json:"icons,omitempty" yaml:"icons,omitempty"`
	FormatVersion   string                     `json:"format_version,omitempty" yaml:"format_version,omitempty"`
	License         string                     `json:"license,omitempty" yaml:"license,omitempty"`
	Categories      []string                   `json:"categories,omitempty" yaml:"categories,omitempty"`
	Conditions      Conditions                 `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Screenshots     []Screenshots              `json:"screenshots,omitempty" yaml:"screenshots,omitempty"`
	Source          Source                     `json:"source,omitempty" yaml:"source,omitempty"`
	Vars            []Var                      `json:"vars,omitempty" yaml:"vars,omitempty"`
	PolicyTemplates []PolicyTemplate           `json:"policy_templates,omitempty" yaml:"policy_templates,omitempty"`
	Owner           Owner                      `json:"owner,omitempty" yaml:"owner,omitempty"`
	Elasticsearch   *ElasticsearchRequirements `json:"elasticsearch,omitempty" yaml:"elasticsearch,omitempty"`
	Agent           *AgentRequirements         `json:"agent,omitempty" yaml:"agent,omitempty"`

	sourceFile string
}

// Path returns the path to the integration manifest.yml.
func (m Manifest) Path() string {
	return m.sourceFile
}

type ElasticsearchRequirements struct {
	Privileges ElasticsearchPrivilegeRequirements `json:"privileges,omitempty" yaml:"privileges,omitempty"`
}

type ElasticsearchPrivilegeRequirements struct {
	// Cluster privilege requirements.
	Cluster []string `json:"cluster,omitempty" yaml:"cluster,omitempty"`
}

// AgentRequirements declares related Agent configurations or requirements.
type AgentRequirements struct {
	Privileges AgentPrivilegeRequirements `json:"privileges,omitempty" yaml:"privileges,omitempty"`
}

type AgentPrivilegeRequirements struct {
	// Set to true if collection requires root privileges in the agent.
	Root bool `json:"root,omitempty" yaml:"root,omitempty"`
}

type Source struct {
	License string `json:"license,omitempty" yaml:"license,omitempty"`
}

type Icons struct {
	Src      string `json:"src,omitempty" yaml:"src,omitempty"`
	Title    string `json:"title,omitempty" yaml:"title,omitempty"`
	Size     string `json:"size,omitempty" yaml:"size,omitempty"`
	Type     string `json:"type,omitempty" yaml:"type,omitempty"`
	DarkMode *bool  `json:"dark_mode,omitempty" yaml:"dark_mode,omitempty"`
}

type Conditions struct {
	Elastic struct {
		Subscription string `json:"subscription,omitempty" yaml:"subscription,omitempty"`
	} `json:"elastic,omitempty" yaml:"elastic,omitempty"`
	Kibana struct {
		Version string `json:"version,omitempty" yaml:"version,omitempty"`
	} `json:"kibana,omitempty" yaml:"kibana,omitempty"`
}

// UnmarshalYAML implement special YAML unmarshal handling for Conditions
// to allow it to accept flattened key names which have been observed
// in packages.
func (c *Conditions) UnmarshalYAML(value *yaml.Node) error {
	type conditions Conditions // Type alias to prevent recursion.
	type permissiveConditions struct {
		KibanaVersion       string `yaml:"kibana.version,omitempty"`
		ElasticSubscription string `yaml:"elastic.subscription,omitempty"`
		conditions          `yaml:",inline"`
	}

	var pc permissiveConditions
	if err := value.Decode(&pc); err != nil {
		return err
	}

	if pc.Kibana.Version != "" {
		c.Kibana.Version = pc.Kibana.Version
	} else {
		c.Kibana.Version = pc.KibanaVersion
	}
	if pc.Elastic.Subscription != "" {
		c.Elastic.Subscription = pc.Elastic.Subscription
	} else {
		c.Elastic.Subscription = pc.ElasticSubscription
	}
	return nil
}

type Screenshots struct {
	Src   string `json:"src,omitempty" yaml:"src,omitempty"`
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	Size  string `json:"size,omitempty" yaml:"size,omitempty"`
	Type  string `json:"type,omitempty" yaml:"type,omitempty"`
}

type Var struct {
	Name        string   `json:"name,omitempty" yaml:"name,omitempty"`
	Default     any      `json:"default,omitempty" yaml:"default,omitempty"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string   `json:"type,omitempty" yaml:"type,omitempty"`
	Title       string   `json:"title,omitempty" yaml:"title,omitempty"`
	Multi       *bool    `json:"multi,omitempty" yaml:"multi,omitempty"`
	Required    *bool    `json:"required,omitempty" yaml:"required,omitempty"`
	Secret      *bool    `json:"secret,omitempty" yaml:"secret,omitempty"`
	ShowUser    *bool    `json:"show_user,omitempty" yaml:"show_user,omitempty"`
	Options     []Option `json:"options,omitempty" yaml:"options,omitempty"` // List of options for 'type: select'.

	FileMetadata `json:"-" yaml:"-"`
}

func (f *Var) UnmarshalYAML(value *yaml.Node) error {
	// Prevent recursion by creating a new type that does not implement Unmarshaler.
	type notVar Var
	x := (*notVar)(f)

	if err := value.Decode(&x); err != nil {
		return err
	}
	f.FileMetadata.line = value.Line
	f.FileMetadata.column = value.Column
	return nil
}

type Option struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`
	Text  string `json:"text,omitempty" yaml:"text,omitempty"`
}

type Input struct {
	Type         string `json:"type,omitempty" yaml:"type,omitempty"`
	Title        string `json:"title,omitempty" yaml:"title,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	InputGroup   string `json:"input_group,omitempty" yaml:"input_group,omitempty"`
	TemplatePath string `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	Multi        *bool  `json:"multi,omitempty" yaml:"multi,omitempty"`
	Vars         []Var  `json:"vars,omitempty" yaml:"vars,omitempty"`
}

type PolicyTemplate struct {
	Name         string        `json:"name,omitempty" yaml:"name,omitempty"`
	Title        string        `json:"title,omitempty" yaml:"title,omitempty"`
	Categories   []string      `json:"categories,omitempty" yaml:"categories,omitempty"`
	Description  string        `json:"description,omitempty" yaml:"description,omitempty"`
	DataStreams  []string      `json:"data_streams,omitempty" yaml:"data_streams,omitempty"`
	Inputs       []Input       `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Icons        []Icons       `json:"icons,omitempty" yaml:"icons,omitempty"`
	Screenshots  []Screenshots `json:"screenshots,omitempty" yaml:"screenshots,omitempty"`
	Multiple     *bool         `json:"multiple,omitempty" yaml:"multiple,omitempty"`
	Type         string        `json:"type,omitempty" yaml:"type,omitempty"` // Type of data stream.
	Input        string        `json:"input,omitempty" yaml:"input,omitempty"`
	TemplatePath string        `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	Vars         []Var         `json:"vars,omitempty" yaml:"vars,omitempty"` // Policy template level variables.
}

type Owner struct {
	Github string `json:"github,omitempty" yaml:"github,omitempty"`
	Type   string `json:"type,omitempty" yaml:"type,omitempty"` // Describes who owns the package and the level of support that is provided. Value may be elastic, partner, or community.
}

type DataStreamManifest struct {
	Dataset         string                 `json:"dataset,omitempty" yaml:"dataset,omitempty"`
	DatasetIsPrefix *bool                  `json:"dataset_is_prefix,omitempty" yaml:"dataset_is_prefix,omitempty"`
	ILMPolicy       string                 `json:"ilm_policy,omitempty" yaml:"ilm_policy,omitempty"`
	Release         string                 `json:"release,omitempty" yaml:"release,omitempty"`
	Title           string                 `json:"title,omitempty" yaml:"title,omitempty"`
	Type            string                 `json:"type,omitempty" yaml:"type,omitempty"`
	Streams         []Stream               `json:"streams,omitempty" yaml:"streams,omitempty"`
	Elasticsearch   *ElasticsearchSettings `json:"elasticsearch,omitempty" yaml:"elasticsearch,omitempty"`

	sourceFile string
}

type ElasticsearchSettings struct {
	IndexMode        string                   `json:"index_mode,omitempty" yaml:"index_mode,omitempty"`
	IndexTemplate    *IndexTemplateOptions    `json:"index_template,omitempty" yaml:"index_template,omitempty"`
	Privileges       *ElasticsearchPrivileges `json:"privileges,omitempty" yaml:"privileges,omitempty"`
	SourceMode       string                   `json:"source_mode,omitempty" yaml:"source_mode,omitempty"`
	DynamicDataset   *bool                    `json:"dynamic_dataset,omitempty" yaml:"dynamic_dataset,omitempty"`
	DynamicNamespace *bool                    `json:"dynamic_namespace,omitempty" yaml:"dynamic_namespace,omitempty"`
}

type IndexTemplateOptions struct {
	Settings       map[string]any         `json:"settings,omitempty" yaml:"settings,omitempty"`
	Mappings       map[string]any         `json:"mappings,omitempty" yaml:"mappings,omitempty"`
	IngestPipeline *IngestPipelineOptions `json:"ingest_pipeline,omitempty" yaml:"ingest_pipeline,omitempty"`
	DataStream     *DataStreamOptions     `json:"data_stream,omitempty" yaml:"data_stream,omitempty"`
}

type IngestPipelineOptions struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

type DataStreamOptions struct {
	Hidden *bool `json:"hidden,omitempty" yaml:"hidden,omitempty"`
}

type ElasticsearchPrivileges struct {
	Properties []string `json:"properties,omitempty" yaml:"properties,omitempty"`
}

func (m *DataStreamManifest) UnmarshalYAML(value *yaml.Node) error {
	type embeddedOptions DataStreamManifest // Type alias to prevent recursion.
	type permissiveOptions struct {
		DynamicDataset   *bool `yaml:"elasticsearch.dynamic_dataset"`
		DynamicNamespace *bool `yaml:"elasticsearch.dynamic_namespace"`
		embeddedOptions  `yaml:",inline"`
	}

	var options permissiveOptions
	if err := value.Decode(&options); err != nil {
		return err
	}
	*m = DataStreamManifest(options.embeddedOptions)

	if options.DynamicNamespace != nil {
		if m.Elasticsearch == nil {
			m.Elasticsearch = &ElasticsearchSettings{}
		}
		m.Elasticsearch.DynamicNamespace = options.DynamicNamespace
	}
	if options.DynamicDataset != nil {
		if m.Elasticsearch == nil {
			m.Elasticsearch = &ElasticsearchSettings{}
		}
		m.Elasticsearch.DynamicDataset = options.DynamicDataset
	}
	return nil
}

// Path returns the path to the data stream manifest.yml.
func (m DataStreamManifest) Path() string {
	return m.sourceFile
}

type Stream struct {
	Input        string `json:"input,omitempty" yaml:"input,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Title        string `json:"title,omitempty" yaml:"title,omitempty"`
	TemplatePath string `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	Vars         []Var  `json:"vars,omitempty" yaml:"vars,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

type SampleEvent struct {
	Event map[string]any `json:"event" yaml:"event"`

	sourceFile string
}

func (e SampleEvent) Path() string {
	return e.sourceFile
}

type IngestPipeline struct {
	// Description of the ingest pipeline.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Processors used to perform transformations on documents before indexing.
	// Processors run sequentially in the order specified.
	Processors []*Processor `json:"processors,omitempty" yaml:"processors,omitempty"`

	// Processors to run immediately after a processor failure.
	OnFailure []*Processor `json:"on_failure,omitempty" yaml:"on_failure,omitempty"`

	// Version number used by external systems to track ingest pipelines.
	Version *int `json:"version,omitempty" yaml:"version,omitempty"`

	// Optional metadata about the ingest pipeline. May have any contents.
	Meta map[string]any `json:"_meta,omitempty" yaml:"_meta,omitempty"`

	sourceFile string
}

// Path returns the path to the ingest node pipeline file.
func (p IngestPipeline) Path() string {
	return p.sourceFile
}

type Processor struct {
	Type       string
	Attributes map[string]any
}

func (p *Processor) UnmarshalYAML(value *yaml.Node) error {
	var procMap map[string]map[string]any
	if err := value.Decode(&procMap); err != nil {
		return err
	}

	// The struct representation used here is much more convenient
	// to work with than the original map of map format.
	for k, v := range procMap {
		p.Type = k
		p.Attributes = v
		break
	}

	return nil
}

func (p *Processor) MarshalYAML() (interface{}, error) {
	return map[string]any{
		p.Type: p.Attributes,
	}, nil
}

func (p *Processor) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		p.Type: p.Attributes,
	})
}

// Read reads the Fleet integration at the specified path. The path should
// point to the directory containing the integration's main manifest.yml.
func Read(path string) (*Integration, error) {
	integration := &Integration{
		DataStreams: map[string]*DataStream{},
		sourceFile:  path,
	}

	sourceFile := filepath.Join(path, "manifest.yml")
	if err := readYAML(sourceFile, &integration.Manifest, true); err != nil {
		return nil, err
	}
	integration.Manifest.sourceFile = sourceFile
	annotateFileMetadata(integration.Manifest.sourceFile, &integration.Manifest)
	sourceFile = filepath.Join(path, "changelog.yml")
	if err := readYAML(sourceFile, &integration.Changelog, true); err != nil {
		return nil, err
	}
	integration.Changelog.sourceFile = sourceFile
	annotateFileMetadata(integration.Changelog.sourceFile, &integration.Changelog)

	sourceFile = filepath.Join(path, "_dev/build/build.yml")
	if err := readYAML(sourceFile, &integration.Build, true); err != nil {
		// Optional file.
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	if integration.Build != nil {
		integration.Build.sourceFile = sourceFile
	}

	var dataStreams []string
	if integration.Manifest.Type == "input" {
		dataStreams = []string{filepath.Join(path, "manifest.yml")}
	} else {
		var err error
		dataStreams, err = filepath.Glob(filepath.Join(path, "data_stream/*/manifest.yml"))
		if err != nil {
			return nil, err
		}
	}
	for _, manifestPath := range dataStreams {
		ds := &DataStream{
			sourceDir: filepath.Dir(manifestPath),
		}
		if integration.Manifest.Type == "input" {
			integration.Input = ds
		} else {
			integration.DataStreams[filepath.Base(ds.sourceDir)] = ds
		}

		if err := readYAML(manifestPath, &ds.Manifest, true); err != nil {
			return nil, err
		}
		ds.Manifest.sourceFile = manifestPath
		annotateFileMetadata(ds.Manifest.sourceFile, &ds.Manifest)

		pipelines, err := filepath.Glob(filepath.Join(ds.sourceDir, "elasticsearch/ingest_pipeline/*.yml"))
		if err != nil {
			return nil, err
		}

		for _, pipelinePath := range pipelines {
			var pipeline IngestPipeline
			if err = readYAML(pipelinePath, &pipeline, true); err != nil {
				return nil, err
			}
			pipeline.sourceFile = pipelinePath

			if ds.Pipelines == nil {
				ds.Pipelines = map[string]IngestPipeline{}
			}
			ds.Pipelines[filepath.Base(pipelinePath)] = pipeline
		}

		// Sample event (optional).
		s := &SampleEvent{
			sourceFile: filepath.Join(ds.sourceDir, "sample_event.json"),
		}
		if err = readJSON(s.sourceFile, &s.Event, false); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
		if s.Event != nil {
			ds.SampleEvent = s
		}

		// Fields files.
		fieldsFiles, err := filepath.Glob(filepath.Join(ds.sourceDir, "fields/*.yml"))
		if err != nil {
			return nil, err
		}

		for _, fieldsFilePath := range fieldsFiles {
			fields, err := readFields(fieldsFilePath)
			if err != nil {
				return nil, err
			}

			if ds.Fields == nil {
				ds.Fields = map[string]FieldsFile{}
			}
			ds.Fields[filepath.Base(fieldsFilePath)] = FieldsFile{
				Fields:     fields,
				sourceFile: fieldsFilePath,
			}
		}
	}

	return integration, nil
}

func readYAML(path string, v any, strict bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	dec.KnownFields(strict)

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("failed decoding %s: %w", path, err)
	}
	return nil
}

func readJSON(path string, v any, strict bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	if strict {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("failed decoding %s: %w", path, err)
	}
	return nil
}
