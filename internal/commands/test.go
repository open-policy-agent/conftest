package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const testDesc = `
This command tests your configuration files using the Open Policy Agent.

The test command expects one or more input files that will be evaluated 
against Open Policy Agent policies. Directories are also supported as valid
inputs. 

Policies are written in the Rego language. For more
information on how to write Rego policies, see the documentation:
https://www.openpolicyagent.org/docs/latest/policy-language/

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest test --policy <my-directory> <input-file(s)/input-folder>

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

// TestRun stores the compiler and store for a test run
type TestRun struct {
	Compiler *ast.Compiler
	Store    storage.Store
}

// NewTestCommand creates a new test command
func NewTestCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "test <file> [file...]",
		Short: "Test your configuration files using Open Policy Agent",
		Long:  testDesc,
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			flagNames := []string{"fail-on-warn", "update", combineConfigFlagName, "trace", "output", "input", "namespace", "all-namespaces", "data", "ignore"}
			for _, name := range flagNames {
				if err := viper.BindPFlag(name, cmd.Flags().Lookup(name)); err != nil {
					return fmt.Errorf("bind flag: %w", err)
				}
			}

			return nil
		},

		RunE: func(cmd *cobra.Command, fileList []string) error {
			outputFormat := viper.GetString("output")
			color := !viper.GetBool("no-color")
			out := output.GetOutputManager(outputFormat, color)
			input := viper.GetString("input")

			files, err := parseFileList(fileList, viper.GetString("ignore"))
			if err != nil {
				return fmt.Errorf("parse files: %w", err)
			}

			configManager := parser.ConfigManager{}
			configurations, err := configManager.GetConfigurations(ctx, input, files)
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

				if err := policy.Download(ctx, policyPath, []string{sourcedURL}); err != nil {
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

			testRun := TestRun{
				Compiler: compiler,
				Store:    store,
			}

			var namespaces []string
			if viper.GetBool("all-namespaces") {
				namespaces, err = policy.GetNamespaces(regoFiles, compiler)
				if err != nil {
					return fmt.Errorf("get namespaces: %w", err)
				}
			} else {
				namespaces = viper.GetStringSlice("namespace")
			}

			var failureFound bool
			if viper.GetBool(combineConfigFlagName) {
				result, err := testRun.GetResult(ctx, namespaces, configurations)
				if err != nil {
					return fmt.Errorf("get combined test result: %w", err)
				}

				if output.IsResultFailure(result, viper.GetBool("fail-on-warn")) {
					failureFound = true
				}

				result.FileName = "Combined"
				if err := out.Put(result); err != nil {
					return fmt.Errorf("writing combined error: %w", err)
				}
			} else {
				for fileName, config := range configurations {
					result, err := testRun.GetResult(ctx, namespaces, config)
					if err != nil {
						return fmt.Errorf("get test result: %w", err)
					}

					if output.IsResultFailure(result, viper.GetBool("fail-on-warn")) {
						failureFound = true
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

			if failureFound {
				os.Exit(1)
			}

			return nil
		},
	}

	cmd.Flags().Bool("fail-on-warn", false, "return a non-zero exit code if only warnings are found")
	cmd.Flags().BoolP(combineConfigFlagName, "", false, "combine all given config files to be evaluated together")
	cmd.Flags().Bool("trace", false, "enable more verbose trace output for rego queries")

	cmd.Flags().StringSliceP("update", "u", []string{}, "a list of urls can be provided to the update flag, which will download before the tests run")
	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", output.ValidOutputs()))
	cmd.Flags().StringP("input", "i", "", fmt.Sprintf("input type for given source, especially useful when using conftest with stdin, valid options are: %s", parser.ValidInputs()))
	cmd.Flags().StringSlice("namespace", []string{"main"}, "namespace in which to find deny and warn rules")
	cmd.Flags().Bool("all-namespaces", false, "find deny and warn rules in all namespaces. If set, the flag \"namespace\" is ignored")
	cmd.Flags().StringSliceP("data", "d", []string{}, "A list of paths from which data for the rego policies will be recursively loaded")
	cmd.Flags().String("ignore", "", "a regex pattern which can be used for ignoring some files/directories in a given input. i.e.: ignoring yamls: --ignore=\".*.yaml\"")

	return &cmd
}

// GetResult returns the result of testing the structured data against their policies
func (t TestRun) GetResult(ctx context.Context, namespaces []string, input interface{}) (output.CheckResult, error) {
	var totalWarnings []output.Result
	var totalFailures []output.Result
	var totalExceptions []output.Result
	var totalSuccesses []output.Result

	for _, namespace := range namespaces {
		warnings, warnExceptions, successes, err := t.runRules(ctx, namespace, input, warnQ)
		if err != nil {
			return output.CheckResult{}, fmt.Errorf("running warn rules: %w", err)
		}
		totalSuccesses = append(totalSuccesses, successes...)

		failures, denyExceptions, successes, err := t.runRules(ctx, namespace, input, denyQ)
		if err != nil {
			return output.CheckResult{}, fmt.Errorf("running deny rules: %w", err)
		}
		totalSuccesses = append(totalSuccesses, successes...)

		totalFailures = append(totalFailures, failures...)
		totalWarnings = append(totalWarnings, warnings...)
		totalExceptions = append(totalExceptions, warnExceptions...)
		totalExceptions = append(totalExceptions, denyExceptions...)
	}

	result := output.CheckResult{
		Warnings:   totalWarnings,
		Failures:   totalFailures,
		Exceptions: totalExceptions,
		Successes:  totalSuccesses,
	}

	return result, nil
}

func (t TestRun) runRules(ctx context.Context, namespace string, input interface{}, regex *regexp.Regexp) ([]output.Result, []output.Result, []output.Result, error) {
	var successes []output.Result
	var exceptions []output.Result
	var errors []output.Result

	var rules []string
	var numberRules int = 0
	for _, module := range t.Compiler.Modules {
		currentNamespace := strings.Replace(module.Package.Path.String(), "data.", "", 1)
		if currentNamespace == namespace {
			for _, rule := range module.Rules {
				ruleName := rule.Head.Name.String()

				if regex.MatchString(ruleName) {
					numberRules += 1
					if !stringInSlice(ruleName, rules) {
						rules = append(rules, ruleName)
					}
				}
			}
		}
	}

	var err error
	var totalErrors []output.Result
	var totalExceptions []output.Result
	var totalSuccesses []output.Result
	for _, rule := range rules {
		query := fmt.Sprintf("data.%s.%s", namespace, rule)
		exceptionQuery := fmt.Sprintf("data.%s.exception[_][_] == %q", namespace, removeDenyPrefix(rule))

		switch input.(type) {
		case []interface{}:
			errors, exceptions, successes, err = t.runMultipleQueries(ctx, query, exceptionQuery, input)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("run multiple queries: %w", err)
			}
		default:
			errors, exceptions, successes, err = t.filterExceptionsQuery(ctx, query, exceptionQuery, input)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("run query: %w", err)
			}
		}

		totalErrors = append(totalErrors, errors...)
		totalExceptions = append(totalExceptions, exceptions...)
		totalSuccesses = append(totalSuccesses, successes...)
	}

	for i := len(totalErrors) + len(totalSuccesses) + len(totalExceptions); i < numberRules; i++ {
		totalSuccesses = append(totalSuccesses, output.Result{})
	}

	return totalErrors, totalExceptions, totalSuccesses, nil
}

func removeDenyPrefix(rule string) string {
	if strings.HasPrefix(rule, "deny_") {
		return strings.TrimPrefix(rule, "deny_")
	} else if strings.HasPrefix(rule, "violation_") {
		return strings.TrimPrefix(rule, "violation_")
	}
	return rule
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}

	return false
}

func (t TestRun) runMultipleQueries(ctx context.Context, query string, exceptionQuery string, inputs interface{}) ([]output.Result, []output.Result, []output.Result, error) {
	var totalViolations []output.Result
	var totalExceptions []output.Result
	var totalSuccesses []output.Result
	for _, input := range inputs.([]interface{}) {
		violations, exceptions, successes, err := t.filterExceptionsQuery(ctx, query, exceptionQuery, input)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("run query: %w", err)
		}

		totalExceptions = append(totalExceptions, exceptions...)
		totalViolations = append(totalViolations, violations...)
		totalSuccesses = append(totalSuccesses, successes...)
	}
	return totalViolations, totalExceptions, totalSuccesses, nil
}

func (t TestRun) filterExceptionsQuery(ctx context.Context, query string, exceptionQuery string, input interface{}) ([]output.Result, []output.Result, []output.Result, error) {
	var totalViolations []output.Result
	var totalExceptions []output.Result
	var totalSuccesses []output.Result
	violations, successes, err := t.runQuery(ctx, query, input)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("run query: %w", err)
	}
	_, exceptions, err := t.runQuery(ctx, exceptionQuery, input)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("exception query: %w", err)
	}
	if len(exceptions) > 0 {
		totalExceptions = append(totalExceptions, exceptions...)
	} else {
		totalViolations = append(totalViolations, violations...)
	}
	totalSuccesses = append(totalSuccesses, successes...)

	return totalViolations, totalExceptions, totalSuccesses, nil
}

func (t TestRun) runQuery(ctx context.Context, query string, input interface{}) ([]output.Result, []output.Result, error) {
	rego, stdout := t.buildRego(viper.GetBool("trace"), query, input)
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

	var errs []output.Result
	var successes []output.Result
	for _, result := range resultSet {
		for _, expression := range result.Expressions {
			if !hasResults(expression.Value) {
				successes = append(successes, output.NewResult(expression.Text, traces))
				continue
			}

			for _, v := range expression.Value.([]interface{}) {
				switch val := v.(type) {
				case string:
					errs = append(errs, output.NewResult(val, traces))
				case map[string]interface{}:
					if _, ok := val["msg"]; !ok {
						return nil, nil, fmt.Errorf("rule missing msg field: %v", val)
					}
					if _, ok := val["msg"].(string); !ok {
						return nil, nil, fmt.Errorf("msg field must be string: %v", val)
					}

					result := output.NewResult(val["msg"].(string), traces)
					for k, v := range val {
						if k != "msg" {
							result.Metadata[k] = v
						}

					}
					errs = append(errs, result)
				}
			}
		}
	}

	return errs, successes, nil
}

func (t TestRun) buildRego(trace bool, query string, input interface{}) (*rego.Rego, *topdown.BufferTracer) {
	var regoObj *rego.Rego
	var regoFunc []func(r *rego.Rego)
	buf := topdown.NewBufferTracer()
	runtime := policy.RuntimeTerm()

	regoFunc = append(regoFunc, rego.Query(query), rego.Compiler(t.Compiler), rego.Input(input), rego.Store(t.Store), rego.Runtime(runtime))
	if trace {
		regoFunc = append(regoFunc, rego.Tracer(buf))
	}

	regoObj = rego.New(regoFunc...)

	return regoObj, buf
}

func parseFileList(fileList []string, exceptions string) ([]string, error) {
	var files []string
	for _, file := range fileList {
		if file == "" {
			continue
		}

		if file == "-" {
			files = append(files, "-")
			continue
		}

		fileInfo, err := os.Stat(file)
		if err != nil {
			return nil, fmt.Errorf("get file info: %w", err)
		}

		if fileInfo.IsDir() {
			directoryFiles, err := getFilesFromDirectory(file, exceptions)
			if err != nil {
				return nil, fmt.Errorf("get files from directory: %w", err)
			}

			files = append(files, directoryFiles...)
		} else {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found")
	}

	return files, nil
}

func getFilesFromDirectory(directory string, exceptions string) ([]string, error) {
	var files []string
	regexp, err := regexp.Compile(exceptions)
	if err != nil {
		return nil, fmt.Errorf("given regexp couldn't be parsed :%w", err)
	}

	walk := func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk path: %w", err)
		}

		if exceptions != "" && regexp.MatchString(currentPath) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		for _, input := range parser.ValidInputs() {
			if strings.HasSuffix(info.Name(), input) {
				files = append(files, currentPath)
			}
		}

		return nil
	}

	err = filepath.Walk(directory, walk)
	if err != nil {
		return nil, err
	}

	return files, nil
}
