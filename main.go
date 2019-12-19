package main

import (
	"os"

	"github.com/instrumenta/conftest/commands"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
