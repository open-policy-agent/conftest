package cue

import (
	"fmt"
	"path/filepath"

	"cuelang.org/go/cue"
	cFormat "cuelang.org/go/cue/format"
	cParser "cuelang.org/go/cue/parser"
	"github.com/ghodss/yaml"
)

type Parser struct {
	FileName string
}

func (c *Parser) Unmarshal(p []byte, v interface{}) error {
	var r cue.Runtime
	filePath, _ := filepath.Abs(c.FileName)
	cfg, err := cParser.ParseFile(filePath, nil)
	if err != nil {
		return fmt.Errorf("load cue config failed: %v", err)
	}
	out, err := cFormat.Node(cfg)
	if err != nil {
		return fmt.Errorf("error occured when formatting cue: %v", err)
	}
	instance, err := r.Parse("name", out)
	if err != nil {
		return fmt.Errorf("error occured parsing cue: %v", err)
	}
	j, err := instance.Value().MarshalJSON()
	if err != nil {
		return fmt.Errorf("Unable to marshal cue config %s: %s", c.FileName, err)
	}
	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from cue-json %s: %s", c.FileName, err)
	}
	return nil
}
