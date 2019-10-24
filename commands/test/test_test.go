package test

import (
	"context"
	"testing"

	"github.com/instrumenta/conftest/parser/docker"
	"github.com/instrumenta/conftest/parser/yaml"
	"github.com/instrumenta/conftest/policy"
	"github.com/spf13/viper"
)

func TestWarnQuery(t *testing.T) {
	tests := []struct {
		in  string
		exp bool
	}{
		{"", false},
		{"warn", true},
		{"warnXYZ", false},
		{"warn_", false},
		{"warn_x", true},
		{"warn_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := warnQ.MatchString(tt.in)

			if tt.exp != res {
				t.Errorf("%s recognized as `warn` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}

func TestCombineConfig(t *testing.T) {
	viper.Set("namespace", "main")
	testTable := []struct {
		name              string
		combineConfigFlag bool
		policyPath        string
		fileList          []string
	}{
		{
			name:              "valid policy with combine=true should namespace the configs into a map (single file)",
			combineConfigFlag: true,
			policyPath:        "testdata/policy/test_policy_multifile.rego",
			fileList:          []string{"testdata/deployment.yaml"},
		},
		{
			name:              "config combine=false no namespacing, individual evaluation (single file)",
			combineConfigFlag: false,
			policyPath:        "testdata/policy/test_policy.rego",
			fileList:          []string{"testdata/deployment.yaml"},
		},
		{
			name:              "config combine=false no namespacing, individual evaluation (multi-file)",
			combineConfigFlag: false,
			policyPath:        "testdata/policy/test_policy.rego",
			fileList:          []string{"testdata/deployment+service.yaml", "testdata/deployment.yaml"},
		},
		{
			name:              "valid policy with combine=true should namespace the configs into a map (multi-file)",
			combineConfigFlag: true,
			policyPath:        "testdata/policy/test_policy_multifile.rego",
			fileList:          []string{"testdata/deployment+service.yaml", "testdata/deployment.yaml"},
		},
	}

	for _, testunit := range testTable {
		t.Run(testunit.name, func(t *testing.T) {
			viper.Set(combineConfigFlagName, testunit.combineConfigFlag)
			viper.Set("policy", testunit.policyPath)

			ctx := context.Background()
			cmd := NewTestCommand(ctx)
			cmd.Run(cmd, testunit.fileList)
			if outputPrinter.PutCallCount() != len(testunit.fileList) && !testunit.combineConfigFlag {
				t.Errorf(
					"Output manager when combine is false should print output for each file: expected %v calls but got %v",
					len(testunit.fileList),
					outputPrinter.PutCallCount(),
				)
			}
			if errorExitCodeFromCall == 0 && testunit.combineConfigFlag {
				t.Errorf(
					"Output manager when combine is true should have failed but it exited with a zero code: %v",
					errorExitCodeFromCall,
				)
			}
		})
	}

	t.Run("combine flag exists", func(t *testing.T) {
		callCount := 0
		cmd := NewTestCommand(func(int) {
			callCount += 1
		}, func() OutputManager {
			return new(FakeOutputManager)
		})
		if cmd.Flag("combine") == nil {
			t.Errorf("combine flag should exist")
		}
	})
}

func TestInputFlag(t *testing.T) {
	testTable := []struct {
		name       string
		fileList   []string
		input      string
		shouldFail bool
	}{
		{
			name:       "when flag exists it should use the flag value",
			input:      "yml",
			fileList:   []string{"testdata/deployment.yaml"},
			shouldFail: false,
		},
		{
			name:       "when flag doesnt exist it should use the file extension",
			input:      "",
			fileList:   []string{"testdata/deployment.yaml"},
			shouldFail: false,
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			viper.Set("policy", "testdata/policy/test_policy.rego")
			viper.Set("input", testUnit.input)

			exitCallCount := 0
			cmd := NewTestCommand(func(int) {
				exitCallCount += 1
			}, func() OutputManager {
				return new(FakeOutputManager)
			})
			cmd.Run(cmd, testUnit.fileList)

			if testUnit.shouldFail == false && exitCallCount >= 1 {
				t.Error("we did not expect to fail here, yet we did")
			}
		})
	}
}
func TestFailQuery(t *testing.T) {
	tests := []struct {
		in  string
		exp bool
	}{
		{"", false},
		{"deny", true},
		{"violation", true},
		{"denyXYZ", false},
		{"violationXYZ", false},
		{"deny_", false},
		{"violation_", false},
		{"deny_x", true},
		{"violation_x", true},
		{"deny_x_y_z", true},
		{"violation_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := denyQ.MatchString(tt.in)

			if tt.exp != res {
				t.Fatalf("%s recognized as `fail` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}

func TestMultifileYaml(t *testing.T) {
	ctx := context.Background()

	config := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
---
apiVersion: v1
kind: Service
metadata:
  name: hello-kubernetes`

	yaml := yaml.Parser{}

	var jsonConfig interface{}
	err := yaml.Unmarshal([]byte(config), &jsonConfig)
	if err != nil {
		t.Fatalf("Could not unmarshal yaml")
	}

	compiler, err := policy.BuildCompiler("testdata/policy/test_policy.rego", false)
	if err != nil {
		t.Fatalf("Could not build rego compiler")
	}

	results, err := GetResult(ctx, jsonConfig, compiler)
	if err != nil {
		t.Fatalf("Could not process policy file")
	}

	const expected = 2
	actual := len(results.Failures)
	if actual != expected {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actual, expected)
	}
}

func TestDockerfile(t *testing.T) {
	ctx := context.Background()

	config := `FROM openjdk:8-jdk-alpine
VOLUME /tmp

ARG DEPENDENCY=target/dependency
COPY ${DEPENDENCY}/BOOT-INF/lib /app/lib
COPY ${DEPENDENCY}/META-INF /app/META-INF
COPY ${DEPENDENCY}/BOOT-INF/classes /app

ENTRYPOINT ["java","-cp","app:app/lib/*","hello.Application"]`

	parser := docker.Parser{}

	var jsonConfig interface{}
	err := parser.Unmarshal([]byte(config), &jsonConfig)
	if err != nil {
		t.Fatalf("Could not unmarshal dockerfile")
	}

	compiler, err := policy.BuildCompiler("testdata/policy/test_policy_dockerfile.rego", false)
	if err != nil {
		t.Fatalf("Could not build rego compiler")
	}

	results, err := GetResult(ctx, jsonConfig, compiler)
	if err != nil {
		t.Fatalf("Could not process policy file")
	}

	const expected = 1
	actual := len(results.Failures)
	if actual != expected {
		t.Errorf("Dockerfile test failure. Got %v failures, expected %v", actual, expected)
	}
}
