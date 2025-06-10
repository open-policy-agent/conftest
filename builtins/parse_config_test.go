package builtins

import (
	"context"
	"strings"
	"testing"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

func TestParseConfig(t *testing.T) {
	testCases := []struct {
		desc       string
		parser     string
		config     string
		wantErrMsg string
	}{
		{
			desc:       "No parser supplied",
			wantErrMsg: "create config parser",
		},
		{
			desc:       "Invalid parser supplied",
			parser:     "no-such-parser",
			wantErrMsg: "create config parser",
		},
		{
			desc:       "Invalid YAML",
			parser:     parser.YAML,
			config:     "```NOTVALID!",
			wantErrMsg: "unmarshal config",
		},
		{
			desc:   "Empty YAML",
			parser: parser.YAML,
		},
		{
			desc:   "Valid YAML",
			parser: parser.YAML,
			config: `some_field: some_value
another_field:
- arr1
- arr2`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			pv, err := ast.InterfaceToValue(tc.parser)
			if err != nil {
				t.Fatalf("Could not convert parser %q to ast.Value: %v", tc.parser, err)
			}
			cv, err := ast.InterfaceToValue(tc.config)
			if err != nil {
				t.Fatalf("Could not convert config %q to ast.Value: %v", tc.config, err)
			}

			bctx := rego.BuiltinContext{Context: context.Background()}
			_, err = parseConfig(bctx, ast.NewTerm(pv), ast.NewTerm(cv))
			if err == nil && tc.wantErrMsg == "" {
				return
			}
			if err != nil && tc.wantErrMsg == "" {
				t.Errorf("Error was returned when no error was expected: %v", err)
				return
			}
			if !strings.Contains(err.Error(), tc.wantErrMsg) {
				t.Errorf("Error %q does not contain expected string %q", err.Error(), tc.wantErrMsg)
				return
			}
		})
	}
}
