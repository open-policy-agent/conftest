package yaml_test

import (
	"reflect"
	"testing"

	"github.com/instrumenta/conftest/parser/yaml"
)

func TestYAMLParser(t *testing.T) {
	t.Run("error parsing a YAML document", func(t *testing.T) {

		testTable := []struct {
			name           string
			controlConfigs []byte
			expectedResult interface{}
			shouldError    bool
		}{
			{
				name:           "a single config",
				controlConfigs: []byte(`sample: true`),
				expectedResult: map[string]interface{}{
					"sample": true,
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
				expectedResult: []interface{}{
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
				var unmarshalledConfigs interface{}
				yamlParser := new(yaml.Parser)

				if err := yamlParser.Unmarshal(test.controlConfigs, &unmarshalledConfigs); err != nil {
					t.Errorf("errors unmarshalling: %v", err)
				}

				if unmarshalledConfigs == nil {
					t.Error("error seeing actual value in object, received nil")
				}

				if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
					t.Errorf("Expected\n%T : %v\n to equal\n%T : %v\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
				}
			})
		}
	})
}
