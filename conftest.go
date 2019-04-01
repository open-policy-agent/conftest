package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:     "conftest <file> [file...]",
	Short:   "Test your configuration files using Open Policy Agent",
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.SilenceErrors = true
			fmt.Println("The first argument should be a file")
			os.Exit(1)
		}
		cmd.SilenceUsage = true

		compiler, err := buildCompiler(viper.GetString("policy"))
		if err != nil {
			fmt.Sprintf("Unable to find policies directory: %s", err)
			os.Exit(1)
		}

		foundFailures := false
		for _, fileName := range args {
			fmt.Println(fileName)
			failures, warnings := processFile(fileName, compiler)
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
			os.Exit(1)
		}
	},
}

func printErrors(err error, color aurora.Color) {
	if merr, ok := err.(*multierror.Error); ok {
		for i := range merr.Errors {
			fmt.Println("  ", aurora.Colorize(merr.Errors[i], color))
		}
	} else {
		fmt.Println(err)
	}
}

// detectLineBreak returns the relevant platform specific line ending
func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytes.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func processFile(fileName string, compiler *ast.Compiler) (error, error) {
	filePath, _ := filepath.Abs(fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Unable to open file %s: %s", fileName, err), nil
	}

	linebreak := detectLineBreak(data)
	bits := bytes.Split(data, []byte(linebreak+"---"+linebreak))

	var failuresList *multierror.Error
	var warningsList *multierror.Error
	for _, element := range bits {
		var input interface{}
		err = yaml.Unmarshal([]byte(element), &input)
		if err != nil {
			return fmt.Errorf("Unable to parse YAML from %s: %s", fileName, err), nil
		}
		failures, warnings := processData(input, compiler)
		if failures != nil {
			failuresList = multierror.Append(failuresList, failures)
		}
		if warnings != nil {
			warningsList = multierror.Append(warningsList, warnings)
		}
	}
	return failuresList.ErrorOrNil(), warningsList.ErrorOrNil()
}

func processData(input interface{}, compiler *ast.Compiler) (error, error) {
	failures := makeQuery("data.main.fail", input, compiler)
	warnings := makeQuery("data.main.warn", input, compiler)
	return failures, warnings
}

func makeQuery(query string, input interface{}, compiler *ast.Compiler) error {
	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}
		return false
	}

	rego := rego.New(
		rego.Query(query),
		rego.Compiler(compiler),
		rego.Input(input))

	ctx := context.Background()
	rs, err := rego.Eval(ctx)
	if err != nil {
		return fmt.Errorf("Problem evaluating rego policies: %s", err)
	}

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

func buildCompiler(path string) (*ast.Compiler, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	modules := map[string]*ast.Module{}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".rego") {
			continue
		}

		out, err := ioutil.ReadFile(path + "/" + file.Name())
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
		panic(compiler.Errors)
	}

	return compiler, nil
}

func init() {
	viper.SetEnvPrefix("CONFTEST")
	viper.AutomaticEnv()

	RootCmd.PersistentFlags().StringP("policy", "p", "policy", "directory for Rego policy files")
	RootCmd.PersistentFlags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")

	RootCmd.SetVersionTemplate(`{{.Version}}`)

	viper.BindPFlag("policy", RootCmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("fail-on-warn", RootCmd.PersistentFlags().Lookup("fail-on-warn"))
}
