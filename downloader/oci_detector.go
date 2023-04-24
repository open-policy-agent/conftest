package downloader

import (
	"fmt"
	"regexp"
	"strings"
)

// OCIDetector implements Detector to detect OCI registry URLs and turn
// them into URLs that the OCI getter can understand.
type OCIDetector struct{}

// Detect will detect if the source is an OCI registry
func (d *OCIDetector) Detect(src, _ string) (string, bool, error) {
	if len(src) == 0 {
		return "", false, nil
	}

	if containsOCIRegistry(src) || containsLocalRegistry(src) {
		url, err := detectHTTP(src)
		if err != nil {
			return "", false, fmt.Errorf("detect http: %w", err)
		}

		return url, true, nil
	}

	return "", false, nil
}

func containsOCIRegistry(src string) bool {
	matchRegistries := []*regexp.Regexp{
		regexp.MustCompile("azurecr.io"),
		regexp.MustCompile("gcr.io"),
		regexp.MustCompile("registry.gitlab.com"),
		regexp.MustCompile("[0-9]{12}.dkr.ecr.[a-z0-9-]*.amazonaws.com"),
		regexp.MustCompile("^quay.io"),
	}

	for _, matchRegistry := range matchRegistries {
		if matchRegistry.MatchString(src) {
			return true
		}
	}

	return false
}

func containsLocalRegistry(src string) bool {
	matched, err := regexp.MatchString(`(?:::1|127\.0\.0\.1|(?i:localhost)):\d{1,5}`, src)

	return err == nil && matched
}

func detectHTTP(src string) (string, error) {
	parts := strings.Split(src, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("URL is not a valid registry URL")
	}

	repository := getRepositoryFromURL(src)
	return "oci://" + repository, nil
}

func getRepositoryFromURL(url string) string {
	pathParts := strings.Split(url, "/")
	lastPathPart := pathParts[len(pathParts)-1]

	if strings.Contains(lastPathPart, ":") {
		return url
	}

	return url + ":latest"
}
