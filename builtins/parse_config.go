package builtins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/ast/location"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func init() {
	registerParseConfig()
	registerParseConfigFile()
	registerParseCombinedConfigFiles()
}

func registerParseConfig() {
	decl := rego.Function{
		Name: "parse_config",
		Decl: types.NewFunction(
			types.Args(types.S, types.S), // parser name, configuration
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]interface{} aka JSON
		),
	}
	rego.RegisterBuiltin2(&decl, parseConfig)
}

func registerParseConfigFile() {
	decl := rego.Function{
		Name: "parse_config_file",
		Decl: types.NewFunction(
			types.Args(types.S), // path to configuration file
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]interface{} aka JSON
		),
	}
	rego.RegisterBuiltin1(&decl, parseConfigFile)
}

func registerParseCombinedConfigFiles() {
	decl := rego.Function{
		Name: "parse_combined_config_files",
		Decl: types.NewFunction(
			types.Args(types.NewArray(nil, types.S)),                                // paths to configuration files
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]interface{} aka JSON
		),
	}
	rego.RegisterBuiltin1(&decl, parseCombinedConfigFiles)
}

// parseConfig takes a parser name and configuration as strings and returns the
// parsed configuration as a Rego object. This can be used to parse all of the
// configuration formats conftest supports in-line in Rego policies.
func parseConfig(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
	args, err := decodeArgs([]*ast.Term{op1, op2})
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	parserName, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("parser name %v [%T] is not expected type string", args[0], args[0])
	}
	config, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("config %v [%T] is not expected type string", args[1], args[1])
	}
	parser, err := parser.New(parserName)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}

	var cfg map[string]interface{}
	if err := parser.Unmarshal([]byte(config), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return toAST(bctx, cfg, []byte(config))
}

// parseConfigFile takes a config file path, parses the config file, and
// returns the parsed configuration as a Rego object.
func parseConfigFile(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	args, err := decodeArgs([]*ast.Term{op1})
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	file, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("file %v [%T] is not expected type string", args[0], args[0])
	}
	filePath := filepath.Join(filepath.Dir(bctx.Location.File), file)
	parser, err := parser.NewFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", filePath, err)
	}

	var cfg map[string]interface{}
	if err := parser.Unmarshal(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return toAST(bctx, cfg, contents)
}

// parseCombinedConfigFiles
func parseCombinedConfigFiles(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	args, err := decodeArgs([]*ast.Term{op1})
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}

	filePaths := []string{}
	fileList := args[0].([]interface{})
	for _, file := range fileList {
		filePaths = append(filePaths, filepath.Join(filepath.Dir(bctx.Location.File), file.(string)))
	}

	cfg, err := parser.ParseConfigurations(filePaths)
	if err != nil {
		return nil, fmt.Errorf("failed to parse combine configurations: %w", err)
	}
	combinedCfg := parser.CombineConfigurations(cfg)
	combinedContent, err := json.Marshal(combinedCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal combined content: %w", err)
	}

	return toAST(bctx, combinedCfg["Combined"], combinedContent)
}

func decodeArgs(args []*ast.Term) ([]interface{}, error) {
	decoded := make([]interface{}, len(args))
	for i, arg := range args {
		iface, err := ast.ValueToInterface(arg.Value, nil)
		if err != nil {
			return nil, fmt.Errorf("ast.ValueToInterface: %w", err)
		}
		decoded[i] = iface
	}

	return decoded, nil
}

func toAST(bctx rego.BuiltinContext, cfg interface{}, contents []byte) (*ast.Term, error) {
	val, err := ast.InterfaceToValue(cfg)
	if err != nil {
		return nil, fmt.Errorf("convert config to ast.Value: %w", err)
	}
	var loc *location.Location
	if bctx.Location != nil {
		loc = bctx.Location
	} else {
		loc = &ast.Location{
			File: "-", // stdin
			Text: contents,
		}
	}

	return &ast.Term{Value: val, Location: loc}, nil
}
