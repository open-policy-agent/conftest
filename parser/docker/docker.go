package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

type Parser struct{}

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
		return fmt.Errorf("parse dockerfile: %w", err)
	}

	// Code attributed to https://github.com/asottile/dockerfile
	// TODO: Just import the package
	var commands []Command
	for _, child := range res.AST.Children {
		cmd := Command{
			Cmd:   child.Value,
			Flags: child.Flags,
		}

		if child.Next != nil && len(child.Next.Children) > 0 {
			cmd.SubCmd = child.Next.Children[0].Value
			child = child.Next.Children[0]
		}

		cmd.JSON = child.Attributes["json"]
		for n := child.Next; n != nil; n = n.Next {
			cmd.Value = append(cmd.Value, n.Value)
		}

		commands = append(commands, cmd)
	}

	var dockerFile [][]Command
	dockerFile = append(dockerFile, commands)

	j, err := json.Marshal(dockerFile)
	if err != nil {
		return fmt.Errorf("marshal dockerfile to json: %w", err)
	}

	if err := json.Unmarshal(j, v); err != nil {
		return fmt.Errorf("unmarshal dockerfile json: %w", err)
	}

	return nil
}
