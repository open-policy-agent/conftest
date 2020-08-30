package policy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
)

// Engine represents the policy engine
type Engine struct {
	Compiler *ast.Compiler
	Store storage.Store
	Trace bool
}

// NewEngine returns a new instatiated Engine
func NewEngine(compiler *ast.Compiler, store storage.Store, trace bool) (*Engine) {
	return &Engine{
		Compiler: compiler,
		Store: store,
		Trace: trace,
	}
}

// Query the policy engine with the given query and given input.
func (e *Engine) Query(ctx context.Context, query string, input interface{}) ([]output.Result, []output.Result, error) {
	rego, stdout := e.buildRego(e.Trace, query, input)
	resultSet, err := rego.Eval(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating policy: %w", err)
	}

	buf := new(bytes.Buffer)
	topdown.PrettyTrace(buf, *stdout)
	var traces []error
	for _, line := range strings.Split(buf.String(), "\n") {
		if len(line) > 0 {
			traces = append(traces, errors.New(line))
		}
	}

	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}

		return false
	}

	var errs []output.Result
	var successes []output.Result
	for _, result := range resultSet {
		for _, expression := range result.Expressions {
			if !hasResults(expression.Value) {
				successes = append(successes, output.NewResult("", traces))
				continue
			}

			for _, v := range expression.Value.([]interface{}) {
				switch val := v.(type) {
				case string:
					errs = append(errs, output.NewResult(val, traces))
				case map[string]interface{}:
					if _, ok := val["msg"]; !ok {
						return nil, nil, fmt.Errorf("rule missing msg field: %v", val)
					}
					if _, ok := val["msg"].(string); !ok {
						return nil, nil, fmt.Errorf("msg field must be string: %v", val)
					}

					result := output.NewResult(val["msg"].(string), traces)
					for k, v := range val {
						if k != "msg" {
							result.Metadata[k] = v
						}

					}
					errs = append(errs, result)
				}
			}
		}
	}

	return errs, successes, nil
}

func (e *Engine) buildRego(trace bool, query string, input interface{}) (*rego.Rego, *topdown.BufferTracer) {
	var regoObj *rego.Rego
	var regoFunc []func(r *rego.Rego)
	buf := topdown.NewBufferTracer()
	runtime := RuntimeTerm()

	regoFunc = append(regoFunc, rego.Query(query), rego.Compiler(e.Compiler), rego.Input(input), rego.Store(e.Store), rego.Runtime(runtime))
	if trace {
		regoFunc = append(regoFunc, rego.Tracer(buf))
	}

	regoObj = rego.New(regoFunc...)

	return regoObj, buf
}