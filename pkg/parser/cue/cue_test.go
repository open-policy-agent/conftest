package cue

import (
	"testing"
)

func TestCueParser(t *testing.T) {
	parser := &Parser{}
	p := `package kubernetes


	deployment "hello-kubernetes": {
		apiVersion: "apps/v1"
		spec: {
			replicas: 3
			template spec containers: [{
				image: "paulbouwer/hello-kubernetes:1.5"
				ports: [{
					containerPort: 8080
				}]
			}]
		}
	}`
	var input interface{}
	err := parser.Unmarshal([]byte(p), &input)
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input.(map[string]interface{})
	item := inputMap["deployment"]
	if len(item.(map[string]interface{})) <= 0 {
		t.Error("There should be at least one item defined in the parsed file, but none found")
	}
}
