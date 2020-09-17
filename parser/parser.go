package parser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/conftest/parser/jsonnet"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/edn"
	"github.com/open-policy-agent/conftest/parser/hcl1"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/hocon"
	"github.com/open-policy-agent/conftest/parser/ini"
	"github.com/open-policy-agent/conftest/parser/json"
	"github.com/open-policy-agent/conftest/parser/toml"
	"github.com/open-policy-agent/conftest/parser/vcl"
	"github.com/open-policy-agent/conftest/parser/xml"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

// ValidInputs returns string array in order to passing valid input types to viper
func ValidInputs() []string {
	return []string{
		"toml",
		"tf",
		"hcl",
		"hcl1",
		"cue",
		"ini",
		"yml",
		"yaml",
		"json",
		"jsonnet",
		"Dockerfile",
		"edn",
		"vcl",
		"xml",
	}
}

// Parser defines all of the methods that every parser definition
// must implement.
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// GetParser returns a specific file parser based on the given
// file extension. The input should be the file extension without the
// period.
func GetParser(fileExtension string) (Parser, error) {
	fileExtension = strings.ToLower(fileExtension)

	switch fileExtension {
	case "toml":
		return &toml.Parser{}, nil
	case "hcl1":
		return &hcl1.Parser{}, nil
	case "cue":
		return &cue.Parser{}, nil
	case "ini":
		return &ini.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "hcl", "tf", "hcl2":
		return &hcl2.Parser{}, nil
	case "dockerfile":
		return &docker.Parser{}, nil
	case "yml", "yaml":
		return &yaml.Parser{}, nil
	case "json":
		return &json.Parser{}, nil
	case "jsonnet":
		return &jsonnet.Parser{}, nil
	case "edn":
		return &edn.Parser{}, nil
	case "vcl":
		return &vcl.Parser{}, nil
	case "xml":
		return &xml.Parser{}, nil
	default:
		return nil, fmt.Errorf("unknown file extension given: %v", fileExtension)
	}
}

// GetParserFromPath returns a file parser based on the file type
// that exists at the given path.
func GetParserFromPath(path string) (Parser, error) {
	fileType := getFileType(path)

	return GetParser(fileType)
}

// ParseConfigurations parses and returns the configurations from the given
// list of files.
func ParseConfigurations(files []string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, "")
	if err != nil {
		return nil, fmt.Errorf("get configurations: %w", err)
	}

	return configurations, nil
}

// ParseConfigurationsAs parses the files as the given file type and returns the
// configurations given in the file list.
func ParseConfigurationsAs(files []string, fileExtension string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, fileExtension)
	if err != nil {
		return nil, fmt.Errorf("parse configurations: %w", err)
	}

	return configurations, nil
}

// CombineConfigurations takes the given configurations and combines them into a single
// configuration.
func CombineConfigurations(configs map[string]interface{}) map[string]interface{} {
	type configuration struct {
		Path     string      `json:"path"`
		Contents interface{} `json:"contents"`
	}

	var allConfigurations []configuration
	for path, config := range configs {
		if subconfigs, exist := config.([]interface{}); exist {
			for _, subconfig := range subconfigs {
				configuration := configuration{
					Path:     path,
					Contents: subconfig,
				}

				allConfigurations = append(allConfigurations, configuration)
			}
			continue
		}

		configuration := configuration{
			Path:     path,
			Contents: config,
		}

		allConfigurations = append(allConfigurations, configuration)
	}

	combinedConfigurations := make(map[string]interface{})
	combinedConfigurations["Combined"] = allConfigurations

	return combinedConfigurations
}

func parseConfigurations(paths []string, fileExtension string) (map[string]interface{}, error) {
	parsedConfigurations := make(map[string]interface{})
	for _, path := range paths {
		var parser Parser
		var err error
		if fileExtension == "" {
			parser, err = GetParserFromPath(path)
		} else {
			parser, err = GetParser(fileExtension)
		}
		if err != nil {
			return nil, fmt.Errorf("get parser: %w", err)
		}

		contents, err := getConfigurationContent(path)
		if err != nil {
			return nil, fmt.Errorf("get configuration content: %w", err)
		}

		var parsed interface{}
		if err := parser.Unmarshal(contents, &parsed); err != nil {
			return nil, fmt.Errorf("parser unmarshal: %w", err)
		}

		parsedConfigurations[path] = parsed
	}

	return parsedConfigurations, nil
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

func getFileType(fileName string) string {
	if fileName == "-" {
		return "yaml"
	}

	if filepath.Ext(fileName) == "" {
		return filepath.Base(fileName)
	}

	fileExtension := filepath.Ext(fileName)
	return fileExtension[1:]
}
