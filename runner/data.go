package runner

import (
	"os"
	"path/filepath"
)

// defaultDataPaths returns paths to auto-discover when --data is not set.
//
// For each policy path that is a directory, a sibling "data" directory inside
// it is included if present. This makes it possible to publish a self-contained
// bundle as policy/ with policies and policy/data/ with data, and have
// conftest pick up both without requiring callers to add a -d flag.
//
// When --data is set explicitly, this discovery is skipped so the user's
// choice is authoritative.
func defaultDataPaths(policyPaths []string) []string {
	var paths []string
	seen := make(map[string]struct{})
	for _, p := range policyPaths {
		info, err := os.Stat(p)
		if err != nil || !info.IsDir() {
			continue
		}
		candidate := filepath.Join(p, "data")
		if _, ok := seen[candidate]; ok {
			continue
		}
		dataInfo, err := os.Stat(candidate)
		if err != nil || !dataInfo.IsDir() {
			continue
		}
		seen[candidate] = struct{}{}
		paths = append(paths, candidate)
	}
	return paths
}
