package parse

import (
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/spf13/viper"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestParseConfig(t *testing.T) {
	testTable := []struct {
		name     string
		fileList []string
	}{
		{
			name:     "valid parse for multi-line yamls",
			fileList: []string{"testdata/dummy-deploy.yaml"},
		},
		{
			name:     "valid parse for unstructured inputs(ini)",
			fileList: []string{"testdata/grafana.ini"},
		},
		{
			name:     "valid parse for unstructured inputs(toml)",
			fileList: []string{"testdata/traefik.toml"},
		},
	}

	for _, testunit := range testTable {
		t.Run(testunit.name, func(t *testing.T) {
			exitCallCount := 0
			cmd := NewParseCommand(func(int) {
				exitCallCount++
			})
			cmd.Run(cmd, testunit.fileList)

			if exitCallCount != 1 {
				t.Errorf(
					"It should called one time but we have exit code: %v",
					exitCallCount,
				)
			}
		})
	}
}

func TestInputFlagforParseConfig(t *testing.T) {
	testTable := []struct {
		name     string
		input    string
		fileList []string
	}{
		{
			name:     "valid input flag parse for terraform version 2",
			input:    "hcl2",
			fileList: []string{"testdata/terraform.tf"},
		},
	}

	for _, testunit := range testTable {
		t.Run(testunit.name, func(t *testing.T) {
			viper.Reset()
			viper.Set("input", "hcl2")
			exitCallCount := 0
			expected := `
			"for_each": "${data.consul_key_prefix.environment.var}"`
			output := capturer.CaptureOutput(func() {
				cmd := NewParseCommand(func(int) {
					exitCallCount++
				})

				cmd.Run(cmd, testunit.fileList)
			})
			viper.Reset()
			assert.Assert(t, is.Contains(output, expected))
		})
	}
}

func TestParseOutput(t *testing.T) {
	unit := struct {
		name     string
		fileList []string
	}{
		name:     "valid parse output",
		fileList: []string{"testdata/grafana.ini"},
	}

	expected := `
	"auth.basic": {
		"enabled": "true"
	},
	"server": {
		"domain": "localhost",
		"enable_gzip": "false",
		"enforce_domain": "false",
		"http_addr": "",
		"http_port": "3000",
		"protocol": "http",
		"root_url": "%(protocol)s://%(domain)s:%(http_port)s/",
		"router_logging": "false",
		"serve_from_sub_path": "false",
		"static_root_path": "public"
	},
	`
	t.Run(unit.name, func(t *testing.T) {
		exitCallCount := 0
		output := capturer.CaptureOutput(func() {
			cmd := NewParseCommand(func(int) {
				exitCallCount++
			})
			cmd.Run(cmd, unit.fileList)
		})
		assert.Assert(t, is.Contains(output, expected))

	})
}
