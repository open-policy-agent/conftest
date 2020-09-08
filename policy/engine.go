package policy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/version"
)

// Engine represents the policy engine.
type Engine struct {
	result   *loader.Result
	compiler *ast.Compiler
	store    storage.Store
	tracing  bool
}

// Namespaces returns all of the namespaces in the Engine.
func (e *Engine) Namespaces() []string {
	var namespaces []string
	for _, module := range e.Modules() {
		namespace := strings.Replace(module.Package.Path.String(), "data.", "", 1)
		if contains(namespaces, namespace) {
			continue
		}

		namespaces = append(namespaces, namespace)
	}

	return namespaces
}

// Documents returns all of the documents loaded into the engine.
func (e *Engine) Documents() map[string]string {
	documents := make(map[string]string)
	for path, content := range e.result.Documents {
		documents[path] = fmt.Sprintf("%v", content)
	}

	return documents
}

// Policies returns all of the policies loaded into the engine.
func (e *Engine) Policies() map[string]string {
	policies := make(map[string]string)
	for m := range e.result.Modules {
		policies[e.result.Modules[m].Name] = string(e.result.Modules[m].Raw)
	}

	return policies
}

// Compiler returns the compiler from the loaded policies.
func (e *Engine) Compiler() *ast.Compiler {
	return e.compiler
}

// Store returns the store from the loaded documents.
func (e *Engine) Store() storage.Store {
	return e.store
}

// Modules returns the modules from the loaded policies.
func (e *Engine) Modules() map[string]*ast.Module {
	return e.result.ParsedModules()
}

// Runtime returns the runtime of the engine.
func (e *Engine) Runtime() *ast.Term {
	env := ast.NewObject()
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}

	obj := ast.NewObject()
	obj.Insert(ast.StringTerm("env"), ast.NewTerm(env))
	obj.Insert(ast.StringTerm("version"), ast.StringTerm(version.Version))
	obj.Insert(ast.StringTerm("commit"), ast.StringTerm(version.Vcs))

	return ast.NewTerm(obj)
}

// Query the policy engine with the given query and given input.
func (e *Engine) Query(ctx context.Context, query string, input interface{}) ([]output.Result, []output.Result, error) {
	stdout := topdown.NewBufferTracer()
	buf := new(bytes.Buffer)

	topdown.PrettyTrace(buf, *stdout)
	var traces []error
	for _, line := range strings.Split(buf.String(), "\n") {
		if len(line) > 0 {
			traces = append(traces, errors.New(line))
		}
	}

	var regoFunc []func(r *rego.Rego)
	regoFunc = append(regoFunc, rego.Query(query), rego.Compiler(e.Compiler()), rego.Input(input), rego.Store(e.Store()), rego.Runtime(e.Runtime()))
	if e.tracing {
		regoFunc = append(regoFunc, rego.Tracer(stdout))
	}

	regoObj := rego.New(regoFunc...)
	resultSet, err := regoObj.Eval(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating policy: %w", err)
	}

	var failures []output.Result
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
					failures = append(failures, output.NewResult(val, traces))
				case map[string]interface{}:
					failure, err := getResult(val, traces)
					if err != nil {
						return nil, nil, fmt.Errorf("get result: %w", err)
					}

					failures = append(failures, failure)
				}
			}
		}
	}

	return failures, successes, nil
}

func getResult(val map[string]interface{}, traces []error) (output.Result, error) {
	if _, ok := val["msg"]; !ok {
		return output.Result{}, fmt.Errorf("rule missing msg field: %v", val)
	}
	if _, ok := val["msg"].(string); !ok {
		return output.Result{}, fmt.Errorf("msg field must be string: %v", val)
	}

	result := output.NewResult(val["msg"].(string), traces)
	for k, v := range val {
		if k != "msg" {
			result.Metadata[k] = v
		}
	}

	return result, nil
}

func hasResults(expression interface{}) bool {
	if v, ok := expression.([]interface{}); ok {
		return len(v) > 0
	}

	return false
}

func contains(collection []string, item string) bool {
	for _, value := range collection {
		if strings.EqualFold(value, item) {
			return true
		}
	}

	return false
}
