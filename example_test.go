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

package fleetpkg_test

import (
	"fmt"
	"log"

	"github.com/andrewkroh/go-fleetpkg"
)

func ExampleRead() {
	// Read an integration package without any options
	pkg, err := fleetpkg.Read("testdata/packages/cel")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Package: %s\n", pkg.Manifest.Name)
	fmt.Printf("Version: %s\n", pkg.Manifest.Version)
	fmt.Printf("Releases: %d\n", len(pkg.Changelog.Releases))

	// Output:
	// Package: cel
	// Version: 0.3.0
	// Releases: 4
}

func ExampleRead_withChangelogDates() {
	// Read an integration package with changelog dates enabled
	pkg, err := fleetpkg.Read("testdata/packages/cel", fleetpkg.WithChangelogDates())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Package: %s\n", pkg.Manifest.Name)

	// With WithChangelogDates(), each release will have a Date field populated
	// based on when the release was added to the changelog (via git blame)
	for _, release := range pkg.Changelog.Releases {
		if release.Date != nil {
			fmt.Printf("Release %s was added on %s\n", release.Version, release.Date.Format("2006-01-02"))
		}
	}
}
