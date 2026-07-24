package main

import (
	"os"

	"github.com/open-policy-agent/conftest/internal/commands"
	"github.com/open-policy-agent/conftest/plugin"
)

func main() {
	if err := commands.NewDefaultCommand().Execute(); err != nil {
		// When a plugin exits non-zero, propagate its exit code rather than
		// always exiting 1, so scripts and CI can rely on it (#741).
		if exitErr, ok := plugin.AsExitCodeError(err); ok {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
