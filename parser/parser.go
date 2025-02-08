package parser

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/open-policy-agent/conftest/parser/cyclonedx"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	"github.com/open-policy-agent/conftest/parser/cue"
	"github.com/open-policy-agent/conftest/parser/docker"
	dotenv "github.com/open-policy-agent/conftest/parser/dotenv"
	"github.com/open-policy-agent/conftest/parser/edn"
	"github.com/open-policy-agent/conftest/parser/hcl1"
	"github.com/open-policy-agent/conftest/parser/hcl2"
	"github.com/open-policy-agent/conftest/parser/hocon"
	"github.com/open-policy-agent/conftest/parser/ignore"
	"github.com/open-policy-agent/conftest/parser/ini"
	"github.com/open-policy-agent/conftest/parser/json"
	"github.com/open-policy-agent/conftest/parser/jsonc"
	"github.com/open-policy-agent/conftest/parser/jsonnet"
	"github.com/open-policy-agent/conftest/parser/properties"
	"github.com/open-policy-agent/conftest/parser/spdx"
	"github.com/open-policy-agent/conftest/parser/textproto"
	"github.com/open-policy-agent/conftest/parser/toml"
	"github.com/open-policy-agent/conftest/parser/vcl"
	"github.com/open-policy-agent/conftest/parser/xml"
	"github.com/open-policy-agent/conftest/parser/yaml"
)

// The defined parsers are the parsers that are valid for
// parsing files.
const (
	CUE        = "cue"
	CYCLONEDX  = "cyclonedx"
	Dockerfile = "dockerfile"
	EDN        = "edn"
	HCL1       = "hcl1"
	HCL2       = "hcl2"
	HOCON      = "hocon"
	IGNORE     = "ignore"
	INI        = "ini"
	JSON       = "json"
	JSONC      = "jsonc"
	JSONNET    = "jsonnet"
	PROPERTIES = "properties"
	SPDX       = "spdx"
	TEXTPROTO  = "textproto"
	TOML       = "toml"
	VCL        = "vcl"
	XML        = "xml"
	YAML       = "yaml"
	DOTENV     = "dotenv"
)

// Parser defines all of the methods that every parser
// definition must implement.
type Parser interface {
	Unmarshal(p []byte, v any) error
}

// PathAwareParser is an optional interface that parsers may implement
// if they need the original file path for relative imports or other logic.
type PathAwareParser interface {
	Parser
	SetPath(path string)
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
	case JSONC:
		return &jsonc.Parser{}, nil
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
	case SPDX:
		return &spdx.Parser{}, nil
	case CYCLONEDX:
		return &cyclonedx.Parser{}, nil
	case DOTENV:
		return &dotenv.Parser{}, nil
	case TEXTPROTO:
		parser := &textproto.Parser{}
		if dirs := viper.GetStringSlice("proto-file-dirs"); len(dirs) > 0 {
			files, err := findFilesWithExt(dirs, ".proto")
			if err != nil {
				return nil, fmt.Errorf("find proto files: %w", err)
			}
			if err := parser.LoadProtoFiles(files); err != nil {
				return nil, fmt.Errorf("load protos: %w", err)
			}
		}

		return parser, nil
	default:
		return nil, fmt.Errorf("unknown parser: %v", parser)
	}
}

func findFilesWithExt(dirs []string, ext string) ([]string, error) {
	var files []string
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(info.Name(), ext) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk dir %q: %w", dir, err)
		}
	}
	return files, nil
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

	if fileExtension == "hcl" || fileExtension == "tf" || fileExtension == "tfvars" {
		return New(HCL2)
	}

	if fileExtension == "gitignore" || fileExtension == "dockerignore" {
		return New(IGNORE)
	}

	// A .env can either be a file named .env, be prefixed with
	// .env, or have .env as its extension.
	//
	// For example: .env, .env.prod, prod.env
	if fileName == ".env" || strings.HasPrefix(fileName, ".env.") || fileExtension == "env" {
		return New(DOTENV)
	}

	if slices.Contains(textproto.TextProtoFileExtensions, fileExtension) {
		return New(TEXTPROTO)
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
		SPDX,
		TEXTPROTO,
		TOML,
		VCL,
		XML,
		YAML,
		DOTENV,
	}

	return parsers
}

// FileSupported returns true if the file at the given path is
// a file that can be parsed.
func FileSupported(path string) bool {
	_, err := NewFromPath(path)
	return err == nil
}

// ParseConfigurations parses and returns the configurations from the given
// list of files. The result will be a map where the key is the file name of
// the configuration.
func ParseConfigurations(files []string) (map[string]any, error) {
	configurations, err := parseConfigurations(files, "")
	if err != nil {
		return nil, err
	}

	return configurations, nil
}

// ParseConfigurationsAs parses the files as the given file type and returns the
// configurations given in the file list. The result will be a map where the key
// is the file name of the configuration.
func ParseConfigurationsAs(files []string, parser string) (map[string]any, error) {
	configurations, err := parseConfigurations(files, parser)
	if err != nil {
		return nil, err
	}

	return configurations, nil
}

// CombineConfigurations takes the given configurations and combines them into a single
// configuration. The result will be a map that contains a single key with a value of
// Combined.
func CombineConfigurations(configs map[string]any) map[string]any {
	type configuration struct {
		Path     string `json:"path"`
		Contents any    `json:"contents"`
	}

	var allConfigurations []configuration
	for path, config := range configs {
		if subconfigs, exist := config.([]any); exist {
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

	combinedConfigurations := make(map[string]any)
	combinedConfigurations["Combined"] = allConfigurations

	return combinedConfigurations
}

func parseConfigurations(paths []string, parser string) (map[string]any, error) {
	parsedConfigurations := make(map[string]any)
	errWithPathInfo := func(err error, msg, path string) error {
		return fmt.Errorf("%s: %w, path: %s", msg, err, path)
	}
	for _, path := range paths {
		var fileParser Parser
		var err error
		if parser == "" {
			fileParser, err = NewFromPath(path)
		} else {
			fileParser, err = New(parser)
		}
		if err != nil {
			return nil, errWithPathInfo(err, "new parser", path)
		}

		contents, err := getConfigurationContent(path)
		if err != nil {
			return nil, errWithPathInfo(err, "get configuration content", path)
		}

		// If our parser needs the path, give it the path
		if p, ok := fileParser.(PathAwareParser); ok {
			p.SetPath(path)
		}

		var parsed any
		if err := fileParser.Unmarshal(contents, &parsed); err != nil {
			return nil, errWithPathInfo(err, "parser unmarshal", path)
		}

		parsedConfigurations[path] = parsed
	}

	return parsedConfigurations, nil
}

func getConfigurationContent(path string) ([]byte, error) {
	if path == "-" {
		contents, err := io.ReadAll(bufio.NewReader(os.Stdin))
		if err != nil {
			return nil, fmt.Errorf("read standard in: %w", err)
		}

		return contents, nil
	}

	filePath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("get abs: %w", err)
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return contents, nil
}
