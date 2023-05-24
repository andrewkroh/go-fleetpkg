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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Integration struct {
	Build       BuildManifest          `json:"build,omitempty" yaml:"build,omitempty"`
	Manifest    Manifest               `json:"manifest,omitempty" yaml:"manifest,omitempty"`
	DataStreams map[string]*DataStream `json:"data_streams,omitempty" yaml:"data_streams,omitempty"`
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
}

type DataStream struct {
	Manifest DataStreamManifest `json:"manifest,omitempty" yaml:"manifest,omitempty"`
}

type Manifest struct {
	Name            string           `json:"name,omitempty" yaml:"name,omitempty"`
	Title           string           `json:"title,omitempty" yaml:"title,omitempty"`
	Version         string           `json:"version,omitempty" yaml:"version,omitempty"`
	Release         string           `json:"release,omitempty" yaml:"release,omitempty"`
	Description     string           `json:"description,omitempty" yaml:"description,omitempty"`
	Type            string           `json:"type,omitempty" yaml:"type,omitempty"`
	Icons           []Icons          `json:"icons,omitempty" yaml:"icons,omitempty"`
	FormatVersion   string           `json:"format_version,omitempty" yaml:"format_version,omitempty"`
	License         string           `json:"license,omitempty" yaml:"license,omitempty"`
	Categories      []string         `json:"categories,omitempty" yaml:"categories,omitempty"`
	Conditions      Conditions       `json:"conditions,omitempty" yaml:"conditions,omitempty"`
	Screenshots     []Screenshots    `json:"screenshots,omitempty" yaml:"screenshots,omitempty"`
	Source          Source           `json:"source,omitempty" yaml:"source,omitempty"`
	Vars            []Var            `json:"vars,omitempty" yaml:"vars,omitempty"`
	PolicyTemplates []PolicyTemplate `json:"policy_templates,omitempty" yaml:"policy_templates,omitempty"`
	Owner           Owner            `json:"owner,omitempty" yaml:"owner,omitempty"`
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
	ShowUser    *bool    `json:"show_user,omitempty" yaml:"show_user,omitempty"`
	Options     []Option `json:"options,omitempty" yaml:"options,omitempty"` // List of options for 'type: select'.
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
}

type DataStreamManifest struct {
	Dataset         string         `json:"dataset,omitempty" yaml:"dataset,omitempty"`
	DatasetIsPrefix *bool          `json:"dataset_is_prefix,omitempty" yaml:"dataset_is_prefix,omitempty"`
	ILMPolicy       string         `json:"ilm_policy,omitempty" yaml:"ilm_policy,omitempty"`
	Release         string         `json:"release,omitempty" yaml:"release,omitempty"`
	Title           string         `json:"title,omitempty" yaml:"title,omitempty"`
	Type            string         `json:"type,omitempty" yaml:"type,omitempty"`
	Streams         []Stream       `json:"streams,omitempty" yaml:"streams,omitempty"`
	Elasticsearch   map[string]any `json:"elasticsearch,omitempty" yaml:"elasticsearch,omitempty"`
}

type Stream struct {
	Input        string `json:"input,omitempty" yaml:"input,omitempty"`
	Description  string `json:"description,omitempty" yaml:"description,omitempty"`
	Title        string `json:"title,omitempty" yaml:"title,omitempty"`
	TemplatePath string `json:"template_path,omitempty" yaml:"template_path,omitempty"`
	Vars         []Var  `json:"vars,omitempty" yaml:"vars,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

// Read reads the Fleet integration at the specified path. The path should
// point to the directory containing the integration's main manifest.yml.
func Read(path string) (*Integration, error) {
	integration := &Integration{
		DataStreams: map[string]*DataStream{},
	}

	if err := readYAML(filepath.Join(path, "manifest.yml"), &integration.Manifest, true); err != nil {
		return nil, err
	}

	if err := readYAML(filepath.Join(path, "_dev/build/build.yml"), &integration.Build, true); err != nil {
		// Optional file.
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	dataStreams, err := filepath.Glob(filepath.Join(path, "data_stream/*/manifest.yml"))
	if err != nil {
		return nil, err
	}
	for _, f := range dataStreams {
		var dataStreamManifest DataStreamManifest
		if err := readYAML(f, &dataStreamManifest, true); err != nil {
			return nil, err
		}
		integration.DataStreams[filepath.Base(filepath.Dir(f))] = &DataStream{
			Manifest: dataStreamManifest,
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
