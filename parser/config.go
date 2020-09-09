package parser

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ParseConfigurations parses and returns the configurations given in the file list.
func ParseConfigurations(files []string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, "")
	if err != nil {
		return nil, fmt.Errorf("get configurations: %w", err)
	}

	return configurations, nil
}

// ParseConfigurationsAs parses the files as the given file type and returns the
// configurations given in the file list.
func ParseConfigurationsAs(files []string, fileType string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, fileType)
	if err != nil {
		return nil, fmt.Errorf("get configurations: %w", err)
	}

	return configurations, nil
}

// FormatAll takes in multiple configurations input and formats the configuration
// to be more human readable. The key of each configuration should be its filepath.
func FormatAll(configurations map[string]interface{}) (string, error) {
	output := "\n"
	for file, config := range configurations {
		output += file + "\n"

		current, err := Format(config)
		if err != nil {
			return "", fmt.Errorf("marshal output to json: %w", err)
		}

		output += current
	}

	return output, nil
}

// Format takes in a single configuration input and formats the configuration
// to be more human readable.
func Format(in interface{}) (string, error) {
	out, err := json.Marshal(in)
	if err != nil {
		return "", fmt.Errorf("marshal output to json: %w", err)
	}

	var prettyJSON bytes.Buffer
	if err = json.Indent(&prettyJSON, out, "", "\t"); err != nil {
		return "", fmt.Errorf("indentation: %w", err)
	}

	if _, err := prettyJSON.WriteString("\n"); err != nil {
		return "", fmt.Errorf("adding line break: %w", err)
	}

	return prettyJSON.String(), nil
}

func parseConfigurations(paths []string, fileType string) (map[string]interface{}, error) {
	var parsedConfigurations []map[string]interface{}
	for _, path := range paths {
		contents, err := getConfigurationContent(path)
		if err != nil {
			return nil, fmt.Errorf("get configuration content: %w", err)
		}

		parsedConfiguration, err := parseConfiguration(path, contents, fileType)
		if err != nil {
			return nil, fmt.Errorf("parsing configuration: %w", err)
		}

		parsedConfigurations = append(parsedConfigurations, parsedConfiguration)
	}

	result := make(map[string]interface{})
	for _, config := range parsedConfigurations {
		for path, contents := range config {
			result[path] = contents
		}
	}

	return result, nil
}

func parseConfiguration(path string, configuration []byte, fileType string) (map[string]interface{}, error) {
	var parser Parser
	var err error
	if fileType == "" {
		parser, err = GetParserFromPath(path)
	} else {
		parser, err = GetParser(fileType)
	}
	if err != nil {
		return nil, fmt.Errorf("get parser: %w", err)
	}

	var parsed interface{}
	if err := parser.Unmarshal(configuration, &parsed); err != nil {
		return nil, fmt.Errorf("parser unmarshal: %w", err)
	}

	parsedConfiguration := make(map[string]interface{})
	parsedConfiguration[path] = parsed

	return parsedConfiguration, nil
}

func getConfigurationContent(path string) ([]byte, error) {
	if path == "-" {
		contents, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			return nil, fmt.Errorf("read standard in: %w", err)
		}

		return contents, nil
	}

	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get abs: %w", err)
	}

	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return contents, nil
}
