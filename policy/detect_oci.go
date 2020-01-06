package policy

import (
	"fmt"
	"strings"
)

// OCIDetector implements Detector to detect OCI registry URLs and turn
// them into URLs that the OCI getter can understand.
type OCIDetector struct{}

func (d *OCIDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if strings.Contains(src, "azurecr.io/") || strings.Contains(src, "127.0.0.1:5000") {
		url, err := d.detectHTTP(src)
		if err != nil {
			return "", false, err
		}

		return url, true, nil
	}

	return "", false, nil
}

func (d *OCIDetector) detectHTTP(src string) (string, error) {
	// Check validity of url and tag with :latest if no tag is available
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
