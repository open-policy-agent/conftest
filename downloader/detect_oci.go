package downloader

import (
	"fmt"
	"regexp"
	"strings"
)

var matchRegistries = []*regexp.Regexp{
	regexp.MustCompile("azurecr.io"),
	regexp.MustCompile("gcr.io"),
	regexp.MustCompile("registry.gitlab.com"),
	regexp.MustCompile("[0-9]{12}.dkr.ecr.[a-z0-9-]*.amazonaws.com"),
}

// OCIDetector implements Detector to detect OCI registry URLs and turn
// them into URLs that the OCI getter can understand.
type OCIDetector struct{}

// Detect will detect if the source is an OCI registry
func (d *OCIDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if containsOCIRegistry(src) || containsLocalRegistry(src) {
		url, err := d.detectHTTP(src)
		if err != nil {
			return "", false, fmt.Errorf("detect http: %w", err)
		}

		return url, true, nil
	}

	return "", false, nil
}

func containsOCIRegistry(src string) bool {
	for _, matchRegistry := range matchRegistries {
		if matchRegistry.MatchString(src) {
			return true
		}
	}

	return false
}

func containsLocalRegistry(src string) bool {
	return strings.Contains(src, "127.0.0.1:5000") || strings.Contains(src, "localhost:5000")
}

func (d *OCIDetector) detectHTTP(src string) (string, error) {
	parts := strings.Split(src, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf(
			"URL is not a valid Azure registry URL")
	}

	return "oci://" + getRepositoryFromURL(src), nil
}

func getRepositoryFromURL(url string) string {
	if repositoryContainsTag(url) {
		return url
	}

	return url + ":latest"
}

func repositoryContainsTag(repository string) bool {
	path := strings.Split(repository, "/")
	return pathContainsTag(path[len(path)-1])
}

func pathContainsTag(path string) bool {
	return strings.Contains(path, ":")
}
