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
		namespace := strings.Replace(compiler.Modules[regoFile].Package.Path.String(), "data.", "", 1)
		if !exists[namespace] {
			namespaces = append(namespaces, namespace)
			exists[namespace] = true
		}
	}
	return namespaces, nil
}
