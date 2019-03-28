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
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/spf13/cobra"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var RootCmd = &cobra.Command{
	Use:     "conftest <file> [file...]",
	Short:   "Test your configuration files using Open Policy Agent",
	Version: "0.1.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cmd.SilenceErrors = true
			return errors.New("The first argument should be a file")
		}
		cmd.SilenceUsage = true

		compiler, err := buildCompiler("policy")
		if err != nil {
			return fmt.Errorf("Unable to find policies directory: %s", err)
		}

		for _, fileName := range args {
			fmt.Println("Processing", fileName)
			err = processFile(fileName, compiler)
			if err != nil {
				fmt.Println("Policy violations found")
				fmt.Println(err)
			} else {
				fmt.Println("No policy violations found")
			}
		}
		return nil
	},
}

// detectLineBreak returns the relevant platform specific line ending
func detectLineBreak(haystack []byte) string {
	windowsLineEnding := bytes.Contains(haystack, []byte("\r\n"))
	if windowsLineEnding && runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func processFile(fileName string, compiler *ast.Compiler) error {
	filePath, _ := filepath.Abs(fileName)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("Unable to open file %s: %s", fileName, err)
	}

	linebreak := detectLineBreak(data)
	bits := bytes.Split(data, []byte(linebreak+"---"+linebreak))

	var errorsList *multierror.Error
	for _, element := range bits {
		var input interface{}
		err = yaml.Unmarshal([]byte(element), &input)
		if err != nil {
			errorsList = multierror.Append(errorsList, fmt.Errorf("Unable to parse YAML from %s: %s", fileName, err))
		}
		err = processData(input, compiler)
		if err != nil {
			errorsList = multierror.Append(errorsList, err)
		}
	}
	return errorsList.ErrorOrNil()
}

func processData(input interface{}, compiler *ast.Compiler) error {

	rego := rego.New(
		rego.Query("data.main.deny"),
		rego.Compiler(compiler),
		rego.Input(input))

	ctx := context.Background()
	rs, err := rego.Eval(ctx)
	if err != nil {
		return fmt.Errorf("Problem evaluating rego policies: %s", err)
	}

	hasResults := func(expression interface{}) bool {
		if v, ok := expression.([]interface{}); ok {
			return len(v) > 0
		}
		return false
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
