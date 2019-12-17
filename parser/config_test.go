package parser

import (
	"testing"
)

func TestGetFileType(t *testing.T) {
	testTable := []struct {
		name             string
		inputFileType    string
		filename         string
		expectedFileType string
		shouldError      bool
	}{
		{"Test YAML file", "", "example/kubernetes/deployment.yaml", "yaml", false},
		{"Test not YAML file", "", "example/traefik/traefik.toml", "toml", false},
		{"Test default file type", "", "-", "yaml", false},
		{"Test unsupported file type", "", "example/filetype.invalid", "", true},
		{"Test unsupperted input file type", "invalid", "example/traefik/traefik.toml", "", true},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			fileType, err := getFileType(testUnit.inputFileType, testUnit.filename)
			if err != nil {
				if testUnit.shouldError {
					return
				}

				t.Fatalf("errors getting filetype: %v", err)
			}

			if testUnit.shouldError {
				t.Fatalf("expected errors but returned no error")
			}

			if fileType != testUnit.expectedFileType {
				t.Fatalf("got wrong filetype got:%s want:%s", fileType, testUnit.expectedFileType)
			}
		})
	}
}
