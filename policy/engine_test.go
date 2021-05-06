package policy

import (
	"context"
	"testing"

	"github.com/open-policy-agent/conftest/parser"
)

func TestException(t *testing.T) {
	ctx := context.Background()

	policies := []string{"../examples/exceptions/policy"}
	engine, err := Load(ctx, policies)
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
	actualSuccesses := results[0].Successes
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v success, expected %v", actualSuccesses, expectedSuccesses)
	}

	const expectedExceptions = 1
	actualExceptions := len(results[0].Exceptions)
	if actualExceptions != expectedExceptions {
		t.Errorf("Multifile yaml test failure. Got %v exceptions, expected %v", actualExceptions, expectedExceptions)
	}
}

func TestTracing(t *testing.T) {
	t.Run("with tracing ", func(t *testing.T) {
		ctx := context.Background()

		policies := []string{"../examples/kubernetes/policy"}
		engine, err := Load(ctx, policies)
		if err != nil {
			t.Fatalf("loading policies: %v", err)
		}

		engine.EnableTracing()

		configFiles := []string{"../examples/kubernetes/service.yaml"}
		configs, err := parser.ParseConfigurations(configFiles)
		if err != nil {
			t.Fatalf("loading configs: %v", err)
		}

		results, err := engine.Check(ctx, configs, "main")
		if err != nil {
			t.Fatalf("could not process policy file: %s", err)
		}

		for _, query := range results[0].Queries {
			if len(query.Traces) == 0 {
				t.Errorf("Tracing error: Expected trace objects, got 0 instead")
			}
		}
	})

	t.Run("without tracing", func(t *testing.T) {
		ctx := context.Background()

		policies := []string{"../examples/kubernetes/policy"}
		engine, err := Load(ctx, policies)
		if err != nil {
			t.Fatalf("loading policies: %v", err)
		}

		configFiles := []string{"../examples/kubernetes/service.yaml"}
		configs, err := parser.ParseConfigurations(configFiles)
		if err != nil {
			t.Fatalf("loading configs: %v", err)
		}

		results, err := engine.Check(ctx, configs, "main")
		if err != nil {
			t.Fatalf("could not process policy file: %s", err)
		}

		for _, query := range results[0].Queries {
			if len(query.Traces) != 0 {
				t.Errorf("Tracing error: Expected no trace objects, got %d", len(query.Traces))
			}
		}
	})

}

func TestMultifileYaml(t *testing.T) {
	ctx := context.Background()

	policies := []string{"../examples/kubernetes/policy"}
	engine, err := Load(ctx, policies)
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

	const expectedFailures = 4
	actualFailures := len(results[0].Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedWarnings = 1
	actualWarnings := len(results[0].Warnings)
	if actualWarnings != expectedWarnings {
		t.Errorf("Multifile yaml test failure. Got %v warnings, expected %v", actualWarnings, expectedWarnings)
	}

	const expectedSuccesses = 5
	actualSuccesses := results[0].Successes
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v successes, expected %v", actualSuccesses, expectedSuccesses)
	}

	// 10 warnings/failures/successes queries, and 2 dummy exception queries
	const expectedQueries = 12
	actualQueries := len(results[0].Queries)
	if actualQueries != expectedQueries {
		t.Errorf("Multifile yaml test failure. Got %v queries, expected %v", actualQueries, expectedQueries)
	}
}

func TestDockerfile(t *testing.T) {
	ctx := context.Background()

	policies := []string{"../examples/docker/policy"}
	engine, err := Load(ctx, policies)
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
	actualSuccesses := results[0].Successes
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Dockerfile test failure. Got %v successes, expected %v", actualSuccesses, expectedSuccesses)
	}

	// 1 failure, and 1 dummy exception query
	const expectedQueries = 2
	actualQueries := len(results[0].Queries)
	if actualQueries != expectedQueries {
		t.Errorf("Dockerfile test failure. Got %v queries, expected %v", actualQueries, expectedQueries)
	}
}

func TestIsWarning(t *testing.T) {
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

func TestIsFailure(t *testing.T) {
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
