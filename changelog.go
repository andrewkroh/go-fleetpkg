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

import "gopkg.in/yaml.v3"

type Changelog struct {
	Releases []Release `json:"releases,omitempty" yaml:"releases,omitempty"`

	sourceFile string
}

func (c *Changelog) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&c.Releases)
}

// Path returns the path to the changelog file.
func (c *Changelog) Path() string {
	return c.sourceFile
}

type Release struct {
	Version string   `json:"version,omitempty" yaml:"version,omitempty"`
	Changes []Change `json:"changes,omitempty" yaml:"changes,omitempty"`

	FileMetadata `json:"-" yaml:"-"`
}

func (r *Release) UnmarshalYAML(value *yaml.Node) error {
	// Prevent recursion by creating a new type that does not implement Unmarshaler.
	type notRelease Release
	x := (*notRelease)(r)

	if err := value.Decode(&x); err != nil {
		return err
	}
	r.FileMetadata.line = value.Line
	r.FileMetadata.column = value.Column
	return nil
}

type Change struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	Link        string `json:"link,omitempty" yaml:"link,omitempty"`

	FileMetadata `json:"-" yaml:"-"`
}

func (c *Change) UnmarshalYAML(value *yaml.Node) error {
	// Prevent recursion by creating a new type that does not implement Unmarshaler.
	type notChange Change
	x := (*notChange)(c)

	if err := value.Decode(&x); err != nil {
		return err
	}
	c.FileMetadata.line = value.Line
	c.FileMetadata.column = value.Column
	return nil
}
