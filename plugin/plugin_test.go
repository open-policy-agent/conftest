package plugin

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"
)

func TestLoadPlugin(t *testing.T) {
	path := "../testdata/plugins/kubectl"
	plugin, err := LoadPlugin(path)
	if err != nil {
		t.Fatalf("Unexpected error loading plugin: %v", err)
	}

	expected := &MetaData{
		Name:    "kubectl",
		Version: "0.1.0",
		Usage:   "conftest kubectl (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME).",
		Description: `A Conftest plugin for using Kubectl to test objects in Kubernetes using Open Policy Agent.
Usage: conftest kubectl (TYPE[.VERSION][.GROUP] [NAME] | TYPE[.VERSION][.GROUP]/NAME).`,
		Command: Command("$CONFTEST_PLUGIN_DIR/kubectl-conftest.sh"),
	}

	if !reflect.DeepEqual(expected, plugin.MetaData) {
		t.Errorf("Loading plugin failed, expected metadata: %v, got: %v", expected, plugin.MetaData)
	}
}

func TestCommand_Prepare(t *testing.T) {
	tests := []struct {
		name          string
		command       Command
		main          string
		args          []string
		expectedError string
	}{
		{
			"Expect error on empty command",
			Command(""),
			"",
			[]string{},
			"prepare plugin command: no command found",
		},
		{
			"Handle commands without arguments",
			Command("top"),
			"top",
			[]string{},
			"",
		},
		{
			"Handle commands with arguments",
			Command("docker inspect"),
			"docker",
			[]string{"inspect"},
			"",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			main, args, err := tc.command.Prepare()
			if err != nil && err.Error() != tc.expectedError {
				t.Errorf("Unexpected error in prepare command: %v, expected error: %v", err, tc.expectedError)
			}

			if main != tc.main && !reflect.DeepEqual(args, tc.args) {
				t.Errorf("Unexpected arguments in prepare command: got main: %v, args: %v want main: %v, args: %v", main, args, tc.main, tc.args)
			}
		})
	}
}

func TestPlugin_Exec(t *testing.T) {
	tests := []struct {
		name           string
		plugin         *Plugin
		additionalArgs []string
		args           []string
		output         []string
		expectedError  string
	}{
		{
			"Can execute a simple command",
			&Plugin{
				MetaData: &MetaData{
					Command: Command("echo hello"),
				},
			},
			[]string{},
			[]string{},
			[]string{"hello"},
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			stdIn := bytes.NewBufferString(strings.Join(tc.args, " "))
			stdOut := bytes.NewBuffer([]byte{})
			stdErr := bytes.NewBuffer([]byte{})

			tc.plugin.SetStdIn(stdIn).SetStdOut(stdOut).SetStdErr(stdErr)

			err := tc.plugin.Exec(ctx, tc.additionalArgs)
			if err != nil && tc.expectedError != err.Error() {
				t.Errorf("Unexpected error in exec command: %v, expected error: %v", err, tc.expectedError)
				return
			}

			processedStdOut := strings.Split(stdOut.String(), "\n")
			if len(processedStdOut) > 0 {
				processedStdOut = processedStdOut[:len(processedStdOut)-1]
			}

			if !reflect.DeepEqual(tc.output, processedStdOut) {
				t.Errorf("Unexpected output in exec command: %v. expected output: %v", processedStdOut, tc.output)
			}
		})
	}
}
