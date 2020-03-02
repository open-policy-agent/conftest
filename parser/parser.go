package parser

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

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
		"hcl1",
		"hcl2",
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

// ConfigDoc stores file contents and it's original filename
type ConfigDoc struct {
	ReadCloser io.ReadCloser
	Filepath   string
}

// BulkUnmarshal iterates through the given cached io.Readers and
// runs the requested parser on the data.
func BulkUnmarshal(configList []ConfigDoc, input string) (map[string]interface{}, error) {
	configContents := make(map[string][]byte)
	for _, config := range configList {
		contents, err := ioutil.ReadAll(config.ReadCloser)
		if err != nil {
			return nil, fmt.Errorf("read config: %w", err)
		}

		configContents[config.Filepath] = contents
		config.ReadCloser.Close()
	}

	var allContents = make(map[string]interface{})
	for filePath, config := range configContents {
		fileType := getFileType(filePath, input)

		fileParser, err := GetParser(fileType)
		if err != nil {
			return nil, fmt.Errorf("get parser: %w", err)
		}

		var singleContent interface{}
		if err := fileParser.Unmarshal(config, &singleContent); err != nil {
			return nil, fmt.Errorf("parser unmarshal: %w", err)
		}

		allContents[filePath] = singleContent
	}

	return allContents, nil
}

// GetParser gets a file parser based on the file type and input
func GetParser(fileType string) (Parser, error) {
	switch fileType {
	case "toml":
		return &toml.Parser{}, nil
	case "hcl1":
		return &terraform.Parser{}, nil
	case "cue":
		return &cue.Parser{}, nil
	case "ini":
		return &ini.Parser{}, nil
	case "hocon":
		return &hocon.Parser{}, nil
	case "hcl", "tf":
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

func getFileType(fileName string, input string) string {
	if input != "" {
		return input
	}

	if fileName == "-" {
		return "yaml"
	}

	if filepath.Ext(fileName) == "" {
		return filepath.Base(fileName)
	}

	fileExtension := filepath.Ext(fileName)

	return fileExtension[1:]
}
