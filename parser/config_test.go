package parser

import (
	"context"
	"testing"
)

func TestGetConfigurations(t *testing.T) {
	testTabel := []struct {
		name     string
		fileList []string
	}{
		{
			"Test single type",
			[]string{"../examples/kubernetes/service.yaml"},
		},
		{
			"Test two input with same type",
			[]string{"../examples/kubernetes/service.yaml", "../examples/kubernetes/deployment.yaml"},
		},
		{
			"Test different types",
			[]string{
				"../examples/kubernetes/service.yaml",
				"../examples/traefik/traefik.toml",
				"../examples/terraform/gke.tf",
				"../examples/edn/sample_config.edn",
			},
		},
	}

	for _, testUnit := range testTabel {
		t.Run(testUnit.name, func(t *testing.T) {
			ctx := context.Background()
			_, err := GetConfigurations(ctx, "", testUnit.fileList)
			if err != nil {
				t.Fatalf("error while getting configurations: %v", err)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	testTable := []struct {
		name             string
		inputFileType    string
		filename         string
		expectedFileType string
	}{
		{"Test YAML file", "", "example/kubernetes/deployment.yaml", "yaml"},
		{"Test not YAML file", "", "example/traefik/traefik.toml", "toml"},
		{"Test default file type", "", "-", "yaml"},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			fileType, err := getFileType(testUnit.inputFileType, testUnit.filename)
			if err != nil {
				t.Fatalf("errors getting filetype: %v", err)
			}

			if fileType != testUnit.expectedFileType {
				t.Fatalf("got wrong filetype got:%s want:%s", fileType, testUnit.expectedFileType)
			}
		})
	}
}
