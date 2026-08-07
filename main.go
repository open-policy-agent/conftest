package main

import (
	"errors"
	"os"

	"github.com/open-policy-agent/conftest/internal/commands"
	"github.com/open-policy-agent/conftest/plugin"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		// A plugin that failed a test exits 1 or 2. Exit with the status it
		// chose, so that a caller can tell a failing plugin from a passing
		// one.
		var exitErr plugin.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}

		os.Exit(1)
	}
}
