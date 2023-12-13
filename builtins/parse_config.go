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
// parsed configuration as a Rego object. This can be used to parse all of the
// configuration formats conftest supports in-line in Rego policies.
func parseConfig(bctx rego.BuiltinContext, op1, op2 *ast.Term) (*ast.Term, error) {
	args, err := decodeArgs[string](op1, op2)
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	parserName, config := args[0], args[1]

	parser, err := parser.New(parserName)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}
	var cfg any
	if err := parser.Unmarshal([]byte(config), &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return toAST(bctx, cfg, []byte(config))
}

// parseConfigFile takes a config file path, parses the config file, and
// returns the parsed configuration as a Rego object.
func parseConfigFile(bctx rego.BuiltinContext, op1 *ast.Term) (*ast.Term, error) {
	path, err := decodeArg[string](op1)
	if err != nil {
		return nil, fmt.Errorf("decode args: %w", err)
	}
	filePath := filepath.Join(filepath.Dir(bctx.Location.File), path)

	parser, err := parser.NewFromPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("create config parser: %w", err)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", filePath, err)
	}

	var cfg any
	if err := parser.Unmarshal(contents, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return toAST(bctx, cfg, contents)
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
	content, err := json.Marshal(combined)
	if err != nil {
		return nil, fmt.Errorf("marshal combined content: %w", err)
	}

	return toAST(bctx, combined["Combined"], content)
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

func toAST(bctx rego.BuiltinContext, cfg any, contents []byte) (*ast.Term, error) {
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
