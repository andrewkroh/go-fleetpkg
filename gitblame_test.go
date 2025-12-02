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
	"time"
)

func TestGitBlameDate(t *testing.T) {
	// Test with this file itself - should get a non-zero date
	cache := newGitBlameCache()
	date, err := cache.getDateForLine("gitblame_test.go", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// If in a git repo, should get a valid date
	if !date.IsZero() {
		t.Logf("Got date: %v", date)
		if date.After(time.Now()) {
			t.Errorf("date is in the future: %v", date)
		}
	}
}

func TestAnnotateReleaseDates(t *testing.T) {
	// Test with one of the test packages using WithChangelogDates option
	pkg, err := Read("testdata/packages/cel", WithChangelogDates())
	if err != nil {
		t.Fatalf("failed to read package: %v", err)
	}

	// Check that some releases have dates
	var datedReleases int
	for _, release := range pkg.Changelog.Releases {
		if release.Date != nil && !release.Date.IsZero() {
			datedReleases++
			t.Logf("Release %s: %v", release.Version, *release.Date)
		}
	}

	if datedReleases == 0 {
		t.Error("Expected at least some releases to have dates when WithChangelogDates() is used")
	}

	t.Logf("Total releases: %d, Releases with dates: %d", len(pkg.Changelog.Releases), datedReleases)
}

func TestWithoutChangelogDates(t *testing.T) {
	// Test without the option - dates should not be populated
	pkg, err := Read("testdata/packages/cel")
	if err != nil {
		t.Fatalf("failed to read package: %v", err)
	}

	// Check that no releases have dates
	for _, release := range pkg.Changelog.Releases {
		if release.Date != nil {
			t.Errorf("Release %s has a date %v, but WithChangelogDates() was not used", release.Version, *release.Date)
		}
	}

	t.Logf("Confirmed: All %d releases have nil dates (as expected)", len(pkg.Changelog.Releases))
}
