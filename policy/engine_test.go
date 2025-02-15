package policy

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
)

func testOptions(t *testing.T) CompilerOptions {
	t.Helper()
	return CompilerOptions{
		Capabilities: ast.CapabilitiesForThisVersion(),
		RegoVersion:  "v0",
	}
}

func TestException(t *testing.T) {
	ctx := context.Background()

	policies := []string{"../examples/exceptions/policy"}
	engine, err := Load(policies, testOptions(t))
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
		engine, err := Load(policies, testOptions(t))
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
		engine, err := Load(policies, testOptions(t))
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
	engine, err := Load(policies, testOptions(t))
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
	engine, err := Load(policies, testOptions(t))
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
			files := fstest.MapFS{
				"policy.rego": &fstest.MapFile{
					Data: []byte("package main\n\n" + tc.body),
				},
			}

			// Explicit conversion needed despite files being fstest.MapFS type
			// to ensure fs.FS interface implementation for loader.WithFS
			fs := fstest.MapFS(files) //nolint:unconvert
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

func TestLoadWithData(t *testing.T) {
	testCases := []struct {
		desc         string
		policyPaths  []string
		dataPaths    []string
		strict       bool
		wantPolicies bool
		wantDocs     bool
		wantErr      bool
	}{
		{
			desc:         "Load both policies and data",
			policyPaths:  []string{"../examples/kubernetes/policy"},
			dataPaths:    []string{"../examples/kubernetes/service.yaml"},
			wantPolicies: true,
			wantDocs:     true,
		},
		{
			desc:         "Load only data",
			dataPaths:    []string{"../examples/kubernetes/service.yaml"},
			wantPolicies: false,
			wantDocs:     true,
		},
		{
			desc:         "Load only policies",
			policyPaths:  []string{"../examples/kubernetes/policy"},
			wantPolicies: true,
			wantDocs:     false,
		},
		{
			desc:        "Invalid policy path",
			policyPaths: []string{"nonexistent/path"},
			dataPaths:   []string{"../examples/kubernetes/service.yaml"},
			wantErr:     true,
		},
		{
			desc:        "Invalid data path",
			policyPaths: []string{"../examples/kubernetes/policy"},
			dataPaths:   []string{"nonexistent/data.yaml"},
			wantErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			opts := CompilerOptions{
				Strict:       tc.strict,
				Capabilities: ast.CapabilitiesForThisVersion(),
				RegoVersion:  "v0",
			}
			engine, err := LoadWithData(tc.policyPaths, tc.dataPaths, opts)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantPolicies {
				if len(engine.Policies()) == 0 {
					t.Error("expected policies to be loaded but got none")
				}
				if len(engine.Modules()) == 0 {
					t.Error("expected modules to be loaded but got none")
				}
				if engine.Compiler() == nil {
					t.Error("expected compiler to be initialized")
				}
			} else {
				if len(engine.Policies()) > 0 {
					t.Error("expected no policies but got some")
				}
			}

			if tc.wantDocs {
				if len(engine.Documents()) == 0 {
					t.Error("expected documents to be loaded but got none")
				}
				if engine.Store() == nil {
					t.Error("expected store to be initialized")
				}
			} else {
				if len(engine.Documents()) > 0 {
					t.Error("expected no documents but got some")
				}
			}
		})
	}
}

func TestLoadCapabilities(t *testing.T) {
	tmpDir := t.TempDir()

	inaccessibleCapabilitiesPath := filepath.Join(tmpDir, "capabilities")
	err := os.WriteFile(inaccessibleCapabilitiesPath, []byte(""), 0o000)
	if err != nil {
		t.Fatalf("failed to write empty policy file: %v", err)
	}

	t.Cleanup(func() {
		err := os.Chmod(inaccessibleCapabilitiesPath, 0o600)
		if err != nil {
			t.Fatalf("failed to restore capabilities file permissions: %v", err)
		}
	})

	invalidCapabilitiesPath := filepath.Join(tmpDir, "invalid-capabilities")
	err = os.WriteFile(invalidCapabilitiesPath, []byte("invalid json"), 0o600)
	if err != nil {
		t.Fatalf("failed to write invalid capabilities file: %v", err)
	}

	tests := []struct {
		desc    string
		path    string
		wantErr bool
	}{
		{
			desc:    "Inaccessible capabilities file",
			path:    inaccessibleCapabilitiesPath,
			wantErr: true,
		},
		{
			desc:    "Invalid capabilities file",
			path:    invalidCapabilitiesPath,
			wantErr: true,
		},
		{
			desc: "No path does not error",
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := LoadCapabilities(tc.path)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("LoadCapabilities(%s) error = %v, want %v", tc.path, gotErr, tc.wantErr)
			}
		})
	}
}

func TestNamespaces(t *testing.T) {
	tests := []struct {
		name     string
		policies map[string][]byte
		want     []string
	}{
		{
			name: "multiple namespaces",
			policies: map[string][]byte{
				"main.rego": []byte(`package main
deny[msg] { msg := "denied" }`),
				"k8s.rego": []byte(`package kubernetes
deny[msg] { msg := "denied" }`),
				"nested.rego": []byte(`package main.sub
deny[msg] { msg := "denied" }`),
				"main_duplicate.rego": []byte(`package main
warn[msg] { msg := "warning" }`),
			},
			want: []string{"main", "kubernetes", "main.sub"},
		},
		{
			name: "single namespace",
			policies: map[string][]byte{
				"main.rego": []byte(`package main
deny[msg] { msg := "denied" }`),
			},
			want: []string{"main"},
		},
		{
			name:     "no policies",
			policies: map[string][]byte{},
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert policies to fstest.MapFS format
			files := make(map[string]*fstest.MapFile)
			for name, data := range tt.policies {
				files[name] = &fstest.MapFile{Data: data}
			}
			// Explicit conversion needed despite files being fstest.MapFS type
			// to ensure fs.FS interface implementation for loader.WithFS
			fs := fstest.MapFS(files) //nolint:unconvert

			l := loader.NewFileLoader().WithFS(fs)

			keys := make([]string, 0, len(tt.policies))
			for k := range tt.policies {
				keys = append(keys, k)
			}
			pols, err := l.All(keys)
			if err != nil {
				t.Fatalf("Load policies: %v", err)
			}

			engine := Engine{
				modules: pols.ParsedModules(),
			}

			got := engine.Namespaces()
			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Namespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryMetadata(t *testing.T) {
	type want struct {
		msg  string
		meta map[string]any
	}

	tests := []struct {
		name   string
		policy []byte
		query  string
		want   []want
	}{
		{
			name: "string return type",
			policy: []byte(`package main
deny[msg] {
	msg := "simple denial"
}`),
			query: "data.main.deny",
			want: []want{
				{
					msg:  "simple denial",
					meta: map[string]any{"query": "data.main.deny"},
				},
			},
		},
		{
			name: "map return type",
			policy: []byte(`package main
violation[result] {
	result := {
		"msg": "violation with metadata",
		"severity": "high"
	}
}`),
			query: "data.main.violation",
			want: []want{
				{
					msg: "violation with metadata",
					meta: map[string]any{
						"query":    "data.main.violation",
						"severity": "high",
					},
				},
			},
		},
		{
			name: "multiple results",
			policy: []byte(`package main
deny[msg] {
	msg := "first denial"
}
deny[msg] {
	msg := "second denial"
}
violation[result] {
	result := {
		"msg": "violation one",
		"severity": "high"
	}
}
violation[result] {
	result := {
		"msg": "violation two",
		"severity": "low"
	}
}`),
			query: "data.main.deny",
			want: []want{
				{
					msg:  "first denial",
					meta: map[string]any{"query": "data.main.deny"},
				},
				{
					msg:  "second denial",
					meta: map[string]any{"query": "data.main.deny"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			files := fstest.MapFS{
				"policy.rego": &fstest.MapFile{
					Data: tt.policy,
				},
			}

			l := loader.NewFileLoader().WithFS(files)
			pols, err := l.All([]string{"policy.rego"})
			if err != nil {
				t.Fatalf("Load policies: %v", err)
			}

			engine := Engine{
				modules:  pols.ParsedModules(),
				compiler: ast.NewCompiler().WithEnablePrintStatements(true),
			}
			engine.compiler.Compile(engine.modules)
			if engine.compiler.Failed() {
				t.Fatalf("Compiler error: %v", engine.compiler.Errors)
			}

			result, err := engine.query(ctx, nil, tt.query)
			if err != nil {
				t.Fatalf("Query error: %v", err)
			}

			if len(result.Results) != len(tt.want) {
				t.Fatalf("got %d results, want %d", len(result.Results), len(tt.want))
			}

			for i, got := range result.Results {
				want := tt.want[i]
				if got.Message != want.msg {
					t.Errorf("result[%d]: got Message=%q, want Message=%q",
						i, got.Message, want.msg)
				}
				if !reflect.DeepEqual(got.Metadata, want.meta) {
					t.Errorf("result[%d]: got Metadata=%v, want Metadata=%v",
						i, got.Metadata, want.meta)
				}
			}
		})
	}
}
