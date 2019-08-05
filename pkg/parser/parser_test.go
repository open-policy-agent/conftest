package parser_test

import (
	"io"
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
							Filepath: "test.yml",
						},
					},
					expectedResult: map[string]interface{}{
						"test.yml": []interface{}{
							map[string]interface{}{
								"sample": true,
							},
						},
					},
					shouldError: false,
				},
				{
					name: "multiple readers",
					controlReaders: []io.Reader{
						strings.NewReader("sample: true"),
						strings.NewReader("hello: true"),
						strings.NewReader("nice: true"),
					},
					expectedResult: []interface{}{
						[]interface{}{
							map[string]interface{}{
								"sample": true,
							},
						},
						[]interface{}{
							map[string]interface{}{
								"hello": true,
							},
						},
						[]interface{}{
							map[string]interface{}{
								"nice": true,
							},
						},
					},
					shouldError: false,
				},
				{
					name: "a single reader with multiple yaml subdocs",
					controlReaders: []io.Reader{
						strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`),
					},
					expectedResult: []interface{}{
						[]interface{}{
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
					controlReaders: []io.Reader{
						strings.NewReader(`---
sample: true
---
hello: true
---
nice: true`),
						strings.NewReader(`sunny: true`),
						strings.NewReader(`fun: true
---
date: false
---
ilk: true`),
					},
					expectedResult: []interface{}{
						[]interface{}{
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
						[]interface{}{
							map[string]interface{}{
								"sunny": true,
							},
						},
						[]interface{}{
							map[string]interface{}{
								"fun": true,
							},
							map[string]interface{}{
								"date": false,
							},
							map[string]interface{}{
								"ilk": true,
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
					case []interface{}:
					default:
						t.Errorf("Expected []map[string]interface{} but instead got %T\n%v", v, unmarshalledConfigs)
					}

					if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
						t.Errorf("Result\n%v\n and type %T\n Expected\n%v\n and type %T\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
					}
				})
			}
		})
	})
}
