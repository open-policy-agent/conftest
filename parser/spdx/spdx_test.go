package spdx

import "testing"

func TestSPDXParser(t *testing.T) {
	p := `SPDXVersion: SPDX-2.2
DataLicense: conftest-demo
SPDXID: SPDXRef-DOCUMENT
DocumentName: hello
DocumentNamespace: https://swinslow.net/spdx-examples/example1/hello-v3
Creator: Person: Steve Winslow (steve@swinslow.net)
Creator: Tool: github.com/spdx/tools-golang/builder
Creator: Tool: github.com/spdx/tools-golang/idsearcher
Created: 2021-08-26T01:46:00Z
`

	parser := &Parser{}

	var input any
	if err := parser.Unmarshal([]byte(p), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input.(map[string]any)
	currentDataLicense := inputMap["dataLicense"]
	expectedDataLicense := "conftest-demo"
	if currentDataLicense != expectedDataLicense {
		t.Errorf("DataLicense of the SPDX file have: %s, want: %s", currentDataLicense, expectedDataLicense)
	}
}
