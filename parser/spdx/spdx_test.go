package spdx

import (
	"bytes"
	"testing"
)

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

	input, err := parser.Parse(bytes.NewBufferString(p))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	currentDataLicense := inputMap["dataLicense"]
	expectedDataLicense := "conftest-demo"
	if currentDataLicense != expectedDataLicense {
		t.Errorf("DataLicense of the SPDX file have: %s, want: %s", currentDataLicense, expectedDataLicense)
	}
}
