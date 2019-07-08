package util

import "strings"

// RepositoryNameContainsTag checks if the repository name includes a tag
func RepositoryNameContainsTag(name string) bool {
	split := strings.Split(name, "/")
	return strings.Contains(split[len(split)-1], ":")
}
