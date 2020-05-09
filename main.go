package main

import (
	"os"

	"github.com/open-policy-agent/conftest/internal/commands"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
