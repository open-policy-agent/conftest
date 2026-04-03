package main

import (
	"os"

	"github.com/open-policy-agent/conftest/internal/commands"
	"github.com/open-policy-agent/conftest/plugin"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		if exitErr, ok := plugin.AsExitCodeError(err); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
