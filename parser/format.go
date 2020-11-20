package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Format takes in multiple configurations input and formats the configuration
// to be more human readable. The key of each configuration should be its filepath.
func Format(configurations map[string]interface{}) (string, error) {
	output := "\n"
	for file, config := range configurations {
		output += file + "\n"

		current, err := format(config)
		if err != nil {
			return "", fmt.Errorf("marshal output to json: %w", err)
		}

		output += current
	}

	return output, nil
}

// FormatCombined takes in multiple configurations, combines them, and formats the
// configuration to be more human readable. The key of each configuration should be
// its filepath.
func FormatCombined(configurations map[string]interface{}) (string, error) {
	combinedConfigurations := CombineConfigurations(configurations)

	formattedConfigs, err := format(combinedConfigurations["Combined"])
	if err != nil {
		return "", fmt.Errorf("formatting configs: %w", err)
	}

	return formattedConfigs, nil
}

func format(configs interface{}) (string, error) {
	out, err := json.Marshal(configs)
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
