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
	"testing"
)

func TestFlattenFields(t *testing.T) {
	p, err := Read("testdata/packages/hashicorp_vault")
	if err != nil {
		t.Fatal(err)
	}

	for _, ds := range p.DataStreams {
		var allFields []Field
		for _, ff := range ds.Fields {
			allFields = append(allFields, ff.Fields...)
		}

		flat, err := FlattenFields(allFields)
		if err != nil {
			t.Fatal(err)
		}

		if len(flat) == 0 {
			t.Fatal("Returned flattened fields are empty.")
		}

		for _, f := range flat {
			if f.Fields != nil {
				t.Errorf("Unexpected nested fields in %v.", f.Name)
			}
		}
	}
}

func TestReadFields(t *testing.T) {
	allFields, err := ReadFields(
		"testdata/packages/*/fields/*.yml",
		"testdata/packages/*/data_stream/*/fields/*.yml")
	if err != nil {
		t.Fatal(err)
	}

	const expected = 154
	if len(allFields) != expected {
		t.Fatalf("got %d, want %d fields", len(allFields), expected)
	}

	allFields, err = FlattenFields(allFields)
	if err != nil {
		t.Fatal(err)
	}

	for _, f := range allFields {
		if f.Path() == "" {
			t.Errorf("sourceFile is empty")
		}
		if f.Line() == 0 {
			t.Errorf("sourceLine is empty")
		}
		if f.Column() == 0 {
			t.Errorf("sourceColumn is empty")
		}
		fmt.Printf("%s:%d:%d %s - %s\n", f.Path(), f.Line(), f.Column(), f.Name, f.YAMLPath)
		if f.YAMLPath == "" {
			t.Errorf("YAMLPath is empty")
		}
	}
}
