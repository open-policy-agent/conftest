package parser

import (
	"context"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/open-policy-agent/conftest/parser/yaml"
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
				"../examples/hcl1/gke.tf",
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
		fileName         string
		expectedFileType string
	}{
		{"Test YAML file", "", "example/kubernetes/deployment.yaml", "yaml"},
		{"Test not YAML file", "", "example/traefik/traefik.toml", "toml"},
		{"Test default file type", "", "-", "yaml"},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			fileType := getFileType(testUnit.fileName, testUnit.inputFileType)
			if fileType != testUnit.expectedFileType {
				t.Fatalf("got wrong filetype got:%s want:%s", fileType, testUnit.expectedFileType)
			}
		})
	}
}

func TestUnmarshaller(t *testing.T) {
	t.Run("error constructing an unmarshaller for a type of file", func(t *testing.T) {
		t.Run("which can be used to BulkUnmarshal file contents into an object", func(t *testing.T) {
			testTable := []struct {
				name           string
				controlReaders []ConfigDoc
				expectedResult map[string]interface{}
				shouldError    bool
			}{
				{
					name: "a single reader",
					controlReaders: []ConfigDoc{
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader("sample: true")),
							Filepath:   "sample.yml",
							Parser:     &yaml.Parser{},
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": map[string]interface{}{
							"sample": true,
						},
					},
					shouldError: false,
				},
				{
					name: "multiple readers",
					controlReaders: []ConfigDoc{
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader("sample: true")),
							Filepath:   "sample.yml",
							Parser:     &yaml.Parser{},
						},
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader("hello: true")),
							Filepath:   "hello.yml",
							Parser:     &yaml.Parser{},
						},
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader("nice: true")),
							Filepath:   "nice.yml",
							Parser:     &yaml.Parser{},
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": map[string]interface{}{
							"sample": true,
						},
						"hello.yml": map[string]interface{}{
							"hello": true,
						},
						"nice.yml": map[string]interface{}{
							"nice": true,
						},
					},
					shouldError: false,
				},
				{
					name: "a single reader with multiple yaml subdocs",
					controlReaders: []ConfigDoc{
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "sample.yml",
							Parser:   &yaml.Parser{},
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
							map[string]interface{}{
								"hello": true,
							},
							map[string]interface{}{
								"nice": true,
							},
						},
					},
					shouldError: false,
				},
				{
					name: "multiple readers with multiple subdocs",
					controlReaders: []ConfigDoc{
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "sample.yml",
							Parser:   &yaml.Parser{},
						},
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "hello.yml",
							Parser:   &yaml.Parser{},
						},
						{
							ReadCloser: ioutil.NopCloser(strings.NewReader("nice: true")),
							Filepath:   "nice.yml",
							Parser:     &yaml.Parser{},
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
							map[string]interface{}{
								"hello": true,
							},
							map[string]interface{}{
								"nice": true,
							},
						},
						"hello.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
							map[string]interface{}{
								"hello": true,
							},
							map[string]interface{}{
								"nice": true,
							},
						},
						"nice.yml": map[string]interface{}{
							"nice": true,
						},
					},
					shouldError: false,
				},
			}

			for _, test := range testTable {
				t.Run(test.name, func(t *testing.T) {
					var unmarshalledConfigs map[string]interface{}
					unmarshalledConfigs, err := bulkUnmarshal(test.controlReaders)
					if err != nil {
						t.Errorf("errors unmarshalling: %v", err)
					}

					if unmarshalledConfigs == nil {
						t.Error("error seeing the actual value of object, received nil")
					}

					if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
						t.Errorf("\nResult\n%v\n and type %T\n Expected\n%v\n and type %T\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
					}
				})
			}
		})
	})
}
