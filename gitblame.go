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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// gitBlameCache caches git blame results per file to avoid repeated git invocations.
type gitBlameCache struct {
	// Map of file path to line number to timestamp
	data map[string]map[int]time.Time
}

// newGitBlameCache creates a new cache instance.
func newGitBlameCache() *gitBlameCache {
	return &gitBlameCache{
		data: make(map[string]map[int]time.Time),
	}
}

// getBlameForFile runs git blame once for the entire file and caches all line dates.
// Returns a map of line number to timestamp.
func (c *gitBlameCache) getBlameForFile(filePath string) (map[int]time.Time, error) {
	// Check if already cached
	if lineMap, exists := c.data[filePath]; exists {
		return lineMap, nil
	}

	// Get the directory containing the file to use as working directory for git
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

	// Run git blame with porcelain format for the entire file
	cmd := exec.Command("git", "-C", dir, "blame", "--porcelain", filename)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame failed for %s: %w", filePath, err)
	}

	// Parse the porcelain format output
	// Format: Each line starts with a header like "<sha> <original-line> <final-line> [<num-lines>]"
	// The first occurrence of a SHA includes metadata lines like "author-time <timestamp>"
	// Subsequent lines from the same commit just have "<sha> <original-line> <final-line>"
	// Each entry ends with the actual source line prefixed with a tab
	lineMap := make(map[int]time.Time)
	lines := strings.Split(string(output), "\n")

	// Map SHA to timestamp for caching within the file
	shaToTimestamp := make(map[string]int64)

	i := 0
	for i < len(lines) {
		line := lines[i]

		// Check if this is a commit header line: <sha> <original-line> <final-line> [<num-lines>]
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			sha := parts[0]

			// Parse line number from the third field (final-line)
			lineNum, err := strconv.Atoi(parts[2])
			if err != nil {
				i++
				continue
			}

			// Check if we already have the timestamp for this SHA
			timestamp, exists := shaToTimestamp[sha]

			if !exists {
				// Read metadata lines until we find author-time or reach the source line (starts with tab)
				i++
				for i < len(lines) && !strings.HasPrefix(lines[i], "\t") {
					if strings.HasPrefix(lines[i], "author-time ") {
						timestampStr := strings.TrimPrefix(lines[i], "author-time ")
						timestamp, _ = strconv.ParseInt(timestampStr, 10, 64)
						shaToTimestamp[sha] = timestamp
					}

					// Check if next line is another commit header (starts with a hex SHA)
					if i+1 < len(lines) {
						nextParts := strings.Fields(lines[i+1])
						if len(nextParts) >= 3 {
							// Try to parse third field as line number - if successful, it's a header
							if _, err := strconv.Atoi(nextParts[2]); err == nil {
								break
							}
						}
					}
					i++
				}
			}

			if timestamp > 0 {
				lineMap[lineNum] = time.Unix(timestamp, 0).UTC()
			}
		}
		i++
	}

	c.data[filePath] = lineMap
	return lineMap, nil
}

// getDateForLine returns the date for a specific line in a file.
func (c *gitBlameCache) getDateForLine(filePath string, line int) (time.Time, error) {
	if filePath == "" || line <= 0 {
		return time.Time{}, nil
	}

	lineMap, err := c.getBlameForFile(filePath)
	if err != nil {
		return time.Time{}, err
	}

	if date, ok := lineMap[line]; ok {
		return date, nil
	}

	return time.Time{}, nil
}

// annotateReleaseDates populates the Date field of all releases in a changelog
// using git blame to determine when each release line was added.
// It uses a cache to run git blame once per file for efficiency.
func annotateReleaseDates(changelog *Changelog) error {
	if changelog == nil {
		return nil
	}

	cache := newGitBlameCache()

	for i := range changelog.Releases {
		release := &changelog.Releases[i]
		if release.Path() != "" && release.Line() > 0 {
			date, err := cache.getDateForLine(release.Path(), release.Line())
			if err != nil {
				return fmt.Errorf("failed to get git blame date for release %s: %w", release.Version, err)
			}
			if !date.IsZero() {
				release.Date = &date
			}
		}
	}

	return nil
}
