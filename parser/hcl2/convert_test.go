package hcl2

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// This file is mostly attributed to https://github.com/tmccombs/hcl2json

const inputa = `
resource "aws_elastic_beanstalk_environment" "example" {
	name        = "test_environment"
	application = "testing"
  
	setting {
	  namespace = "aws:autoscaling:asg"
	  name      = "MinSize"
	  value     = "1"
	}
  
	dynamic "setting" {
	  for_each = data.consul_key_prefix.environment.var
	  content {
		heredoc = <<-EOF
		This is a heredoc template.
		It references ${local.other.3}
		EOF
		simple = "${4 - 2}"
		cond = test3 > 2 ? 1: 0
		heredoc2 = <<EOF
			Another heredoc, that
			doesn't remove indentation
			${local.other.3}
			%{if true ? false : true}"gotcha"\n%{else}4%{endif}
		EOF
		loop = "This has a for loop: %{for x in local.arr}x,%{endfor}"
		namespace = "aws:elasticbeanstalk:application:environment"
		name      = setting.key
		value     = setting.value
	  }
	}
  }`

const outputa = `{
	"resource": {
		"aws_elastic_beanstalk_environment": {
			"example": {
				"application": "testing",
				"dynamic": {
					"setting": {
						"content": {
							"cond": "${test3 \u003e 2 ? 1: 0}",
							"heredoc": "This is a heredoc template.\nIt references ${local.other.3}\n",
							"heredoc2": "\t\t\tAnother heredoc, that\n\t\t\tdoesn't remove indentation\n\t\t\t${local.other.3}\n\t\t\t%{if true ? false : true}\"gotcha\"\\n%{else}4%{endif}\n",
							"loop": "This has a for loop: %{for x in local.arr}x,%{endfor}",
							"name": "${setting.key}",
							"namespace": "aws:elasticbeanstalk:application:environment",
							"simple": "${4 - 2}",
							"value": "${setting.value}"
						},
						"for_each": "${data.consul_key_prefix.environment.var}"
					}
				},
				"name": "test_environment",
				"setting": {
					"name": "MinSize",
					"namespace": "aws:autoscaling:asg",
					"value": "1"
				}
			}
		}
	}
}`

// Test that conversion works as expected
func TestConversion(t *testing.T) {
	testTable := map[string]struct {
		input  string
		output string
	}{
		"simple": {input: inputa, output: outputa},
	}
	for name, tc := range testTable {
		bytes := []byte(tc.input)
		conf, diags := hclsyntax.ParseConfig(bytes, "test", hcl.Pos{Byte: 0, Line: 1, Column: 1})
		if diags.HasErrors() {
			t.Errorf("Failed to parse config: %v", diags)
		}
		converted, err := convertFile(conf)

		if err != nil {
			t.Errorf("Unable to convert from hcl: %v", err)
		}

		jb, err := json.MarshalIndent(converted, "", "\t")
		if err != nil {
			t.Errorf("Failed to serialize to json: %v", err)
		}
		computedJSON := string(jb)

		if computedJSON != tc.output {
			t.Errorf("For test %s\nExpected:\n%s\n\nGot:\n%s", name, tc.output, computedJSON)
		}
	}
}
