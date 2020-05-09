package parser

import (
	"fmt"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/edn"
	"github.com/open-policy-agent/conftest/parser/hcl1"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/hocon"
	"github.com/open-policy-agent/conftest/parser/ini"
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
		return &hcl1.Parser{}, nil
	case "cue":
		return &cue.Parser{}, nil
	case "ini":
		return &ini.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "hcl", "tf", "hcl2":
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
