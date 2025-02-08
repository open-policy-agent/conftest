package parser

import (
	"encoding/json"
	"fmt"
)

// Format takes in multiple configurations input and formats the configuration
// to be more human readable. The key of each configuration should be its filepath.
func Format(configurations map[string]any) (string, error) {
	var output string
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

// FormatJSON takes in multiple configurations and formats them as a JSON
// object where each key is the path to the file and the contents are the
// parsed configurations.
func FormatJSON(configurations map[string]any) (string, error) {
	marshaled, err := json.MarshalIndent(configurations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal configs: %w", err)
	}

	return string(marshaled), nil
}

// FormatCombined takes in multiple configurations, combines them, and formats the
// configuration to be more human readable. The key of each configuration should be
// its filepath.
func FormatCombined(configurations map[string]any) (string, error) {
	combinedConfigurations := CombineConfigurations(configurations)

	formattedConfigs, err := format(combinedConfigurations["Combined"])
	if err != nil {
		return "", fmt.Errorf("formatting configs: %w", err)
	}

	return formattedConfigs, nil
}

func format(configs any) (string, error) {
	out, err := json.MarshalIndent(configs, "", "\t")
	if err != nil {
		return "", fmt.Errorf("marshal output to json: %w", err)
	}

	return string(out) + "\n", nil
}
