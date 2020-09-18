package runner

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
)

type VerifyRunner struct {
	Policy  []string
	Data    []string
	Output  string
	NoColor bool `mapstructure:"no-color"`
	Trace   bool
}

// Run executes the Rego tests at the given PolicyPath(s)
func (r *VerifyRunner) Run(ctx context.Context) ([]output.CheckResult, error) {
	loader := policy.Loader{
		DataPaths:   r.Data,
		PolicyPaths: r.Policy,
	}
	engine, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}

	runner := tester.NewRunner().SetCompiler(engine.Compiler()).SetStore(engine.Store()).SetModules(engine.Modules()).EnableTracing(true).SetRuntime(engine.Runtime())
	ch, err := runner.RunTests(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("running tests: %w", err)
	}

	var results []output.CheckResult
	for result := range ch {
		if result.Error != nil {
			return nil, fmt.Errorf("run test: %w", result.Error)
		}

		buf := new(bytes.Buffer)
		topdown.PrettyTrace(buf, result.Trace)
		var traces []string
		for _, line := range strings.Split(buf.String(), "\n") {
			if len(line) > 0 {
				traces = append(traces, line)
			}
		}

		var outputResult output.Result
		if result.Fail {
			outputResult.Message = result.Package + "." + result.Name
		}

		queryResult := output.QueryResult{
			Query:   result.Name,
			Results: []output.Result{outputResult},
			Traces:  traces,
		}

		checkResult := output.CheckResult{
			FileName: result.Location.File,
			Queries:  []output.QueryResult{queryResult},
		}
		if result.Fail {
			checkResult.Failures = []output.Result{outputResult}
		} else {
			checkResult.Successes++
		}

		results = append(results, checkResult)
	}

	return results, nil
}
