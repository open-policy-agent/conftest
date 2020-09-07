package output

import (
	"reflect"
	"testing"
)

func TestSupportedOutputManagers(t *testing.T) {
	for _, testunit := range []struct {
		name          string
		outputFormat  string
		outputManager OutputManager
	}{
		{
			name:          "std output should exist",
			outputFormat:  outputSTD,
			outputManager: NewDefaultStandardOutputManager(true),
		},
		{
			name:          "json output should exist",
			outputFormat:  outputJSON,
			outputManager: NewDefaultJSONOutputManager(),
		},
		{
			name:          "tap output should exist",
			outputFormat:  outputTAP,
			outputManager: NewDefaultTAPOutputManager(),
		},
		{
			name:          "table output should exist",
			outputFormat:  outputTable,
			outputManager: NewDefaultTableOutputManager(),
		},
		{
			name:          "JUnit should exist",
			outputFormat:  outputJUnit,
			outputManager: NewDefaultJUnitOutputManager(),
		},
		{
			name:          "default output should exist",
			outputFormat:  "somedefault",
			outputManager: NewDefaultStandardOutputManager(true),
		},
	} {
		outputManager := GetOutputManager(testunit.outputFormat, true)
		if !reflect.DeepEqual(outputManager, testunit.outputManager) {
			t.Errorf(
				"We expected the output manager to be of type %v : %T and it was %T",
				testunit.outputFormat,
				testunit.outputManager,
				outputManager,
			)
		}

	}
}
