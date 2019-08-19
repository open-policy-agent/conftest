package parser

import (
	"github.com/instrumenta/conftest/pkg/parser/cue"
	"github.com/instrumenta/conftest/pkg/parser/ini"
	"github.com/instrumenta/conftest/pkg/parser/terraform"
	"github.com/instrumenta/conftest/pkg/parser/toml"
	"github.com/instrumenta/conftest/pkg/parser/yaml"
	"github.com/instrumenta/conftest/pkg/parser/docker"
)

// Parser is the interface implemented by objects that can unmarshal
// bytes into a golang interface
type Parser interface {
	Unmarshal(p []byte, v interface{}) error
}

// GetParser returns a Parser for the given input type. Defaults to returning the YAML parser.
func GetParser(i *Input) Parser {

	switch i.input {
	case "toml":
		return &toml.Parser{
			FileName: i.fName,
		}
	case "tf", "hcl":
		return &terraform.Parser{
			FileName: i.fName,
		}
	case "cue":
		return &cue.Parser{
			FileName: i.fName,
		}
	case "ini":
		return &ini.Parser{
			FileName: i.fName,
		}
	case "Dockerfile":
		return &docker.Parser{
			FileName: i.fName,
		}
	default:
		return &yaml.Parser{
			FileName: i.fName,
		}
	}
}
