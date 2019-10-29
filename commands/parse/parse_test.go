package parse

import (
	"context"
	"testing"

	"github.com/spf13/viper"
	"gotest.tools/assert"
	is "gotest.tools/assert/cmp"
)

func TestParseConfig(t *testing.T) {
	ctx := context.Background()
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
			cmd := NewParseCommand(ctx)
			err := cmd.RunE(cmd, testunit.fileList)
			if err != nil {
				t.Errorf("problem running parse command in test: %w", err)
			}
		})
	}
}

func TestInputFlagForparseInput(t *testing.T) {
	ctx := context.Background()
	testunit := struct {
		name     string
		input    string
		fileList []string
	}{
		name:     "valid input flag parse for terraform version 2",
		input:    "hcl2",
		fileList: []string{"testdata/terraform.tf"},
	}
	t.Run(testunit.name, func(t *testing.T) {
		expectedFile := "testdata/terraform.tf"
		expected := `{
	"data.consul_key_prefix.environment": {
		"path": "apps/example/env"
	},
	"output.environment": {
		"value": "${{\n    id           = aws_elastic_beanstalk_environment.example.id\n    vpc_settings = {\n      for s in aws_elastic_beanstalk_environment.example.all_settings :\n      s.name =\u003e s.value\n      if s.namespace == \"aws:ec2:vpc\"\n    }\n  }}"
	},
	"resource.aws_elastic_beanstalk_environment.example": {
		"application": "testing",
		"dynamic.setting": {
			"content": {
				"name": "${setting.key}",
				"namespace": "aws:elasticbeanstalk:application:environment",
				"value": "${setting.value}"
			},
			"for_each": "${data.consul_key_prefix.environment.var}"
		},
		"name": "test_environment",
		"setting": {
			"name": "MinSize",
			"namespace": "aws:autoscaling:asg",
			"value": "1"
		}
	}
}`
		viper.Reset()
		viper.Set("input", testunit.input)
		parsed, _ := parseInput(ctx, testunit.fileList)
		viper.Reset()
		assert.Assert(t, is.Contains(string(parsed), expected))
		assert.Assert(t, is.Contains(string(parsed), expectedFile))
	})
}

func TestParseOutputwithNoFlag(t *testing.T) {
	ctx := context.Background()
	unit := struct {
		name     string
		fileList []string
	}{
		name:     "valid parse output",
		fileList: []string{"testdata/grafana.ini"},
	}
	expectedFile := "testdata/grafana.ini"
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
		parsed, _ := parseInput(ctx, unit.fileList)
		assert.Assert(t, is.Contains(string(parsed), expected))
		assert.Assert(t, is.Contains(string(parsed), expectedFile))

	})
}
