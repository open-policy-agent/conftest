package util

import (
	"fmt"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ghodss/yaml"
)

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// GetParser returns a Parser for the given file type. If no supported parser is found, it returns an error.
func GetParser(fileName string) (Parser, error) {
	suffix := filepath.Ext(fileName)
	switch suffix {
	case ".yaml",".yml", ".json":
		return &yamlParser{
			fileName: fileName,
		}, nil
	case ".toml":
		return &tomlParser{
			fileName: fileName,
		}, nil
	default:
		return nil, fmt.Errorf("Unable to find available parser for extension %s", suffix)
	}

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
