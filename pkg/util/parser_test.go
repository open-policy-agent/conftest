package util

import (
	"testing"
)

func TestGetParser(t *testing.T) {

	t.Run("GetParser on a valid terraform hcl file", func(t *testing.T) {
		testParser := GetParser("testdata/sample.tf")

		t.Run("should parse the hcl into a input object without error", func(t *testing.T) {
			var input interface{}
			err := testParser.Unmarshal(nil, &input)
			if err != nil {
				t.Errorf("parser should not have thrown an error: %v", err)
			}

			if input == nil {
				t.Error("there should be information parsed but its nil")
			}

			inputMap := input.(map[string]interface{})
			if len(inputMap["Resources"].([]interface{})) <= 0 {
				t.Error("there should be resources defined in the parsed file, but none found")
			}
		})
	})
}
