package yaml

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/ghodss/yaml"
)

type Parser struct{}

//Format returns the expected format of the input to be parsed
func (yp *Parser) Format() string {
	return "yaml"
}

func (yp *Parser) separateSubDocuments(data []byte) [][]byte {
	linebreak := "\n"
	windowsLineEnding := bytes.Contains(data, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		linebreak = "\r\n"
	}
	return bytes.Split(data, []byte(linebreak+"---"+linebreak))
}

func (yp *Parser) unmarshalMultipleDocuments(subDocuments [][]byte, v interface{}) error {
	var documentStore []interface{}
	for _, subDocument := range subDocuments {
		var documentObject interface{}
		err := yaml.Unmarshal(subDocument, &documentObject)
		if err != nil {
			return fmt.Errorf("Unable to parse YAML: %s", err)
		}
		documentStore = append(documentStore, documentObject)
	}

	yamlConfigBytes, err := yaml.Marshal(documentStore)
	if err != nil {
		return fmt.Errorf("Unable to marshal documentStore %v: %s", documentStore, err)
	}
	err = yaml.Unmarshal(yamlConfigBytes, v)
	if err != nil {
		return fmt.Errorf("Unable to Unmarshal yamlConfigBytes %s: %s", string(yamlConfigBytes), err)
	}
	return nil
}

func (yp *Parser) Unmarshal(p []byte, v interface{}) error {
	subDocuments := yp.separateSubDocuments(p)
	if len(subDocuments) > 1 {
		return yp.unmarshalMultipleDocuments(subDocuments, v)
	}

	err := yaml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to Unmarshal yamlConfigBytes %s: %s", string(p), err)
	}
	return nil
}
