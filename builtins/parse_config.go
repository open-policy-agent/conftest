package builtins

import (
	"fmt"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
)

func init() {
	decl := rego.Function{
		Name: "parse_config",
		Decl: types.NewFunction(
			types.Args(types.S, types.S), // parser name, configuration
			types.NewObject(nil, types.NewDynamicProperty(types.S, types.NewAny())), // map[string]interface{} aka JSON
		),
	}
	rego.RegisterBuiltin2(&decl, parseConfig)
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
	value, err := ast.InterfaceToValue(cfg)
	if err != nil {
		return nil, fmt.Errorf("convert config to ast.Value: %w", err)
	}
	loc := &ast.Location{
		File: "-", // stdin
		Text: []byte(config),
	}
	if bctx.Location != nil {
		loc = bctx.Location
	}
	term := &ast.Term{
		Value:    value,
		Location: loc,
	}

	return term, nil
}

func decodeArgs(args []*ast.Term) ([]interface{}, error) {
	decoded := make([]interface{}, len(args))
	for i, arg := range args {
		v, err := ast.ValueToInterface(arg.Value, nil)
		if err != nil {
			return nil, fmt.Errorf("ast.ValueToInterface: %w", err)
		}
		decoded[i] = v
	}
	return decoded, nil
}
