package parser

import (
	"fmt"

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
		"tf",
		"hcl",
		"hcl1",
		"cue",
		"ini",
		"yml",
		"yaml",
		"json",
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

// GetParser gets a file parser based on the file type and input
func GetParser(fileType string) (Parser, error) {
	switch fileType {
	case "toml":
		return &toml.Parser{}, nil
	case "hcl1":
		// route old hcl parser
		return &terraform.Parser{}, nil
	case "cue":
		return &cue.Parser{}, nil
	case "ini":
		return &ini.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "hcl", "tf", "hcl2":
		// route to new hcl
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
