package edn_test

import (
	"reflect"
	"testing"

	"github.com/instrumenta/conftest/parser/edn"
)

func TestEDNParser(t *testing.T) {
	t.Run("we should be able to parse an EDN document", func(t *testing.T) {

		testTable := []struct {
			name           string
			controlConfigs []byte
			expectedResult interface{}
			shouldError    bool
		}{
			{
				name:           "a single config",
				controlConfigs: []byte(`{:sample true}`),
				expectedResult: map[string]interface{}{
					":sample": "true",
				},
				shouldError: false,
			},
			{
				name: "a basic edn file with a sample of types",
				controlConfigs: []byte(`{;; This is a comment and should be ignored by the parser
:sample1 "my-username",
:sample2 false,
:sample3 5432}`),
				expectedResult: map[string]interface{}{
					":sample1": "my-username",
					":sample2": "false",
					":sample3": "5432",
				},
				shouldError: false,
			},
		}

		for _, test := range testTable {
			t.Run(test.name, func(t *testing.T) {
				var unmarshalledConfigs interface{}
				ednParser := new(edn.Parser)

				err := ednParser.Unmarshal(test.controlConfigs, &unmarshalledConfigs)

				if err != nil {
					t.Errorf("we should not have any errors on unmarshalling: %v", err)
				}

				if unmarshalledConfigs == nil {
					t.Error("we should see an actual value in our object, but we are nil")
				}

				if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
					t.Errorf("Expected\n%T : %v\n to equal\n%T : %v\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
				}
			})
		}
	})
}
