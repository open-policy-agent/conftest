package runner

import (
	"context"
	"testing"

	"github.com/open-policy-agent/conftest/parser/docker"
	"github.com/open-policy-agent/conftest/parser/yaml"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/storage/inmem"
)

func TestException(t *testing.T) {
	ctx := context.Background()

	config := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: cannot-run-as-root
spec:
  template:
    spec:
      containers:
      - name: root-container
        image: nginx
        ports:
        - containerPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: can-run-as-root
spec:
  template:
    spec:
      containers:
      - name: root-container
        image: nginx
        ports:
        - containerPort: 8080`

	yaml := yaml.Parser{}

	var manifests interface{}
	err := yaml.Unmarshal([]byte(config), &manifests)
	if err != nil {
		t.Fatalf("could not unmarshal yaml: %s", err)
	}

	regoFiles := []string{"../../examples/exceptions/policy/policy.rego", "../../examples/exceptions/policy/exception.rego"}
	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		t.Fatalf("could not build rego compiler: %s", err)
	}

	engine := &policy.Engine{
		Compiler: compiler,
		Store:    inmem.New(),
	}

	testRun := TestRunner{
		engine: engine,
	}

	defaultNamespace := []string{"main"}
	results, err := testRun.GetResult(ctx, defaultNamespace, manifests)
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 1
	actualFailures := len(results.Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 0
	actualSuccesses := len(results.Successes)
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v success, expected %v", actualSuccesses, expectedSuccesses)
	}

	const expectedExceptions = 1
	actualExceptions := len(results.Exceptions)
	if actualExceptions != expectedExceptions {
		t.Errorf("Multifile yaml test failure. Got %v exceptions, expected %v", actualExceptions, expectedExceptions)
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
		t.Fatalf("could not unmarshal yaml: %s", err)
	}

	regoFiles := []string{"../../examples/kubernetes/policy/kubernetes.rego", "../../examples/kubernetes/policy/deny.rego"}
	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		t.Fatalf("could not build rego compiler: %s", err)
	}

	engine := &policy.Engine{
		Compiler: compiler,
		Store:    inmem.New(),
	}

	testRun := TestRunner{
		engine: engine,
	}

	defaultNamespace := []string{"main"}
	results, err := testRun.GetResult(ctx, defaultNamespace, jsonConfig)
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 2
	actualFailures := len(results.Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Multifile yaml test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 2
	actualSuccesses := len(results.Successes)
	if actualSuccesses != expectedSuccesses {
		t.Errorf("Multifile yaml test failure. Got %v success, expected %v", actualSuccesses, expectedSuccesses)
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
		t.Fatalf("could not unmarshal dockerfile: %s", err)
	}

	regoFiles := []string{"../../examples/docker/policy/base.rego"}
	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		t.Fatalf("could not build rego compiler: %s", err)
	}

	engine := &policy.Engine{
		Compiler: compiler,
		Store:    inmem.New(),
	}

	testRun := TestRunner{
		engine: engine,
	}

	defaultNamespace := []string{"main"}
	results, err := testRun.GetResult(ctx, defaultNamespace, jsonConfig)
	if err != nil {
		t.Fatalf("could not process policy file: %s", err)
	}

	const expectedFailures = 1
	actualFailures := len(results.Failures)
	if actualFailures != expectedFailures {
		t.Errorf("Dockerfile test failure. Got %v failures, expected %v", actualFailures, expectedFailures)
	}

	const expectedSuccesses = 0
	actualSuccesses := len(results.Successes)
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
			res := warnQ.MatchString(tt.in)

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
			res := denyQ.MatchString(tt.in)

			if tt.exp != res {
				t.Fatalf("%s recognized as `fail` query - expected: %v actual: %v", tt.in, tt.exp, res)
			}
		})
	}
}
