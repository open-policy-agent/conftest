package hcl2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHCL2(t *testing.T) {

	tests := []struct {
		desc  string
		input string
		want  map[string]any
	}{
		{
			desc: "simple-resources",
			input: `
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
}`,
			want: map[string]any{
				"resource": map[string]any{
					"aws_elastic_beanstalk_environment": map[string]any{
						"example": []any{
							map[string]any{
								"name":        "test_environment",
								"application": "testing",
								"setting": []any{
									map[string]any{
										"name":      "MinSize",
										"namespace": "aws:autoscaling:asg",
										"value":     "1",
									},
								},
								"dynamic": map[string]any{
									"setting": []any{
										map[string]any{
											"for_each": "${data.consul_key_prefix.environment.var}",
											"content": []any{
												map[string]any{
													"cond":      "${test3 \u003e 2 ? 1: 0}",
													"heredoc":   "This is a heredoc template.\nIt references ${local.other.3}\n",
													"heredoc2":  "        Another heredoc, that\n        doesn't remove indentation\n        ${local.other.3}\n        %{if true ? false : true}\"gotcha\"\\n%{else}4%{endif}\n",
													"loop":      "This has a for loop: %{for x in local.arr}x,%{endfor}",
													"name":      "${setting.key}",
													"namespace": "aws:elasticbeanstalk:application:environment",
													"simple":    "${4 - 2}",
													"value":     "${setting.value}",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			desc: "single-provider",
			input: `
provider "aws" {
  version = "=2.46.0"
  alias   = "one"
}
`,
			want: map[string]any{
				"provider": map[string]any{
					"aws": []any{
						map[string]any{
							"alias":   "one",
							"version": "=2.46.0",
						},
					},
				},
			},
		},
		{
			desc: "multiple-providers",
			input: `
provider "aws" {
  version = "=2.46.0"
  alias   = "one"
}
provider "aws" {
  version = "=2.47.0"
  alias   = "two"
}
`,
			want: map[string]any{
				"provider": map[string]any{
					"aws": []any{
						map[string]any{
							"alias":   "one",
							"version": "=2.46.0",
						},
						map[string]any{
							"alias":   "two",
							"version": "=2.47.0",
						},
					},
				},
			},
		},
	}

	p := Parser{}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			var got map[string]any
			if err := p.Unmarshal([]byte(tc.input), &got); err != nil {
				t.Fatalf("Unmarshal: unexpected error %v", err)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("HCL2 produced unexpected diff (-want,+got):\n%s", diff)
			}
		})
	}
}
