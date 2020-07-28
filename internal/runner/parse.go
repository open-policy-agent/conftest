package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/parser"
)

type ParseParams struct {
	Input   string
	Combine bool
}

type ParseRunner struct {
	Params        *ParseParams
	ConfigManager parser.CustomConfigManager
}

func (r *ParseRunner) Run(ctx context.Context, fileList []string) (string, error) {
	configurations, err := r.ConfigManager.GetConfigurations(ctx, r.Params.Input, fileList)
	if err != nil {
		return "", fmt.Errorf("calling the parser method: %w", err)
	}

	parsedConfigurations, err := r.parseConfigurations(configurations)
	if err != nil {
		return "", fmt.Errorf("parsing configs: %w", err)
	}

	return parsedConfigurations, nil
}

func (r *ParseRunner) parseConfigurations(configurations map[string]interface{}) (string, error) {
	var output string
	if r.Params.Combine {
		content, err := r.marshal(configurations)
		if err != nil {
			return "", fmt.Errorf("marshal output to json: %w", err)
		}

		output = strings.Replace(output+"\n"+content, "\\r", "", -1)
	} else {
		for filename, config := range configurations {
			content, err := r.marshal(config)
			if err != nil {
				return "", fmt.Errorf("marshal output to json: %w", err)
			}

			output = strings.Replace(output+filename+"\n"+content, "\\r", "", -1)
		}
	}

	return output, nil
}

func (r *ParseRunner) marshal(in interface{}) (string, error) {
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
