package policy

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"

	auth "github.com/deislabs/oras/pkg/auth/docker"
)

// Policy represents a policy
type Policy struct {
	Repository string
	Tag        string
}

// Download downloads the given policies
func Download(ctx context.Context, path string, policies []Policy) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("make policy directory: %w", err)
	}

	cli, err := auth.NewClient()
	if err != nil {
		return fmt.Errorf("new auth client: %w", err)
	}

	resolver, err := cli.Resolver(ctx)
	if err != nil {
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	fileStore := content.NewFileStore(path)
	defer fileStore.Close()

	for _, policy := range policies {
		repository := getRepositoryFromPolicy(policy)

		_, _, err = oras.Pull(ctx, resolver, repository, fileStore)
		if err != nil {
			return fmt.Errorf("pulling policy: %w", err)
		}
	}

	return nil
}

func getRepositoryFromPolicy(policy Policy) string {
	var repository string
	if repositoryContainsTag(policy.Repository) {
		repository = policy.Repository
	} else if policy.Tag == "" {
		repository = policy.Repository + ":latest"
	} else {
		repository = policy.Repository + ":" + policy.Tag
	}

	return repository
}

func repositoryContainsTag(repository string) bool {
	split := strings.Split(repository, "/")
	return strings.Contains(split[len(split)-1], ":")
}

// ReadFiles returns all of the policy files (not including tests)
// at the given path including its subdirectories.
func ReadFiles(path string) ([]string, error) {
	files, err := getPolicyFiles(path)
	if err != nil {
		return nil, fmt.Errorf("search rego files: %w", err)
	}

	return files, nil
}

func getPolicyFiles(path string) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(currentPath) == ".rego" && !strings.HasSuffix(info.Name(), "_test.rego") {
			filepaths = append(filepaths, currentPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return filepaths, nil
}

// ReadFilesWithTests returns all of the policies and test files
// at the given path including its subdirectories.
// Test files are Rego files that have a suffix of _test.rego
func ReadFilesWithTests(path string) ([]string, error) {
	files, err := getTestFiles(path)
	if err != nil {
		return nil, fmt.Errorf("search rego test files: %w", err)
	}

	return files, nil
}

func getTestFiles(path string) ([]string, error) {
	var filepaths []string
	err := filepath.Walk(path, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".rego") {
			filepaths = append(filepaths, currentPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return filepaths, nil
}
