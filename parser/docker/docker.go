package docker

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/moby/buildkit/frontend/dockerfile/instructions"
)

// Parser is a Dockerfile parser.
type Parser struct{}

// Command represents a command in a Dockerfile.
type Command struct {
	// lowercased command name (ex: `from`)
	Cmd    string   
	// for ONBUILD only this holds the sub-command
	SubCmd string   
	// whether the value is written in json form
	JSON   bool     	
	// Any flags such as `--from=...` for `COPY`.
	Flags  []string
	// The contents of the command (ex: `ubuntu:xenial`)
	Value  []string
	// Stage indicates which stage the command is found in a multistage docker build
	Stage int
}

// Unmarshal unmarshals Dockerfiles
func (dp *Parser) Unmarshal(p []byte, v interface{}) error {
	r := bytes.NewReader(p)
	res, err := parser.Parse(r)
	if err != nil {
		return fmt.Errorf("parse dockerfile: %w", err)
	}

	var commands []Command
	var stages []*instructions.Stage
	for _, child := range res.AST.Children {
		instr, err := instructions.ParseInstruction(child)
		if err != nil {
			return fmt.Errorf("process dockerfile instructions: %w", err)
		}

		stage, ok := instr.(*instructions.Stage)
		if ok {
			stages = append(stages, stage)
		}

		cmd := Command{
			Cmd:   child.Value,
			Flags: child.Flags,
			Stage: currentStage(stages),
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

// Return the index of the stages. If no stages are present,
// we set the index to zero.
func currentStage(stages []*instructions.Stage) int {
	if len(stages) == 0 {
		return 0
	}

	return len(stages) - 1
}
