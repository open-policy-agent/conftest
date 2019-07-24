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
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora"
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

// NewTestCommand creates a new test command
func NewTestCommand(osExit func(int)) *cobra.Command {

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

			foundFailures := false
			for _, fileName := range args {
				if fileName != "-" {
					fmt.Println(fileName)
				}
				failures, warnings := processFile(ctx, fileName, compiler)
				if failures != nil {
					foundFailures = true
					printErrors(failures, aurora.RedFg)
				}
				if warnings != nil {
					if viper.GetBool("fail-on-warn") {
						foundFailures = true
					}
					printErrors(warnings, aurora.BrownFg)
				}
			}
			if foundFailures {
				osExit(1)
			}
		},
	}

	cmd.Flags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")
	cmd.Flags().BoolP("update", "", false, "update any policies before running the tests")
	cmd.Flags().BoolP("combine-files", "", false, "treat files as a single file object")

	viper.BindPFlag("fail-on-warn", cmd.Flags().Lookup("fail-on-warn"))
	viper.BindPFlag("update", cmd.Flags().Lookup("update"))
	viper.BindPFlag("combine-files", cmd.Flags().Lookup("combine-files"))

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

func processFile(ctx context.Context, fileName string, compiler *ast.Compiler) (error, error) {
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
		return fmt.Errorf("Unable to open file %s: %s", fileName, err), nil
	}

	linebreak := detectLineBreak(data)
	bits := bytes.Split(data, []byte(linebreak+"---"+linebreak))

	parser := parser.GetParser(fileName)

	var failuresList *multierror.Error
	var warningsList *multierror.Error
	for _, element := range bits {
		var input interface{}
		err = parser.Unmarshal([]byte(element), &input)
		if err != nil {
			return err, nil
		}
		failures, warnings := processData(ctx, input, compiler)
		if failures != nil {
			failuresList = multierror.Append(failuresList, failures)
		}
		if warnings != nil {
			warningsList = multierror.Append(warningsList, warnings)
		}
	}
	return failuresList.ErrorOrNil(), warningsList.ErrorOrNil()
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

func processData(ctx context.Context, input interface{}, compiler *ast.Compiler) (error, error) {

	// collect warnings
	var warnings error
	for _, r := range getRules(ctx, warnQ, compiler) {
		warnings = multierror.Append(warnings, runQuery(ctx, makeQuery(r), input, compiler))
	}

	// collect failures
	var failures error
	for _, r := range getRules(ctx, denyQ, compiler) {
		failures = multierror.Append(failures, runQuery(ctx, makeQuery(r), input, compiler))
	}

	return failures, warnings
}

func runQuery(ctx context.Context, query string, input interface{}, compiler *ast.Compiler) error {
	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}
		return false
	}

	rego, stdout := buildRego(viper.GetBool("trace"), query, input, compiler)
	rs, err := rego.Eval(ctx)

	if err != nil {
		return fmt.Errorf("Problem evaluating rego policy: %s", err)
	}

	topdown.PrettyTrace(os.Stdout, *stdout)

	var errorsList *multierror.Error

	for _, r := range rs {
		for _, e := range r.Expressions {
			value := e.Value
			if hasResults(value) {
				for _, v := range value.([]interface{}) {
					errorsList = multierror.Append(errorsList, errors.New(v.(string)))
				}
			}
		}
	}

	return errorsList.ErrorOrNil()
}

func getAurora() aurora.Aurora {
	enableColors := !viper.GetBool("no-color")
	return aurora.NewAurora(enableColors)
}

func printErrors(err error, color aurora.Color) {
	aur := getAurora()
	if merr, ok := err.(*multierror.Error); ok {
		for i := range merr.Errors {
			fmt.Println("  ", aur.Colorize(merr.Errors[i], color))
		}
	} else {
		fmt.Println(err)
	}
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
