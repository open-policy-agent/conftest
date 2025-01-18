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
	// Determine line ending style
	linebreak := "\n"
	if bytes.Contains(data, []byte("\r\n")) {
		linebreak = "\r\n"
	}
	separator := fmt.Sprintf("%s---%s", linebreak, linebreak)

	// Count actual document separators
	parts := bytes.Split(data, []byte(separator))

	// If we have a directive, first part is not a separate document
	if bytes.HasPrefix(data, []byte("%")) {
		if len(parts) <= 2 {
			// Single document with directive
			return [][]byte{data}
		}
		// Multiple documents - combine directive with first real document
		firstDoc := append(parts[0], append([]byte(separator), parts[1]...)...)
		result := [][]byte{firstDoc}
		result = append(result, parts[2:]...)
		return result
	}

	// No directive case
	if len(parts) <= 1 {
		// Single document
		return [][]byte{data}
	}
	return parts
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
