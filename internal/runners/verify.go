package runners

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
)

type VerifyRunner struct {
	Trace        bool
	Policy   string
	Data    []string
}

// Run executes the Rego tests at the given PolicyPath
func (r *VerifyRunner) Run(ctx context.Context) ([]output.CheckResult, error){
	path := r.Policy
	regoFiles, err := policy.ReadFilesWithTests(path)
	if err != nil {
		return nil, fmt.Errorf("read rego test files: %s", err)
	}

	if len(regoFiles) < 1 {
		return nil, fmt.Errorf("no policies found in %s", path)
	}

	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		return nil, fmt.Errorf("build compiler: %w", err)
	}

	dataPaths := r.Data
	store, err := policy.StoreFromDataFiles(dataPaths)
	if err != nil {
		return nil, fmt.Errorf("build store: %w", err)
	}

	runtime := policy.RuntimeTerm()
	runner := tester.NewRunner().SetCompiler(compiler).SetStore(store).SetModules(compiler.Modules).EnableTracing(r.Trace).SetRuntime(runtime)
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var results []output.CheckResult
	for result := range ch {
		msg := fmt.Errorf("%s", result.Package+"."+result.Name)

		var failure []output.Result
		var success []output.Result

		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)
		var traces []error
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, errors.New(line))
			}
		}

		if result.Fail {
			failure = append(failure, output.NewResult(msg.Error(), traces))
		} else {
			success = append(success, output.NewResult(msg.Error(), traces))
		}

		checkResult := output.CheckResult{
			FileName:  result.Location.File,
			Successes: success,
			Failures:  failure,
		}

		results = append(results, checkResult)
	}

	return results, nil
}