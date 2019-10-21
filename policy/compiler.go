package policy

import (
	"io/ioutil"

	"github.com/open-policy-agent/opa/ast"
)

// BuildCompiler compiles all of the given Rego policies
// and returns the Compiler containing the compilation state
func BuildCompiler(files []string) (*ast.Compiler, error) {
	modules := map[string]*ast.Module{}

	for _, file := range files {
		out, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		parsed, err := ast.ParseModule(file, string(out[:]))
		if err != nil {
			return nil, err
		}

		modules[file] = parsed
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)
	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}
