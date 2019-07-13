package main

import (
	"github.com/instrumenta/conftest/pkg/commands"
	"os"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
