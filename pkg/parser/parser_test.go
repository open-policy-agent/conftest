package parser_test

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/instrumenta/conftest/pkg/parser"
	"github.com/instrumenta/conftest/pkg/parser/cue"
	"github.com/instrumenta/conftest/pkg/parser/ini"
	"github.com/instrumenta/conftest/pkg/parser/terraform"
	"github.com/instrumenta/conftest/pkg/parser/toml"
	"github.com/instrumenta/conftest/pkg/parser/yaml"
)

// array should be:
// a map [k,v] where k is the filename and v is the document

func TestUnmarshaller(t *testing.T) {
	t.Run("we should be able to construct a unmarshaller for a type of file", func(t *testing.T) {
		configManager := parser.NewConfigManager("yml")
		t.Run("which can be used to BulkUnmarshal file contents into an object", func(t *testing.T) {

			testTable := []struct {
				name           string
				controlReaders []parser.ConfigDoc
				expectedResult map[string]interface{}
				shouldError    bool
			}{
				{
					name: "a single reader",
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("sample: true")),
							Filepath:   "sample.yml",
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
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("sample: true")),
							Filepath:   "sample.yml",
						},
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("hello: true")),
							Filepath:   "hello.yml",
						},
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("nice: true")),
							Filepath:   "nice.yml",
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
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "sample.yml",
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
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "sample.yml",
						},
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "hello.yml",
						},
						parser.ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("nice: true")),
							Filepath:   "nice.yml",
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
					unmarshalledConfigs, err := configManager.BulkUnmarshal(test.controlReaders)
					if err != nil {
						t.Errorf("we should not have any errors on unmarshalling: %v", err)
					}

					if unmarshalledConfigs == nil {
						t.Error("we should see an actual value in our object, but we are nil")
					}

					if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
						t.Errorf("\nResult\n%v\n and type %T\n Expected\n%v\n and type %T\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
					}
				})
			}
		})
	})
}

func TestGetParser(t *testing.T) {
	testTable := []struct {
		name     string
		fileType string
		expected parser.Parser
	}{
		{
			name:     "Test getting Terraform parser from HCL input",
			fileType: "hcl",
			expected: new(terraform.Parser),
		},
		{
			name:     "Test getting Terraform parser from .tf input",
			fileType: "tf",
			expected: new(terraform.Parser),
		},
		{
			name:     "Test getting TOML parser",
			fileType: "toml",
			expected: new(toml.Parser),
		},
		{
			name:     "Test getting Cue parser",
			fileType: "cue",
			expected: new(cue.Parser),
		},
		{
			name:     "Test getting INI parser",
			fileType: "ini",
			expected: new(ini.Parser),
		},
		{
			name:     "Test getting YAML parser from JSON input",
			fileType: "json",
			expected: new(yaml.Parser),
		},
		{
			name:     "Test getting YAML parser from YAML input",
			fileType: "yaml",
			expected: new(yaml.Parser),
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			received := parser.GetParser(testUnit.fileType)

			if !reflect.DeepEqual(received, testUnit.expected) {
				t.Errorf("expected: %T \n got this: %T", testUnit.expected, received)
			}
		})
	}
}
