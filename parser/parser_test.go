package parser

import (
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/instrumenta/conftest/parser/cue"
	"github.com/instrumenta/conftest/parser/edn"
	"github.com/instrumenta/conftest/parser/hcl2"
	"github.com/instrumenta/conftest/parser/ini"
	"github.com/instrumenta/conftest/parser/terraform"
	"github.com/instrumenta/conftest/parser/toml"
	"github.com/instrumenta/conftest/parser/yaml"
)

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
						ConfigDoc{
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
						ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("sample: true")),
							Filepath:   "sample.yml",
							Parser:     &yaml.Parser{},
						},
						ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader("hello: true")),
							Filepath:   "hello.yml",
							Parser:     &yaml.Parser{},
						},
						ConfigDoc{
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
						ConfigDoc{
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
						ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "sample.yml",
							Parser:   &yaml.Parser{},
						},
						ConfigDoc{
							ReadCloser: ioutil.NopCloser(strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`)),
							Filepath: "hello.yml",
							Parser:   &yaml.Parser{},
						},
						ConfigDoc{
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

func TestGetParser(t *testing.T) {
	testTable := []struct {
		name        string
		fileType    string
		expected    Parser
		expectError bool
	}{
		{
			name:        "Test getting Terraform parser from HCL1 input",
			fileType:    "hcl1",
			expected:    new(terraform.Parser),
			expectError: false,
		},
		{
			name:        "Test getting HCL2 parser from .tf input",
			fileType:    "tf",
			expected:    new(hcl2.Parser),
			expectError: false,
		},
		{
			name:        "Test getting TOML parser",
			fileType:    "toml",
			expected:    new(toml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting Cue parser",
			fileType:    "cue",
			expected:    new(cue.Parser),
			expectError: false,
		},
		{
			name:        "Test getting INI parser",
			fileType:    "ini",
			expected:    new(ini.Parser),
			expectError: false,
		},
		{
			name:        "Test getting YAML parser from JSON input",
			fileType:    "json",
			expected:    new(yaml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting YAML parser from YAML input",
			fileType:    "yaml",
			expected:    new(yaml.Parser),
			expectError: false,
		},
		{
			name:        "Test getting EDN parser",
			fileType:    "edn",
			expected:    new(edn.Parser),
			expectError: false,
		},
		{
			name:        "Test getting invalid filetype",
			fileType:    "epicfailure",
			expected:    nil,
			expectError: true,
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			received, err := GetParser(testUnit.fileType)

			if !reflect.DeepEqual(received, testUnit.expected) {
				t.Errorf("expected: %T \n got this: %T", testUnit.expected, received)
			}
			if !testUnit.expectError && err != nil {
				t.Errorf("error here: %v", err)
			}
			if testUnit.expectError && err == nil {
				t.Error("error expected but not received")
			}
		})
	}
}
