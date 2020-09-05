package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/conftest/policy"
)

type TestRunner struct {
	Trace         bool
	Policy        []string
	Data          []string
	Update        []string
	Ignore        string
	Input         string
	Namespace     []string
	AllNamespaces bool `mapstructure:"all-namespaces"`
	Combine       bool

	engine *policy.Engine
}

var (
	denyQ = regexp.MustCompile("^(deny|violation)(_[a-zA-Z0-9]+)*$")
	warnQ = regexp.MustCompile("^warn(_[a-zA-Z0-9]+)*$")
)

// Run executes the TestRunner, verifying all Rego policies against the given
// list of configuration files.
func (t *TestRunner) Run(ctx context.Context, fileList []string) ([]output.CheckResult, error) {
	files, err := parseFileList(fileList, t.Ignore)
	if err != nil {
		return nil, fmt.Errorf("parse files: %w", err)
	}

	configManager := parser.ConfigManager{}
	configurations, err := configManager.GetConfigurations(ctx, t.Input, files)
	if err != nil {
		return nil, fmt.Errorf("get configurations: %w", err)
	}

	loader := &policy.Loader{
		DataPaths:   t.Data,
		PolicyPaths: t.Policy,
		URLs:        t.Update,
	}

	regoFiles, store, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load failed: %w", err)
	}

	compiler, err := policy.BuildCompiler(regoFiles)
	if err != nil {
		return nil, fmt.Errorf("build compiler: %w", err)
	}

	engine := &policy.Engine{
		Compiler: compiler,
		Store:    store,
		Trace:    t.Trace,
	}
	t.engine = engine

	namespaces := t.Namespace
	if t.AllNamespaces {
		namespaces, err = policy.GetNamespaces(regoFiles, compiler)
		if err != nil {
			return nil, fmt.Errorf("get namespaces: %w", err)
		}
	}

	var results []output.CheckResult
	if t.Combine {
		result, err := t.GetResult(ctx, namespaces, configurations)
		if err != nil {
			return nil, fmt.Errorf("get combined test result: %w", err)
		}

		result.FileName = "Combined"
		results = append(results, result)
		return results, nil
	}

	for fileName, config := range configurations {
		result, err := t.GetResult(ctx, namespaces, config)
		if err != nil {
			return nil, fmt.Errorf("get test result: %w", err)
		}

		result.FileName = fileName
		results = append(results, result)
	}

	return results, nil
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

// GetResult returns the result of testing the structured data against their policies
func (t *TestRunner) GetResult(ctx context.Context, namespaces []string, input interface{}) (output.CheckResult, error) {
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

func (t *TestRunner) runRules(ctx context.Context, namespace string, input interface{}, ruleRegex *regexp.Regexp) ([]output.Result, []output.Result, []output.Result, error) {
	var rules []string
	var numberOfRules int
	for _, module := range t.engine.Compiler.Modules {
		currentNamespace := strings.Replace(module.Package.Path.String(), "data.", "", 1)
		if currentNamespace != namespace {
			continue
		}

		for _, rule := range module.Rules {
			ruleName := rule.Head.Name.String()

			if !ruleRegex.MatchString(ruleName) {
				continue
			}

			numberOfRules++
			if !stringInSlice(ruleName, rules) {
				rules = append(rules, ruleName)
			}
		}
	}

	// When there are multiple configurations to be tested, we multiply the number of rules
	// by how many configurations will be tested to get the total number of tests.
	totalTests := numberOfRules
	if _, ok := input.([]interface{}); ok {
		totalTests = numberOfRules * len(input.([]interface{}))
	}

	var totalErrors []output.Result
	var totalExceptions []output.Result
	var totalSuccesses []output.Result
	for _, rule := range rules {
		var successes []output.Result
		var exceptions []output.Result
		var errors []output.Result
		var err error

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

	for i := len(totalErrors) + len(totalSuccesses) + len(totalExceptions); i < totalTests; i++ {
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

func (t *TestRunner) runMultipleQueries(ctx context.Context, query string, exceptionQuery string, inputs interface{}) ([]output.Result, []output.Result, []output.Result, error) {
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

func (t *TestRunner) filterExceptionsQuery(ctx context.Context, query string, exceptionQuery string, input interface{}) ([]output.Result, []output.Result, []output.Result, error) {
	violations, successes, err := t.engine.Query(ctx, query, input)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("run query: %w", err)
	}

	_, exceptions, err := t.engine.Query(ctx, exceptionQuery, input)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("exception query: %w", err)
	}

	if len(exceptions) > 0 {
		violations = []output.Result{}
	}

	return violations, exceptions, successes, nil
}
