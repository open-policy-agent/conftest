package parser

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/edn"
	"github.com/open-policy-agent/conftest/parser/hcl1"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/hocon"
	"github.com/open-policy-agent/conftest/parser/ignore"
	"github.com/open-policy-agent/conftest/parser/ini"
	"github.com/open-policy-agent/conftest/parser/json"
	"github.com/open-policy-agent/conftest/parser/jsonnet"
	"github.com/open-policy-agent/conftest/parser/properties"
	"github.com/open-policy-agent/conftest/parser/toml"
	"github.com/open-policy-agent/conftest/parser/vcl"
	"github.com/open-policy-agent/conftest/parser/xml"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

// The defined parsers are the parsers that are valid for
// parsing files.
const (
	CUE        = "cue"
	Dockerfile = "dockerfile"
	EDN        = "edn"
	HCL1       = "hcl1"
	HCL2       = "hcl2"
	HOCON      = "hocon"
	IGNORE     = "ignore"
	INI        = "ini"
	JSON       = "json"
	JSONNET    = "jsonnet"
	PROPERTIES = "properties"
	TOML       = "toml"
	VCL        = "vcl"
	XML        = "xml"
	YAML       = "yaml"
)

// Parser defines all of the methods that every parser
// definition must implement.
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// New returns a new Parser.
func New(parser string) (Parser, error) {
	switch parser {
	case TOML:
		return &toml.Parser{}, nil
	case CUE:
		return &cue.Parser{}, nil
	case INI:
		return &ini.Parser{}, nil
	case HOCON:
		return &hocon.Parser{}, nil
	case HCL1:
		return &hcl1.Parser{}, nil
	case HCL2:
		return &hcl2.Parser{}, nil
	case Dockerfile:
		return &docker.Parser{}, nil
	case YAML:
		return &yaml.Parser{}, nil
	case JSON:
		return &json.Parser{}, nil
	case JSONNET:
		return &jsonnet.Parser{}, nil
	case EDN:
		return &edn.Parser{}, nil
	case VCL:
		return &vcl.Parser{}, nil
	case XML:
		return &xml.Parser{}, nil
	case IGNORE:
		return &ignore.Parser{}, nil
	case PROPERTIES:
		return &properties.Parser{}, nil
	default:
		return nil, fmt.Errorf("unknown parser: %v", parser)
	}
}

// NewFromPath returns a file parser based on the file type
// that exists at the given path.
func NewFromPath(path string) (Parser, error) {

	// We use the YAML parser as the default when passing in configuration
	// data through standard input. This can be overridden by using the parser flag.
	if path == "-" {
		return New(YAML)
	}

	fileName := strings.ToLower(filepath.Base(path))

	fileExtension := "yml"
	if len(filepath.Ext(path)) > 0 {
		fileExtension = strings.ToLower(filepath.Ext(path)[1:])
	}

	// A Dockerfile can either be a file named Dockerfile, be prefixed with
	// Dockerfile, or have Dockerfile as its extension.
	//
	// For example: Dockerfile, Dockerfile.debug, dev.Dockerfile
	if fileName == "dockerfile" || strings.HasPrefix(fileName, "dockerfile.") || fileExtension == "dockerfile" {
		return New(Dockerfile)
	}

	if fileExtension == "yml" || fileExtension == "yaml" {
		return New(YAML)
	}

	if fileExtension == "tf" || fileExtension == "tfvars" {
		return New(HCL2)
	}

	if fileExtension == "gitignore" || fileExtension == "dockerignore" {
		return New(IGNORE)
	}

	parser, err := New(fileExtension)
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	return parser, nil
}

// Parsers returns a list of the supported Parsers.
func Parsers() []string {
	parsers := []string{
		CUE,
		Dockerfile,
		EDN,
		HCL1,
		HCL2,
		HOCON,
		IGNORE,
		INI,
		JSON,
		JSONNET,
		PROPERTIES,
		TOML,
		VCL,
		XML,
		YAML,
	}

	return parsers
}

// FileSupported returns true if the file at the given path is
// a file that can be parsed.
func FileSupported(path string) bool {
	if _, err := NewFromPath(path); err != nil {
		return false
	}

	return true
}

// ParseConfigurations parses and returns the configurations from the given
// list of files. The result will be a map where the key is the file name of
// the configuration.
func ParseConfigurations(files []string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, "")
	if err != nil {
		return nil, err
	}

	return configurations, nil
}

// ParseConfigurationsAs parses the files as the given file type and returns the
// configurations given in the file list. The result will be a map where the key
// is the file name of the configuration.
func ParseConfigurationsAs(files []string, parser string) (map[string]interface{}, error) {
	configurations, err := parseConfigurations(files, parser)
	if err != nil {
		return nil, err
	}

	return configurations, nil
}

// CombineConfigurations takes the given configurations and combines them into a single
// configuration. The result will be a map that contains a single key with a value of
// Combined.
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

	// For consistency when printing the results, sort the configurations by
	// their file paths.
	sort.Slice(allConfigurations, func(i, j int) bool {
		return allConfigurations[i].Path < allConfigurations[j].Path
	})

	combinedConfigurations := make(map[string]interface{})
	combinedConfigurations["Combined"] = allConfigurations

	return combinedConfigurations
}

func parseConfigurations(paths []string, parser string) (map[string]interface{}, error) {
	parsedConfigurations := make(map[string]interface{})
	for _, path := range paths {
		var fileParser Parser
		var err error
		if parser == "" {
			fileParser, err = NewFromPath(path)
		} else {
			fileParser, err = New(parser)
		}
		if err != nil {
			return nil, fmt.Errorf("new parser: %w", err)
		}

		contents, err := getConfigurationContent(path)
		if err != nil {
			return nil, fmt.Errorf("get configuration content: %w", err)
		}

		var parsed interface{}
		if err := fileParser.Unmarshal(contents, &parsed); err != nil {
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
