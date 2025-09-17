package cue

import (
	"bytes"
	"testing"
)

func TestCueParser(t *testing.T) {
	p := `apiVersion: "apps/v1"
	kind:       "Deployment"
	metadata: name: "hello-kubernetes"
	spec: {
		replicas: 3
		selector: matchLabels: app: "hello-kubernetes"
		template: {
			metadata: labels: app: "hello-kubernetes"
			spec: containers: [{
				name:  "hello-kubernetes"
				image: "paulbouwer/hello-kubernetes:1.5"
				ports: [{
					containerPort: 8080
				}]
			}]
		}
	}`

	parser := &Parser{}

	input, err := parser.Parse(bytes.NewReader([]byte(p)))
	if err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if len(input) != 1 {
		t.Error("There should be information parsed but its nil")
	}

	inputMap := input[0].(map[string]any)
	kind := inputMap["kind"]
	if kind != "Deployment" {
		t.Error("Parsed cuelang file should be a deployment, but was not")
	}
}
