package yaml_test

import (
	"reflect"
	"testing"

	"github.com/instrumenta/conftest/pkg/parser/yaml"
)

func TestYAMLParser(t *testing.T) {
	t.Run("we should be able to parse a YAML document", func(t *testing.T) {

		testTable := []struct {
			name           string
			controlConfigs []byte
			expectedResult interface{}
			shouldError    bool
		}{
			{
				name:           "a single config",
				controlConfigs: []byte(`sample: true`),
				expectedResult: []map[string]interface{}{
					{"sample": true},
				},
				shouldError: false,
			},
			{
				name: "a single config with multiple yaml subdocs",
				controlConfigs: []byte(`---
sample: true
---
hello: true
---
nice: true`),
				expectedResult: []map[string]interface{}{
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
				shouldError: false,
			},
		}

		for _, test := range testTable {
			t.Run(test.name, func(t *testing.T) {
				var unmarshalledConfigs []map[string]interface{}
				yamlParser := new(yaml.Parser)

				err := yamlParser.Unmarshal(test.controlConfigs, &unmarshalledConfigs)

				if err != nil {
					t.Errorf("we should not have any errors on unmarshalling: %v", err)
				}

				if unmarshalledConfigs == nil {
					t.Error("we should see an actual value in our object, but we are nil")
				}

				if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
					t.Errorf("Expected\n%T\n to equal\n%T\n", unmarshalledConfigs, test.expectedResult)
				}
			})
		}
	})
}
