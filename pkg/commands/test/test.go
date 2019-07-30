package test

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/instrumenta/conftest/pkg/commands/update"
	"github.com/instrumenta/conftest/pkg/constants"
	"github.com/instrumenta/conftest/pkg/parser"

	"github.com/containerd/containerd/log"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	denyQ = regexp.MustCompile("^deny(_[a-zA-Z]+)*$")
	warnQ = regexp.MustCompile("^warn(_[a-zA-Z]+)*$")
)

// checkResult describes the result of a conftest evaluation.
// warning and failure "errors" produced by rego should be considered separate
// from other classes of exceptions.
type checkResult struct {
	warnings []error
	failures []error
}

// NewTestCommand creates a new test command
func NewTestCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "test <file> [file...]",
		Short:   "Test your configuration files using Open Policy Agent",
		Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", constants.Version, constants.Commit, constants.Date),

		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()

			if len(args) < 1 {
				cmd.SilenceErrors = true
				log.G(ctx).Fatal("The first argument should be a file")
			}

			if viper.GetBool("update") {
				update.NewUpdateCommand().Run(cmd, args)
			}

			compiler, err := buildCompiler(viper.GetString("policy"))
			if err != nil {
				log.G(ctx).Fatalf("Problem building rego compiler: %s", err)
			}

			out := getOutputManager(viper.GetString("output"), !viper.GetBool("no-color"))

			foundFailures := false
			for _, fileName := range args {

				// run query engine on file
				res, err := processFile(ctx, fileName, compiler)
				if err != nil {
					log.G(ctx).Fatalf("Problem running evaluation: %s", err)
				}

				// record results
				err = out.put(fileName, checkResult{
					warnings: res.warnings,
					failures: res.failures,
				})
				if err != nil {
					log.G(ctx).Fatalf("Problem compiling results: %s", err)
				}

				if len(res.failures) > 0 || (len(res.warnings) > 0 && viper.GetBool("fail-on-warn")) {
					foundFailures = true
				}
			}

			err = out.flush()
			if err != nil {
				log.G(ctx).Fatal(err)
			}

			if foundFailures {
				os.Exit(1)
			}
		},
	}

	cmd.Flags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")
	cmd.Flags().BoolP("update", "", false, "update any policies before running the tests")
	cmd.Flags().StringP("output", "o", "", fmt.Sprintf("output format for conftest results - valid options are: %s", validOutputs()))

	viper.BindPFlag("fail-on-warn", cmd.Flags().Lookup("fail-on-warn"))
	viper.BindPFlag("update", cmd.Flags().Lookup("update"))
	viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	return cmd
}

func buildRego(trace bool, query string, input interface{}, compiler *ast.Compiler) (*rego.Rego, *topdown.BufferTracer) {
	var regoObj *rego.Rego
	var regoFunc []func(r *rego.Rego)
	buf := topdown.NewBufferTracer()

	regoFunc = append(regoFunc, rego.Query(query), rego.Compiler(compiler), rego.Input(input))
	if trace {
		regoFunc = append(regoFunc, rego.Tracer(buf))
	}
	regoObj = rego.New(regoFunc...)

	return regoObj, buf
}

func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytes.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func processFile(ctx context.Context, fileName string, compiler *ast.Compiler) (checkResult, error) {
	var data []byte
	var err error

	if fileName == "-" {
		reader := bufio.NewReader(os.Stdin)
		data, err = ioutil.ReadAll(reader)
	} else {
		filePath, _ := filepath.Abs(fileName)
		data, err = ioutil.ReadFile(filePath)
	}

	if err != nil {
		return checkResult{}, fmt.Errorf("Unable to open file %s: %s", fileName, err)
	}

	linebreak := detectLineBreak(data)
	bits := bytes.Split(data, []byte(linebreak+"---"+linebreak))

	p := parser.GetParser(fileName)

	var failures []error
	var warnings []error

	for _, element := range bits {
		var input interface{}

		// load individual data segments
		err = p.Unmarshal([]byte(element), &input)
		if err != nil {
			return checkResult{}, err
		}

		// run rules over each data segment
		res, err := processData(ctx, input, compiler)
		if err != nil {
			return checkResult{}, err
		}

		// aggregate errors
		failures = append(failures, res.failures...)
		warnings = append(warnings, res.warnings...)
	}

	return checkResult{
		failures: failures,
		warnings: warnings,
	}, nil
}

// finds all queries in the compiler
func getRules(ctx context.Context, re *regexp.Regexp, compiler *ast.Compiler) []string {

	var res []string

	for _, m := range compiler.Modules {
		for _, r := range m.Rules {
			n := r.Head.Name.String()
			if re.MatchString(n) {
				// the same rule names can be used multiple times, but
				// we only want to run the query and report results once
				if !stringInSlice(n, res) {
					res = append(res, n)
				}
			}
		}
	}

	return res
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func makeQuery(rule string) string {
	return fmt.Sprintf("data.%s.%s", viper.GetString("namespace"), rule)
}

func processData(ctx context.Context, input interface{}, compiler *ast.Compiler) (checkResult, error) {
	// collect warnings
	var warnings []error
	for _, r := range getRules(ctx, warnQ, compiler) {
		ws, err := runQuery(ctx, makeQuery(r), input, compiler)
		if err != nil {
			return checkResult{}, err
		}

		warnings = append(warnings, ws...)
	}

	// collect failures
	var failures []error
	for _, r := range getRules(ctx, denyQ, compiler) {
		fs, err := runQuery(ctx, makeQuery(r), input, compiler)
		if err != nil {
			return checkResult{}, err
		}
		failures = append(failures, fs...)
	}

	return checkResult{
		failures: failures,
		warnings: warnings,
	}, nil
}

func runQuery(ctx context.Context, query string, input interface{}, compiler *ast.Compiler) ([]error, error) {
	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}
		return false
	}

	r, stdout := buildRego(viper.GetBool("trace"), query, input, compiler)
	rs, err := r.Eval(ctx)

	if err != nil {
		return nil, fmt.Errorf("Problem evaluating r policy: %s", err)
	}

	topdown.PrettyTrace(os.Stdout, *stdout)

	var errs []error

	for _, r := range rs {
		for _, e := range r.Expressions {
			value := e.Value
			if hasResults(value) {
				for _, v := range value.([]interface{}) {
					errs = append(errs, errors.New(v.(string)))
				}
			}
		}
	}

	return errs, nil
}

func buildCompiler(path string) (*ast.Compiler, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var files []os.FileInfo
	var dirPath string
	if info.IsDir() {
		files, err = ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		dirPath = path
	} else {
		files = []os.FileInfo{info}
		dirPath = filepath.Dir(path)
	}

	modules := map[string]*ast.Module{}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".rego") {
			continue
		}

		out, err := ioutil.ReadFile(dirPath + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		parsed, err := ast.ParseModule(file.Name(), string(out[:]))
		if err != nil {
			return nil, err
		}
		modules[file.Name()] = parsed
	}

	compiler := ast.NewCompiler()
	compiler.Compile(modules)

	if compiler.Failed() {
		return nil, compiler.Errors
	}

	return compiler, nil
}
