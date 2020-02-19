package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/instrumenta/conftest/parser"
	"github.com/instrumenta/conftest/policy"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const testDesc = `
This command tests your configuration files using the Open Policy Agent.

The test command expects a one or more input files that will be evaluated against
Open Policy Agent policies. Policies are written in the Rego language. For more
information on how to write Rego policies, see the documentation:
https://www.openpolicyagent.org/docs/latest/policy-language/

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest test --policy <my-directory> <input-file>

Some policies are dependant on external data. This data is loaded in seperatly 
from policies. The location of any data directory or file can be specified with 
the '--data' flag. If a directory is specified, it will be recursively searched for 
any data files. Right now any '.json' or '.yaml' file will be loaded in 
and made available in the Rego policies. Data will be made available in Rego based on 
the file path where the data was found. For example, if data is stored 
under 'policy/exceptions/my_data.yaml', and we execute the following command:

	$ conftest test --data policy <input-file>

The data is available under 'import data.exceptions'.

The test command supports the '--output' flag to specify the type, e.g.:

	$ conftest test -o table -p examples/kubernetes/policy examples/kubernetes/deployment.yaml

Which will return the following output:
+---------+----------------------------------+--------------------------------+
| RESULT  |               FILE               |            MESSAGE             |
+---------+----------------------------------+--------------------------------+
| success | examples/kubernetes/service.yaml |                                |
| warning | examples/kubernetes/service.yaml | Found service hello-kubernetes |
|         |                                  | but services are not allowed   |
+---------+----------------------------------+--------------------------------+

By default, it will use the regular stdout output. For a full list of available output types, see the of the '--output' flag.

The test command supports the '--update' flag to fetch the latest version of the policy at the given url.
It expects one or more urls to fetch the latest policies from, e.g.:

	$ conftest test --update instrumenta.azurecr.io/test

See the pull command for more details on supported protocols for fetching policies.

When debugging policies it can be useful to use a more verbose policy evaluation output. By using the '--trace' flag
the output will include a detailed trace of how the policy was evaluated, e.g.

	$ conftest test --trace <input-file>
`

var (
	denyQ                 = regexp.MustCompile("^(deny|violation)(_[a-zA-Z0-9]+)*$")
	warnQ                 = regexp.MustCompile("^warn(_[a-zA-Z0-9]+)*$")
	combineConfigFlagName = "combine"
)

// Result describes the result of a single rule evaluation.
type Result struct {
	Info   map[string]interface{}
	Traces []error
}

func (r Result) Error() string {
	return r.Info["msg"].(string)
}

// CheckResult describes the result of a conftest evaluation.
// warning and failure "errors" produced by rego should be considered separate
// from other classes of exceptions.
type CheckResult struct {
	FileName  string
	Warnings  []Result
	Failures  []Result
	Successes []Result
}

func NewResult(message string, traces []error) Result {
	result := Result{
		Info:   map[string]interface{}{"msg": message},
		Traces: traces,
	}

	return result
}

// NewTestCommand creates a new test command
func NewTestCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "test <file> [file...]",
		Short: "Test your configuration files using Open Policy Agent",
		Long:  testDesc,
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"fail-on-warn", "update", combineConfigFlagName, "trace", "output", "input", "namespace", "data"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, fileList []string) error {
			out := GetOutputManager()

			// Remove any blank files from the array
			var nonBlankFileList []string
			for _, name := range fileList {
				if name != "" {
					nonBlankFileList = append(nonBlankFileList, name)
				}
			}

			if len(nonBlankFileList) < 1 {
				return fmt.Errorf("no file specified")
			}

			configurations, err := parser.GetConfigurations(ctx, viper.GetString("input"), nonBlankFileList)
			if err != nil {
				return fmt.Errorf("get configurations: %w", err)
			}

			policyPath := viper.GetString("policy")
			urls := viper.GetStringSlice("update")
			for _, url := range urls {
				sourcedURL, err := policy.Detect(url, policyPath)
				if err != nil {
					return fmt.Errorf("detect policies: %w", err)
				}

				if err = policy.Download(ctx, policyPath, []string{sourcedURL}); err != nil {
					return fmt.Errorf("update policies: %w", err)
				}
			}

			regoFiles, err := policy.ReadFiles(policyPath)
			if err != nil {
				return fmt.Errorf("read rego files: %w", err)
			}

			compiler, err := policy.BuildCompiler(regoFiles)
			if err != nil {
				return fmt.Errorf("build compiler: %w", err)
			}

			dataPaths := viper.GetStringSlice("data")
			store, err := policy.StoreFromDataFiles(dataPaths)
			if err != nil {
				return fmt.Errorf("build store: %w", err)
			}

			namespace := viper.GetString("namespace")

			var failures int
			if viper.GetBool(combineConfigFlagName) {
				result, err := GetResult(ctx, namespace, configurations, compiler, store)
				if err != nil {
					return fmt.Errorf("get combined test result: %w", err)
				}

				if isResultFailure(result) {
					failures++
				}

				result.FileName = "Combined"
				if err := out.Put(result); err != nil {
					return fmt.Errorf("writing combined error: %w", err)
				}
			} else {
				for fileName, config := range configurations {
					result, err := GetResult(ctx, namespace, config, compiler, store)
					if err != nil {
						return fmt.Errorf("get test result: %w", err)
					}

					if isResultFailure(result) {
						failures++
					}

					result.FileName = fileName
					if err := out.Put(result); err != nil {
						return fmt.Errorf("writing error: %w", err)
					}
				}
			}

			if err := out.Flush(); err != nil {
				return fmt.Errorf("flushing output: %w", err)
			}

			if failures > 0 {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")
	cmd.Flags().BoolP(combineConfigFlagName, "", false, "combine all given config files to be evaluated together")
	cmd.Flags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")

	cmd.Flags().StringSliceP("update", "u", []string{}, "a list of urls can be provided to the update flag, which will download before the tests run")
	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", ValidOutputs()))
	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	cmd.Flags().StringP("namespace", "", "main", "namespace in which to find deny and warn rules")
	cmd.Flags().StringSliceP("data", "d", []string{}, "A list of paths from which data for the rego policies will be recursively loaded")

	return &cmd
}

// GetResult returns the result of testing the structured data against their policies
func GetResult(ctx context.Context, namespace string, input interface{}, compiler *ast.Compiler, store storage.Store) (CheckResult, error) {
	var totalSuccesses []Result
	warnings, successes, err := runRules(ctx, namespace, input, warnQ, compiler, store)
	if err != nil {
		return CheckResult{}, err
	}
	totalSuccesses = append(totalSuccesses, successes...)

	failures, successes, err := runRules(ctx, namespace, input, denyQ, compiler, store)
	if err != nil {
		return CheckResult{}, err
	}
	totalSuccesses = append(totalSuccesses, successes...)

	result := CheckResult{
		Warnings:  warnings,
		Failures:  failures,
		Successes: totalSuccesses,
	}

	return result, nil
}

func isResultFailure(result CheckResult) bool {
	return len(result.Failures) > 0 || (len(result.Warnings) > 0 && viper.GetBool("fail-on-warn"))
}

func runRules(ctx context.Context, namespace string, input interface{}, regex *regexp.Regexp, compiler *ast.Compiler, store storage.Store) ([]Result, []Result, error) {
	var totalErrors []Result
	var totalSuccesses []Result
	var successes []Result
	var errors []Result
	var err error

	var rules []string
	if regex == nil {
		rules = getRules(ctx, denyQ, compiler)
		rules = append(rules, getRules(ctx, warnQ, compiler)...)
	} else {
		rules = getRules(ctx, regex, compiler)
	}

	for _, rule := range rules {
		query := fmt.Sprintf("data.%s.%s", namespace, rule)

		switch input.(type) {
		case []interface{}:
			errors, successes, err = runMultipleQueries(ctx, query, input, compiler, store)
		default:
			errors, successes, err = runQuery(ctx, query, input, compiler, store)
		}

		if err != nil {
			return nil, nil, err
		}

		totalErrors = append(totalErrors, errors...)
		totalSuccesses = append(totalSuccesses, successes...)
	}

	return totalErrors, totalSuccesses, nil
}

func getRules(ctx context.Context, re *regexp.Regexp, compiler *ast.Compiler) []string {
	var rules []string
	for _, module := range compiler.Modules {
		for _, rule := range module.Rules {
			ruleName := rule.Head.Name.String()

			// the same rule names can be used multiple times, but
			// we only want to run the query and report results once
			if re.MatchString(ruleName) && !stringInSlice(ruleName, rules) {
				rules = append(rules, ruleName)
			}
		}
	}

	return rules
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func runMultipleQueries(ctx context.Context, query string, inputs interface{}, compiler *ast.Compiler, store storage.Store) ([]Result, []Result, error) {
	var totalViolations []Result
	var totalSuccesses []Result
	for _, input := range inputs.([]interface{}) {
		violations, successes, err := runQuery(ctx, query, input, compiler, store)
		if err != nil {
			return nil, nil, fmt.Errorf("run query: %w", err)
		}

		totalViolations = append(totalViolations, violations...)
		totalSuccesses = append(totalSuccesses, successes...)
	}

	return totalViolations, totalSuccesses, nil
}

func runQuery(ctx context.Context, query string, input interface{}, compiler *ast.Compiler, store storage.Store) ([]Result, []Result, error) {
	rego, stdout := buildRego(viper.GetBool("trace"), query, input, compiler, store)
	resultSet, err := rego.Eval(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("evaluating policy: %w", err)
	}

	buf := new(bytes.Buffer)
	topdown.PrettyTrace(buf, *stdout)
	var traces []error
	for _, line := range strings.Split(buf.String(), "\n") {
		if len(line) > 0 {
			traces = append(traces, errors.New(line))
		}
	}

	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}

		return false
	}

	var errs []Result
	var successes []Result
	for _, result := range resultSet {
		for _, expression := range result.Expressions {
			value := expression.Value
			if hasResults(value) {
				for _, v := range value.([]interface{}) {
					switch val := v.(type) {
					case string:
						errs = append(errs, NewResult(val, traces))
					case map[string]interface{}:
						if _, ok := val["msg"]; !ok {
							return nil, nil, fmt.Errorf("rule missing msg field: %v", val)
						}
						if _, ok := val["msg"].(string); !ok {
							return nil, nil, fmt.Errorf("msg field must be string: %v", val)
						}

						result := NewResult(val["msg"].(string), traces)
						for k, v := range val {
							result.Info[k] = v
						}
						errs = append(errs, result)
					}
				}
			} else {
				successes = append(successes, NewResult(expression.Text, traces))
			}
		}
	}

	return errs, successes, nil
}

func buildRego(trace bool, query string, input interface{}, compiler *ast.Compiler, store storage.Store) (*rego.Rego, *topdown.BufferTracer) {
	var regoObj *rego.Rego
	var regoFunc []func(r *rego.Rego)
	buf := topdown.NewBufferTracer()

	regoFunc = append(regoFunc, rego.Query(query), rego.Compiler(compiler), rego.Input(input), rego.Store(store))
	if trace {
		regoFunc = append(regoFunc, rego.Tracer(buf))
	}

	regoObj = rego.New(regoFunc...)

	return regoObj, buf
}
