package yaml

import (
	"bytes"
	"fmt"

	"sigs.k8s.io/yaml"
)

// Parser is a YAML parser.
type Parser struct{}

// Unmarshal unmarshals YAML files.
func (yp *Parser) Unmarshal(p []byte, v interface{}) error {
	subDocuments := separateSubDocuments(p)
	if len(subDocuments) > 1 {
		if err := unmarshalMultipleDocuments(subDocuments, v); err != nil {
			return fmt.Errorf("unmarshal multiple documents: %w", err)
		}

		return nil
	}

	if err := yaml.Unmarshal(p, v); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}

	return nil
}

func separateSubDocuments(data []byte) [][]byte {
	linebreak := "\n"
	if bytes.Contains(data, []byte("\r\n---\r\n")) {
		linebreak = "\r\n"
	}

	return bytes.Split(data, []byte(linebreak+"---"+linebreak))
}

func unmarshalMultipleDocuments(subDocuments [][]byte, v interface{}) error {
	var documentStore []interface{}
	for _, subDocument := range subDocuments {
		var documentObject interface{}
		if err := yaml.Unmarshal(subDocument, &documentObject); err != nil {
			return fmt.Errorf("unmarshal subdocument yaml: %w", err)
		}

		documentStore = append(documentStore, documentObject)
	}

	yamlConfigBytes, err := yaml.Marshal(documentStore)
	if err != nil {
		return fmt.Errorf("marshal yaml document: %w", err)
	}

	if err := yaml.Unmarshal(yamlConfigBytes, v); err != nil {
		return fmt.Errorf("unmarshal yaml: %w", err)
	}

	return nil
}
