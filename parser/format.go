package parser

import (
	"encoding/json"
	"fmt"
)

// Format takes in multiple configurations and formats them as a JSON
// object where each key is the path to the file and the contents are the
// parsed configurations.
// If there's only one entry in the given configurations, omits the
func Format(configurations map[string][]any) (string, error) {
	var v any
	if len(configurations) == 1 {
		for _, singleConfig := range configurations {
			v = flattenConfigs(singleConfig)
		}
	} else {
		flattenedConfigurations := map[string]any{}
		for path, subconfigs := range configurations {
			flattenedConfigurations[path] = flattenConfigs(subconfigs)
		}
		v = flattenedConfigurations
	}
	return marshalJSON(v)
}

// flattenConfigs maintain backwards compatibility of formatted JSON output
// for parsers which may return multiple documents (for example, YAML parser)
func flattenConfigs(configs []any) any {
	if len(configs) == 1 {
		return configs[0]
	}
	return configs
}

// FormatCombined takes in multiple configurations, combines them, and formats the
// configuration to be more human-readable. The key of each configuration should be
// its filepath.
func FormatCombined(configurations map[string][]any) (string, error) {
	return marshalJSON(CombineConfigurations(configurations)["Combined"])
}

func marshalJSON(v any) (string, error) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("formatting config to json: %w", err)
	}
	return string(out), nil
}
