package edn_test

import (
	"reflect"
	"testing"

	"github.com/open-policy-agent/conftest/parser/edn"
)

func TestEDNParser(t *testing.T) {
	testTable := []struct {
		name           string
		controlConfigs []byte
		expectedResult any
	}{
		{
			name:           "a single config",
			controlConfigs: []byte(`{:sample true}`),
			expectedResult: map[string]any{
				":sample": "true",
			},
		},
		{
			name: "a basic edn file with a sample of types",
			controlConfigs: []byte(`{;; This is a comment and should be ignored by the parser
:sample1 "my-username",
:sample2 false,
:sample3 5432}`),
			expectedResult: map[string]any{
				":sample1": "my-username",
				":sample2": "false",
				":sample3": "5432",
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			var unmarshalledConfigs any
			ednParser := new(edn.Parser)

			if err := ednParser.Unmarshal(test.controlConfigs, &unmarshalledConfigs); err != nil {
				t.Errorf("err on unmarshalling: %v", err)
			}

			if unmarshalledConfigs == nil {
				t.Error("expected actual value in our object, got nil")
			}

			if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
				t.Errorf("expected\n%T : %v\n to equal\n%T : %v\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
			}
		})
	}
}
