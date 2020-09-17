package plugin

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	executable, arguments, err := parseCommand("plugin.sh arg1")
	if err != nil {
		t.Fatal("parse command:", err)
	}

	if executable != "plugin.sh" {
		t.Errorf("Unexpected executable. expected %v, actual %v", "plugin.sh", executable)
	}

	if arguments[0] != "arg1" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg1", arguments[0])
	}
}

func TestParseCommand_WindowsShell(t *testing.T) {
	executable, arguments, err := parseWindowsCommand("plugin.sh arg1")
	if err != nil {
		t.Fatal("parse command:", err)
	}

	if executable != "sh" {
		t.Errorf("Unexpected executable. expected %v, actual %v", "sh", executable)
	}

	if arguments[0] != "plugin.sh" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "plugin.sh", arguments[0])
	}

	if arguments[1] != "arg1" {
		t.Errorf("Unexpected argument. expected %v, actual %v", "arg1", arguments[1])
	}
}
