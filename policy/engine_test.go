package policy

import (
	"context"
	"testing"

	"github.com/open-policy-agent/conftest/parser"
)

func TestException(t *testing.T) {
	ctx := context.Background()

	regoFiles := []string{"../examples/exceptions/policy/policy.rego", "../examples/exceptions/policy/exception.rego"}
	loader := Loader{
		PolicyPaths: regoFiles,
	}

	engine, err := loader.Load(ctx)
	if err != nil {
		t.Fatalf("loading policies: %v", err)
	}

	configFiles := []string{"../examples/exceptions/deployments.yaml"}
	configs, err := parser.ParseConfigurations(configFiles)
	if err != nil {
		t.Fatalf("loading configs: %v", err)
	}

	results, err := engine.Check(ctx, configs, "main")
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 1
	actualFailures := len(results[0].Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 0
	actualSuccesses := len(results[0].Successes)
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v success, expected %v", actualSuccesses, expectedSuccesses)
	}

	const expectedExceptions = 1
	actualExceptions := len(results[0].Exceptions)
	if actualExceptions != expectedExceptions {
		t.Errorf("Multifile yaml test failure. Got %v exceptions, expected %v", actualExceptions, expectedExceptions)
	}
}

func TestMultifileYaml(t *testing.T) {
	ctx := context.Background()

	regoFiles := []string{"../examples/kubernetes/policy/kubernetes.rego", "../examples/kubernetes/policy/deny.rego"}
	loader := Loader{
		PolicyPaths: regoFiles,
	}

	engine, err := loader.Load(ctx)
	if err != nil {
		t.Fatalf("loading policies: %v", err)
	}

	configFiles := []string{"../examples/kubernetes/deployment+service.yaml"}
	configs, err := parser.ParseConfigurations(configFiles)
	if err != nil {
		t.Fatalf("loading configs: %v", err)
	}

	results, err := engine.Check(ctx, configs, "main")
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 2
	actualFailures := len(results[0].Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 2
	actualSuccesses := len(results[0].Successes)
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v success, expected %v", actualSuccesses, expectedSuccesses)
	}
}

func TestDockerfile(t *testing.T) {
	ctx := context.Background()

	regoFiles := []string{"../examples/docker/policy/base.rego"}
	loader := Loader{
		PolicyPaths: regoFiles,
	}

	engine, err := loader.Load(ctx)
	if err != nil {
		t.Fatalf("loading policies: %v", err)
	}

	configFiles := []string{"../examples/docker/Dockerfile"}
	configs, err := parser.ParseConfigurations(configFiles)
	if err != nil {
		t.Fatalf("loading configs: %v", err)
	}

	results, err := engine.Check(ctx, configs, "main")
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 1
	actualFailures := len(results[0].Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Dockerfile test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 0
	actualSuccesses := len(results[0].Successes)
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Dockerfile test failure. Got %v successes, expected %v", actualSuccesses, expectedSuccesses)
	}
}

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
		{"warn_1", true},
		{"warn_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := isWarning(tt.in)
			if tt.exp != res {
				t.Errorf("%s recognized as `warn` query - expected: %v actual: %v", tt.in, tt.exp, res)
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
		{"deny_1", true},
		{"violation_1", true},
		{"deny_x_y_z", true},
		{"violation_x_y_z", true},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			res := isFailure(tt.in)

			if tt.exp != res {
				t.Fatalf("%s recognized as `fail` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}
