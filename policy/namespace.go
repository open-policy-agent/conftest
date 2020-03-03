package policy

import (
	"strings"

	"github.com/open-policy-agent/opa/ast"
)

// GetNamespaces returns a list of all namespaces in a set of given policies
func GetNamespaces(regoFiles []string, compiler *ast.Compiler) ([]string, error) {
	namespaces := []string{}
	exists := map[string]bool{}
	for _, regoFile := range regoFiles {
		for _, path := range compiler.Modules[regoFile].Package.Path {
			value := path.Value.String()
			if value != "data" && !exists[value] {
				namespaces = append(namespaces, strings.ReplaceAll(path.Value.String(), "\"", ""))
				exists[value] = true
			}
		}
	}
	return namespaces, nil
}
