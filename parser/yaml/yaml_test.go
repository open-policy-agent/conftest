package yaml_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/open-policy-agent/conftest/parser/yaml"
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
				name:           "empty config",
				controlConfigs: []byte(``),
				expectedResult: nil,
				shouldError:    false,
			},
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
			{
				name: "a single config with multiple yaml subdocs with crlf line endings",
				controlConfigs: []byte(strings.ReplaceAll(`---
sample: true
---
hello: true
---
nice: true`, "\n", "\r\n")),
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
			{
				name: "multiple documents with one invalid yaml",
				controlConfigs: []byte(`---
valid: true
---
invalid:
  - not closed
[
---
also_valid: true`),
				expectedResult: nil,
				shouldError:    true,
			},
			{
				name: "yaml with version directive",
				controlConfigs: []byte(`%YAML 1.1
---
group_id: 1234`),
				expectedResult: map[string]interface{}{
					"group_id": float64(1234),
				},
				shouldError: false,
			},
			{
				name: "yaml with version directive and multiple documents",
				controlConfigs: []byte(`%YAML 1.1
---
group_id: 1234
---
other_id: 5678
---
third_id: 9012`),
				expectedResult: []interface{}{
					map[string]interface{}{
						"group_id": float64(1234),
					},
					map[string]interface{}{
						"other_id": float64(5678),
					},
					map[string]interface{}{
						"third_id": float64(9012),
					},
				},
				shouldError: false,
			},
		}

		for _, test := range testTable {
			t.Run(test.name, func(t *testing.T) {
				var unmarshalledConfigs interface{}
				yamlParser := new(yaml.Parser)

				err := yamlParser.Unmarshal(test.controlConfigs, &unmarshalledConfigs)
				if test.shouldError && err == nil {
					t.Error("expected error but got none")
				} else if !test.shouldError && err != nil {
					t.Errorf("errors unmarshalling: %v", err)
				}

				if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
					t.Errorf("Expected\n%T : %v\n to equal\n%T : %v\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
				}
			})
		}
	})
}
