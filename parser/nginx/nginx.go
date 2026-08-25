package nginx

import (
	"fmt"

	gonginxconfig "github.com/tufanbarisyildirim/gonginx/config"
	gonginxparser "github.com/tufanbarisyildirim/gonginx/parser"

	"github.com/open-policy-agent/conftest/parser/internal/convert"
)

// Config represents a parsed nginx configuration.
type Config struct {
	Directives []Directive `json:"directives"`
}

// Directive represents a single nginx directive.
type Directive struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters,omitempty"`
	Block      *Block   `json:"block,omitempty"`
}

// Block represents a block of directives.
type Block struct {
	Directives []Directive `json:"directives"`
}

// Parser is an nginx configuration parser.
type Parser struct{}

// Unmarshal unmarshals nginx configuration files.
func (p *Parser) Unmarshal(b []byte, v any) error {
	pr := gonginxparser.NewStringParser(string(b), gonginxparser.WithSkipComments())
	cfg, err := pr.Parse()
	if err != nil {
		return fmt.Errorf("parse nginx: %w", err)
	}

	config := Config{Directives: convertDirectives(cfg.GetDirectives())}
	return convert.Remarshal("nginx", config, v)
}

func convertDirectives(directives []gonginxconfig.IDirective) []Directive {
	converted := make([]Directive, 0, len(directives))
	for _, d := range directives {
		converted = append(converted, convertDirective(d))
	}
	return converted
}

func convertDirective(d gonginxconfig.IDirective) Directive {
	params := make([]string, 0, len(d.GetParameters()))
	for _, p := range d.GetParameters() {
		params = append(params, p.String())
	}

	directive := Directive{
		Name:       d.GetName(),
		Parameters: params,
	}
	if block := d.GetBlock(); block != nil {
		directive.Block = &Block{Directives: convertDirectives(block.GetDirectives())}
	}

	return directive
}
