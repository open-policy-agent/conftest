package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/parser"
)

type ParseRunner struct {
	Input string
	Combine bool
}

func (r *ParseRunner) Run(ctx context.Context, fileList []string) (string, error) {
	return parseInput(ctx, r.Input,r.Combine, fileList)
}

func parseInput(ctx context.Context, input string, combine bool, fileList []string) (string, error) {
	configurations, err := parser.GetConfigurations(ctx, input, fileList)
	if err != nil {
		return "", fmt.Errorf("calling the parser method: %w", err)
	}

	parsedConfigurations, err := parseConfigurations(configurations, combine)
	if err != nil {
		return "", fmt.Errorf("parsing configs: %w", err)
	}

	return parsedConfigurations, nil

}

func parseConfigurations(configurations map[string]interface{}, combine bool) (string, error) {
	var output string
	if combine {
		content, err := marshal(configurations)
		if err != nil {
			return "", fmt.Errorf("marshal output to json: %w", err)
		}

		output = strings.Replace(output+"\n"+content, "\\r", "", -1)
	} else {
		for filename, config := range configurations {
			content, err := marshal(config)
			if err != nil {
				return "", fmt.Errorf("marshal output to json: %w", err)
			}

			output = strings.Replace(output+filename+"\n"+content, "\\r", "", -1)
		}
	}

	return output, nil
}

func marshal(in interface{}) (string, error) {
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
