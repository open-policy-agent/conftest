package groovy

import (
	"encoding/json"
	"strings"
	"testing"

	groovylang "github.com/albertocavalcante/groovy-parser-go"
	groovycst "github.com/albertocavalcante/groovy-parser-go/cst"
	groovyparser "github.com/albertocavalcante/groovy-parser-go/parser"
	groovysource "github.com/albertocavalcante/groovy-parser-go/source"
	"github.com/google/go-cmp/cmp"
)

func TestParser(t *testing.T) {
	t.Parallel()
	sample := `pipeline {
  agent any
  stages {
    stage('Build') {
      steps {
        sh 'make build'
      }
    }
  }
}`

	var got Node
	if err := (&Parser{}).Unmarshal([]byte(sample), &got); err != nil {
		t.Fatal("parse Jenkinsfile:", err)
	}

	if diff := cmp.Diff("FileNode", got.Kind); diff != "" {
		t.Errorf("root kind mismatch (-want +got):\n%s", diff)
	}

	pipeline := findToken(got, "Identifier", "pipeline")
	if pipeline == nil {
		t.Fatal("expected pipeline identifier")
	}
	if pipeline.Line != 1 || pipeline.Column != 1 {
		t.Errorf("pipeline position = %d:%d, want 1:1", pipeline.Line, pipeline.Column)
	}

	shell := findToken(got, "Identifier", "sh")
	if shell == nil {
		t.Fatal("expected shell identifier")
	}
	if shell.Line != 6 || shell.Column != 9 {
		t.Errorf("shell position = %d:%d, want 6:9", shell.Line, shell.Column)
	}

	if findNode(got, "ClosureExpr") == nil {
		t.Error("expected nested closure expressions")
	}
	if findToken(got, "StringLiteral", "'make build'") == nil {
		t.Error("expected command string literal")
	}
}

func TestParserScriptedPipelineAndGString(t *testing.T) {
	t.Parallel()
	sample := `node {
  def target = "${env.BRANCH_NAME}"
  sh "make deploy TARGET=${target}"
}`

	var got Node
	if err := (&Parser{}).Unmarshal([]byte(sample), &got); err != nil {
		t.Fatal("parse scripted pipeline:", err)
	}

	if findToken(got, "Identifier", "node") == nil {
		t.Error("expected node call")
	}
	if findNode(got, "GStringExpr") == nil {
		t.Error("expected interpolated string expression")
	}
	if findToken(got, "Identifier", "sh") == nil {
		t.Error("expected command-style shell call")
	}
}

func TestParserRejectsInvalidSource(t *testing.T) {
	t.Parallel()
	var got any
	err := (&Parser{}).Unmarshal([]byte("pipeline {"), &got)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse groovy at 1:11") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParserOmitsTriviaAndSyntheticTokens(t *testing.T) {
	t.Parallel()
	sample := "// comment\npipeline { agent any }\n"

	var got any
	if err := (&Parser{}).Unmarshal([]byte(sample), &got); err != nil {
		t.Fatal("parse Jenkinsfile:", err)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal("marshal parsed tree:", err)
	}
	text := string(encoded)
	for _, unexpected := range []string{"comment", "EOF", `\"text\":\"\"`} {
		if strings.Contains(text, unexpected) {
			t.Errorf("output contains omitted value %q: %s", unexpected, text)
		}
	}
}

func TestParserJSONContract(t *testing.T) {
	t.Parallel()

	var got Node
	if err := (&Parser{}).Unmarshal([]byte("sh 'echo ok'"), &got); err != nil {
		t.Fatal("parse Jenkinsfile:", err)
	}

	encoded, err := json.Marshal(got)
	if err != nil {
		t.Fatal("marshal parsed tree:", err)
	}

	want := `{"kind":"FileNode","line":1,"column":1,"children":[{"kind":"ExpressionStmt","line":1,"column":1,"children":[{"kind":"CallExpr","line":1,"column":1,"children":[{"kind":"IdentExpr","line":1,"column":1,"children":[{"kind":"Identifier","text":"sh","line":1,"column":1}]},{"kind":"LiteralExpr","line":1,"column":4,"children":[{"kind":"StringLiteral","text":"'echo ok'","line":1,"column":4}]}]}]}]}`
	if diff := cmp.Diff(want, string(encoded)); diff != "" {
		t.Errorf("JSON contract mismatch (-want +got):\n%s", diff)
	}
}

func TestParserRejectsExcessiveNesting(t *testing.T) {
	t.Parallel()

	sample := strings.Repeat("(", maxTreeDepth) + "1" + strings.Repeat(")", maxTreeDepth)
	var got any
	err := (&Parser{}).Unmarshal([]byte(sample), &got)
	if err == nil {
		t.Fatal("expected syntax tree depth error")
	}
	if !strings.Contains(err.Error(), "syntax tree exceeds maximum depth") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateTreeRejectsExcessiveNodes(t *testing.T) {
	t.Parallel()

	tree := parseTree(t, "sh 'echo ok'")
	err := validateTree(tree, maxTreeDepth, 1)
	if err == nil {
		t.Fatal("expected syntax tree node count error")
	}
	if !strings.Contains(err.Error(), "syntax tree exceeds maximum node count of 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func parseTree(t *testing.T, sample string) *groovycst.Node {
	t.Helper()
	source := groovysource.New("", []byte(sample))
	parser := groovyparser.New(source, groovylang.ParseOptions{Version: groovylang.Groovy2_4})
	root, parseErrors := parser.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse syntax tree: %v", parseErrors)
	}
	return groovycst.NewTree(root)
}

func findNode(node Node, kind string) *Node {
	if node.Kind == kind {
		return &node
	}
	for _, child := range node.Children {
		if found := findNode(child, kind); found != nil {
			return found
		}
	}
	return nil
}

func findToken(node Node, kind, text string) *Node {
	if node.Kind == kind && node.Text == text {
		return &node
	}
	for _, child := range node.Children {
		if found := findToken(child, kind, text); found != nil {
			return found
		}
	}
	return nil
}
