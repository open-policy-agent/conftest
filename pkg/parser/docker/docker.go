package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type Parser struct {
	FileName string
}

type Command struct {
	Cmd    string   // lowercased command name (ex: `from`)
	SubCmd string   // for ONBUILD only this holds the sub-command
	JSON   bool     // whether the value is written in json form
	Flags  []string // Any flags such as `--from=...` for `COPY`.
	Value  []string // The contents of the command (ex: `ubuntu:xenial`)
}

func (dp *Parser) Unmarshal(p []byte, v interface{}) error {
	r := bytes.NewReader(p)
	res, err := parser.Parse(r)
	if err != nil {
		return fmt.Errorf("Unable to parse Dockerfile from %s: %s", dp.FileName, err)
	}

	// Code attributed to https://github.com/asottile/dockerfile
	var ret []Command
	for _, child := range res.AST.Children {
		cmd := Command{
			Cmd:   child.Value,
			Flags: child.Flags,
		}

		// Only happens for ONBUILD
		if child.Next != nil && len(child.Next.Children) > 0 {
			cmd.SubCmd = child.Next.Children[0].Value
			child = child.Next.Children[0]
		}

		cmd.JSON = child.Attributes["json"]
		for n := child.Next; n != nil; n = n.Next {
			cmd.Value = append(cmd.Value, n.Value)
		}

		ret = append(ret, cmd)
	}

	j, err := json.Marshal(ret)
	if err != nil {
		return fmt.Errorf("Unable to marshal config %s: %s", dp.FileName, err)
	}

	err = yaml.Unmarshal(j, v)
	if err != nil {
		return fmt.Errorf("Unable to parse YAML from Docker-json %s: %s", dp.FileName, err)
	}

	return nil
}
