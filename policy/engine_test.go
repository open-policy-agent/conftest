package policy

import (
	"context"
	"testing"

	"github.com/open-policy-agent/conftest/internal/testing/memfs"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
)

func TestException(t *testing.T) {
	ctx := context.Background()

	policies := []string{"../examples/exceptions/policy"}
	compilerOptions, _ := newCompilerOptions(false, "")
	engine, err := Load(policies, compilerOptions)
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
		compilerOptions, _ := newCompilerOptions(false, "")
		engine, err := Load(policies, compilerOptions)
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
		compilerOptions, _ := newCompilerOptions(false, "")
		engine, err := Load(policies, compilerOptions)
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
	compilerOptions, _ := newCompilerOptions(false, "")
	engine, err := Load(policies, compilerOptions)
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
	compilerOptions, _ := newCompilerOptions(false, "")
	engine, err := Load(policies, compilerOptions)
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

func TestAddFileInfo(t *testing.T) {
	tests := []struct {
		desc  string
		input string
		name  string
	}{
		{
			desc:  "SingleFile",
			input: "foobar.txt",
			name:  "foobar.txt",
		},
		{
			desc:  "RelativePath",
			input: "../foobar.txt",
			name:  "foobar.txt",
		},
		{
			desc:  "FullPath",
			input: "/some/dir/foobar.txt",
			name:  "foobar.txt",
		},
	}

	modules := make(map[string]string, 1)
	modules["test.rego"] = `package main

deny[{"msg": msg}] {
	msg := data.conftest.file.name
}
`

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var e Engine
			ctx := context.Background()
			if err := e.addFileInfo(ctx, tt.input); err != nil {
				t.Error(err)
			}

			module, err := ast.ParseModule("test.rego", modules["test.rego"])
			if err != nil {
				t.Error(err)
			}
			e.modules = make(map[string]*ast.Module, 1)
			e.modules["test.rego"] = module
			compiler, err := ast.CompileModules(modules)
			if err != nil {
				t.Error(err)
			}
			e.compiler = compiler

			qr, err := e.query(ctx, nil, "data.main.deny")
			if err != nil {
				t.Error(err)
			}
			if qr.Results[0].Message != tt.name {
				t.Errorf("mismatch. have [%v], want [%v]", qr.Results[0].Message, tt.name)
			}
		})
	}
}

func TestProblematicIf(t *testing.T) {
	testCases := []struct {
		desc    string
		body    string
		wantErr bool
	}{
		{
			desc: "No rules",
			body: "",
		},
		{
			desc: "Bare deny",
			body: "deny { true }\n",
		},
		{
			desc: "Rule not using if statement",
			body: "deny[msg] {\n 1 == 1\nmsg := \"foo\"\n}\n",
		},
		{
			desc: "Unrelated rule using problematic if",
			body: "import future.keywords.if\nunrelated[msg] if {\n 1 == 1\nmsg := \"foo\"\n}\n",
		},
		{
			desc:    "Rule using if without contains",
			body:    "import future.keywords.if\ndeny[msg] if {\n 1 == 1\nmsg := \"foo\"\n}\n",
			wantErr: true,
		},
		{
			desc: "Rule using if with contains",
			body: "import future.keywords.if\nimport future.keywords.contains\ndeny contains msg if {\n 1 == 1\nmsg := \"foo\"\n}\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			files := map[string][]byte{
				"policy.rego": []byte("package main\n\n" + tc.body),
			}
			fs := memfs.New(files)
			l := loader.NewFileLoader().WithFS(fs)

			pols, err := l.All([]string{"policy.rego"})
			if err != nil {
				t.Fatalf("Load policies: %v", err)
			}
			err = problematicIf(pols.ParsedModules())
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("problematicIf = %v, want %v", gotErr, tc.wantErr)
			}
		})
	}
}
