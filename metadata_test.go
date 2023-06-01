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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotateFileMetadata(t *testing.T) {
	integ := Integration{
		Manifest: Manifest{
			Vars: []Var{
				{
					Name: "proxy_url",
				},
			},
			PolicyTemplates: []PolicyTemplate{
				{
					Vars: []Var{
						{
							Name: "proxy_url",
						},
					},
					Inputs: []Input{
						{
							Vars: []Var{
								{
									Name: "proxy_url",
								},
							},
						},
					},
				},
			},
		},
		DataStreams: map[string]*DataStream{
			"log": {
				Manifest: DataStreamManifest{
					Streams: []Stream{
						{
							Vars: []Var{
								{
									Name: "interval",
								},
							},
						},
					},
				},
			},
		},
	}

	const file = "manifest.yml"
	annotateFileMetadata(file, &integ)
	assert.Equal(t, file, integ.Manifest.Vars[0].file)
	assert.Equal(t, file, integ.Manifest.PolicyTemplates[0].Vars[0].file)
	assert.Equal(t, file, integ.Manifest.PolicyTemplates[0].Inputs[0].Vars[0].file)
	assert.Equal(t, file, integ.DataStreams["log"].Manifest.Streams[0].Vars[0].file)
}
