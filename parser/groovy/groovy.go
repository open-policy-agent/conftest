package groovy

import (
	"encoding/json"
	"fmt"

	groovylang "github.com/albertocavalcante/groovy-parser-go"
	groovycst "github.com/albertocavalcante/groovy-parser-go/cst"
	groovyparser "github.com/albertocavalcante/groovy-parser-go/parser"
	groovysource "github.com/albertocavalcante/groovy-parser-go/source"
	groovytoken "github.com/albertocavalcante/groovy-parser-go/token"
)

const (
	// Bound the recursive conversion and JSON encoding of untrusted pipeline source.
	maxTreeDepth = 256
	maxTreeNodes = 100_000
)

// Node represents a Jenkins Groovy concrete syntax tree node. Interior nodes
// may have children, while token nodes have text. Locations are one-based.
type Node struct {
	Kind     string `json:"kind"`
	Text     string `json:"text,omitempty"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Children []Node `json:"children,omitempty"`
}

// Parser parses Groovy source using the Groovy 2.4 syntax supported by Jenkins.
type Parser struct{}

// Unmarshal converts Groovy source into a JSON-compatible concrete syntax tree.
func (p *Parser) Unmarshal(b []byte, v any) error {
	source := groovysource.New("", b)
	parser := groovyparser.New(source, groovylang.ParseOptions{Version: groovylang.Groovy2_4})
	root, parseErrors := parser.Parse()
	if len(parseErrors) > 0 {
		first := parseErrors[0]
		position := source.Position(first.Offset)
		return fmt.Errorf("parse groovy at %s: %s", position, first.Message)
	}

	rootNode := groovycst.NewTree(root)
	if err := validateTree(rootNode, maxTreeDepth, maxTreeNodes); err != nil {
		return fmt.Errorf("parse groovy: %w", err)
	}

	tree := convertNode(rootNode, source)
	encoded, err := json.Marshal(tree)
	if err != nil {
		return fmt.Errorf("marshal groovy to json: %w", err)
	}

	if err := json.Unmarshal(encoded, v); err != nil {
		return fmt.Errorf("unmarshal groovy json: %w", err)
	}

	return nil
}

func validateTree(root *groovycst.Node, maxDepth, maxNodes int) error {
	type entry struct {
		node  *groovycst.Node
		depth int
	}

	stack := []entry{{node: root, depth: 1}}
	nodes := 1
	if nodes > maxNodes {
		return fmt.Errorf("syntax tree exceeds maximum node count of %d", maxNodes)
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]

		if current.depth > maxDepth {
			return fmt.Errorf("syntax tree exceeds maximum depth of %d", maxDepth)
		}

		childCount := current.node.NumChildren()
		if childCount > maxNodes-nodes {
			return fmt.Errorf("syntax tree exceeds maximum node count of %d", maxNodes)
		}
		if childCount > 0 && current.depth == maxDepth {
			return fmt.Errorf("syntax tree exceeds maximum depth of %d", maxDepth)
		}

		children := current.node.Children()
		nodes += childCount
		for i := len(children) - 1; i >= 0; i-- {
			stack = append(stack, entry{node: children[i], depth: current.depth + 1})
		}
	}

	return nil
}

func convertNode(node *groovycst.Node, source *groovysource.File) Node {
	converted := Node{
		Kind: node.Kind().String(),
	}

	if node.IsToken() {
		token := node.Token()
		position := source.Position(token.Offset)
		converted.Kind = token.Kind.String()
		converted.Text = token.Text
		converted.Line = position.Line
		converted.Column = position.Column
		return converted
	}

	for _, child := range node.Children() {
		if child.IsToken() && skipToken(child.Token()) {
			continue
		}
		converted.Children = append(converted.Children, convertNode(child, source))
	}
	if len(converted.Children) > 0 {
		converted.Line = converted.Children[0].Line
		converted.Column = converted.Children[0].Column
	} else {
		position := source.Position(0)
		converted.Line = position.Line
		converted.Column = position.Column
	}

	return converted
}

func skipToken(token groovytoken.Token) bool {
	return token.Kind == groovytoken.EOF || token.Text == ""
}
