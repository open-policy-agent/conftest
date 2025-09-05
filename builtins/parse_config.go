package builtins

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
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
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]any aka JSON
		),
	}
	rego.RegisterBuiltin2(&decl, parseConfig)
}

func registerParseConfigFile() {
	decl := rego.Function{
		Name: "parse_config_file",
		Decl: types.NewFunction(
			types.Args(types.S), // path to configuration file
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]any aka JSON
		),
	}
	rego.RegisterBuiltin1(&decl, parseConfigFile)
}

func registerParseCombinedConfigFiles() {
	decl := rego.Function{
		Name: "parse_combined_config_files",
		Decl: types.NewFunction(
			types.Args(types.NewArray(nil, types.S)),                                // paths to configuration files
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]any aka JSON
		),
	}
	rego.RegisterBuiltin1(&decl, parseCombinedConfigFiles)
}

// parseConfig takes a parser name and configuration as strings and returns the
// parsed configuration as a Rego object. This can be used to parse all the
// configuration formats conftest supports in-line in Rego policies.
func parseConfig(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
	args, err := decodeArgs[string](op1, op2)
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	parserName, config := args[0], args[1]
	configParser, err := parser.New(parserName)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}
	if bctx.Location == nil {
		bctx.Location = &ast.Location{
			File: "-", // stdin (inline)
			Text: []byte(config),
		}
	}
	return parseConfigToAST(bctx, configParser, bytes.NewBufferString(config))
}

func parseConfigToAST(bctx rego.BuiltinContext, configParser parser.Parser, config io.Reader) (*ast.Term, error) {
	parsed, err := configParser.Parse(config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	// maintain backwards compatibility and provide single document as non-slice cfg
	var cfg any
	if len(parsed) == 1 {
		cfg = parsed[0]
	} else {
		cfg = parsed
	}
	return toAST(bctx, cfg)
}

// parseConfigFile takes a config file path, parses the config file, and
// returns the parsed configuration as a Rego object.
func parseConfigFile(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	path, err := decodeArg[string](op1)
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	filePath := filepath.Join(filepath.Dir(bctx.Location.File), path)

	configParser, err := parser.NewFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open config file %s: %w", filePath, err)
	}
	return parseConfigToAST(bctx, configParser, file)
}

// parseCombinedConfigFiles takes multiple config file paths, parses the configs,
// combines them, and returns that as a Rego object.
func parseCombinedConfigFiles(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	paths, err := decodeSliceArg[string](op1)
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	for i, p := range paths {
		paths[i] = filepath.Join(filepath.Dir(bctx.Location.File), p)
	}

	cfg, err := parser.ParseConfigurations(paths)
	if err != nil {
		return nil, fmt.Errorf("parse combine configurations: %w", err)
	}
	combined := parser.CombineConfigurations(cfg)
	return toAST(bctx, combined["Combined"])
}

func decodeSliceArg[T any](arg *ast.Term) ([]T, error) {
	iface, err := ast.ValueToInterface(arg.Value, nil)
	if err != nil {
		return nil, fmt.Errorf("decode arg: %w", err)
	}
	ifaceSlice, ok := iface.([]any)
	if !ok {
		return nil, fmt.Errorf("decodeSliceArg used with non-slice value: (%T)%v", iface, iface)
	}

	var t T
	slice := make([]T, len(ifaceSlice))
	for i, val := range ifaceSlice {
		v, ok := val.(T)
		if !ok {
			return nil, fmt.Errorf("slice index %d is not expected type %T, got %T", i, t, val)
		}
		slice[i] = v
	}

	return slice, nil
}

func decodeArg[T any](arg *ast.Term) (T, error) {
	iface, err := ast.ValueToInterface(arg.Value, nil)
	if err != nil {
		return *new(T), fmt.Errorf("ast.ValueToInterface: %w", err)
	}
	v, ok := iface.(T)
	if !ok {
		return *new(T), fmt.Errorf("argument is not expected type, have %T", iface)
	}

	return v, nil
}

func decodeArgs[T any](args ...*ast.Term) ([]T, error) {
	decoded := make([]T, len(args))
	for i, arg := range args {
		v, err := decodeArg[T](arg)
		if err != nil {
			return nil, err
		}
		decoded[i] = v
	}

	return decoded, nil
}

func toAST(bctx rego.BuiltinContext, cfg any) (*ast.Term, error) {
	val, err := ast.InterfaceToValue(cfg)
	if err != nil {
		return nil, fmt.Errorf("convert config to ast.Value: %w", err)
	}
	return &ast.Term{Value: val, Location: bctx.Location}, nil
}
