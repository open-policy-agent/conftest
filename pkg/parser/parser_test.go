package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/instrumenta/conftest/pkg/parser"
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
							Reader:   strings.NewReader("sample: true"),
							Filepath: "sample.yml",
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
						},
					},
					shouldError: false,
				},
				{
					name: "multiple readers",
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							Reader:   strings.NewReader("sample: true"),
							Filepath: "sample.yml",
						},
						parser.ConfigDoc{
							Reader:   strings.NewReader("hello: true"),
							Filepath: "hello.yml",
						},
						parser.ConfigDoc{
							Reader:   strings.NewReader("nice: true"),
							Filepath: "nice.yml",
						},
					},
					expectedResult: map[string]interface{}{
						"sample.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
						},
						"hello.yml": []interface{}{
							map[string]interface{}{
								"hello": true,
							},
						},
						"nice.yml": []interface{}{
							map[string]interface{}{
								"nice": true,
							},
						},
					},
					shouldError: false,
				},
				{
					name: "a single reader with multiple yaml subdocs",
					controlReaders: []parser.ConfigDoc{
						parser.ConfigDoc{
							Reader: strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`),
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
							Reader: strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`),
							Filepath: "sample.yml",
						},
						parser.ConfigDoc{
							Reader: strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`),
							Filepath: "hello.yml",
						},
						parser.ConfigDoc{
							Reader:   strings.NewReader("nice: true"),
							Filepath: "nice.yml",
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
						"nice.yml": []interface{}{
							map[string]interface{}{
								"nice": true,
							},
						},
					},
					shouldError: false,
				},
			}

			for _, test := range testTable {
				t.Run(test.name, func(t *testing.T) {
					var unmarshalledConfigs interface{}
					unmarshalledConfigs, err := configManager.BulkUnmarshal(test.controlReaders)
					if err != nil {
						t.Errorf("we should not have any errors on unmarshalling: %v", err)
					}

					if unmarshalledConfigs == nil {
						t.Error("we should see an actual value in our object, but we are nil")
					}

					switch v := unmarshalledConfigs.(type) {
					case map[string]interface{}:
					default:
						t.Errorf("Expected %T but instead got %T", test.expectedResult, v)
					}

					if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
						t.Errorf("\nResult\n%v\n and type %T\n Expected\n%v\n and type %T\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
					}
				})
			}
		})
	})
}
