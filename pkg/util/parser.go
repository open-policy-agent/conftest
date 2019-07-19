package util

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform/config"
)

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// GetParser returns a Parser for the given file type. Defaults to returning the YAML parser.
func GetParser(fileName string) Parser {
	suffix := filepath.Ext(fileName)
	switch suffix {
	case ".toml":
		return &tomlParser{
			fileName: fileName,
		}
	case ".tf":
		return &tfParser{
			fileName: fileName,
		}
	default:
		return &yamlParser{
			fileName: fileName,
		}
	}
}

type tfParser struct {
	fileName string
}

func (s *tfParser) Unmarshal(p []byte, v interface{}) error {
	filePath, _ := filepath.Abs(s.fileName)
	cfg, err := config.LoadFile(filePath)
	if err != nil {
		return fmt.Errorf("load terraform config failed: %v", err)
	}

	j, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("Unable to marshal config %s: %s", s.fileName, err)
	}

	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from HCL-json %s: %s", s.fileName, err)
	}

	return nil
}

type yamlParser struct {
	fileName string
}

func (yp *yamlParser) Unmarshal(p []byte, v interface{}) error {
	err := yaml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from %s: %s", yp.fileName, err)
	}

	return nil
}

type tomlParser struct {
	fileName string
}

func (tp *tomlParser) Unmarshal(p []byte, v interface{}) error {
	err := toml.Unmarshal(p, v)
	if err != nil {
		return fmt.Errorf("Unable to parse TOML from %s: %s", tp.fileName, err)
	}

	return nil
}
