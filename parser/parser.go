package parser

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/instrumenta/conftest/parser/cue"
	"github.com/instrumenta/conftest/parser/docker"
	"github.com/instrumenta/conftest/parser/edn"
	"github.com/instrumenta/conftest/parser/hcl2"
	"github.com/instrumenta/conftest/parser/hocon"
	"github.com/instrumenta/conftest/parser/ini"
	"github.com/instrumenta/conftest/parser/terraform"
	"github.com/instrumenta/conftest/parser/toml"
	"github.com/instrumenta/conftest/parser/vcl"
	"github.com/instrumenta/conftest/parser/xml"
	"github.com/instrumenta/conftest/parser/yaml"
)

// ValidInputs returns string array in order to passing valid input types to viper
func ValidInputs() []string {
	return []string{
		"toml",
		"tf|hcl",
		"hcl2",
		"cue",
		"ini",
		"yml|yaml|json",
		"Dockerfile",
		"edn",
		"vcl",
		"xml",
	}
}

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// ConfigDoc stores file contents and it's original filename
type ConfigDoc struct {
	ReadCloser io.ReadCloser
	Filepath   string
}

// ConfigManager the implementation of ReadUnmarshaller and io.Reader
// byte storage.
type ConfigManager struct {
	parser         Parser
	configContents map[string][]byte
}

// BulkUnmarshal iterates through the given cached io.Readers and
// runs the requested parser on the data.
func (s *ConfigManager) BulkUnmarshal(configList []ConfigDoc) (map[string]interface{}, error) {
	if err := s.setConfigs(configList); err != nil {
		return nil, fmt.Errorf("set configuration: %w", err)
	}

	var allContents = make(map[string]interface{})
	for filepath, config := range s.configContents {
		var singleContent interface{}
		if err := s.parser.Unmarshal(config, &singleContent); err != nil {
			return nil, fmt.Errorf("parser unmarshal: %w", err)
		}

		allContents[filepath] = singleContent
	}

	return allContents, nil
}

func (s *ConfigManager) setConfigs(configList []ConfigDoc) error {
	s.configContents = make(map[string][]byte)
	for _, config := range configList {
		contents, err := ioutil.ReadAll(config.ReadCloser)
		defer config.ReadCloser.Close()
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		s.configContents[config.Filepath] = contents
	}

	return nil
}

// NewConfigManager is the instatiation function for ConfigManager
func NewConfigManager(fileType string) (*ConfigManager, error) {
	parser, err := GetParser(fileType)
	if err != nil {
		return nil, fmt.Errorf("get parser: %w", err)
	}

	config := ConfigManager{
		parser: parser,
	}

	return &config, nil
}

// GetParser gets a parser that works on a given fileType
func GetParser(fileType string) (Parser, error) {
	switch fileType {
	case "toml":
		return &toml.Parser{}, nil
	case "tf", "hcl":
		return &terraform.Parser{}, nil
	case "cue":
		return &cue.Parser{}, nil
	case "ini":
		return &ini.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "hcl2":
		return &hcl2.Parser{}, nil
	case "Dockerfile", "dockerfile":
		return &docker.Parser{}, nil
	case "yml", "yaml", "json":
		return &yaml.Parser{}, nil
	case "edn":
		return &edn.Parser{}, nil
	case "vcl":
		return &vcl.Parser{}, nil
	case "xml":
		return &xml.Parser{}, nil
	default:
		return nil, fmt.Errorf("unknown filetype given: %v", fileType)
	}
}
