package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	auth "github.com/instrumenta/conftest/pkg/auth/docker"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Config struct {
	Policy    string
	Namespace string
	Policies  []Policy
}

type Policy struct {
	Repository string
	Tag        string
}

func main() {
	RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:     "conftest <subcommand>OA",
	Short:   "Test your configuration files using Open Policy Agent",
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
}

var testCmd = &cobra.Command{
	Use:     "test <file> [file...]",
	Short:   "Test your configuration files using Open Policy Agent",
	Version: fmt.Sprintf("Version: %s\nCommit: %s\nDate: %s\n", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		if len(args) < 1 {
			cmd.SilenceErrors = true
			log.G(ctx).Fatal("The first argument should be a file")
		}

		if viper.GetBool("update") {
			updateCmd.Run(cmd, args)
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

	var failuresList *multierror.Error
	var warningsList *multierror.Error
	for _, element := range bits {
		var input interface{}
		err = yaml.Unmarshal([]byte(element), &input)
		if err != nil {
			return fmt.Errorf("Unable to parse YAML from %s: %s", fileName, err), nil
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

func processData(ctx context.Context, input interface{}, compiler *ast.Compiler) (error, error) {
	namespace := viper.GetString("namespace")
	deny := fmt.Sprintf("data.%s.deny", namespace)
	warn := fmt.Sprintf("data.%s.warn", namespace)

	failures := makeQuery(ctx, deny, input, compiler)
	warnings := makeQuery(ctx, warn, input, compiler)
	return failures, warnings
}

func makeQuery(ctx context.Context, query string, input interface{}, compiler *ast.Compiler) error {
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

	rs, err := rego.Eval(ctx)
	if err != nil {
		return fmt.Errorf("Problem evaluating rego policy: %s", err)
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
		return nil, compiler.Errors
	}

	return compiler, nil
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Download policy from registry",
	Long:  `Download latest policy files according to configuration file`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var config Config
		if err := viper.Unmarshal(&config); err != nil {
			log.G(ctx).Fatal(err)
		}
		downloadPolicy(ctx, config.Policies)
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download individual policies",
	Long:  `Download individual policies from a registry`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		policies := []Policy{}
		for _, ref := range args {
			policies = append(policies, Policy{Repository: ref})
		}
		downloadPolicy(ctx, policies)
	},
}

func downloadPolicy(ctx context.Context, policies []Policy) {
	policyDir := filepath.Join(".", viper.GetString("policy"))
	os.MkdirAll(policyDir, os.ModePerm)

	cli, err := auth.NewClient()
	if err != nil {
		log.G(ctx).Warnf("Error loading auth file: %v\n", err)
	}
	resolver, err := cli.Resolver(ctx)
	if err != nil {
		log.G(ctx).Warnf("Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	fileStore := content.NewFileStore(policyDir)
	defer fileStore.Close()

	for _, policy := range policies {
		var ref string
		if strings.Contains(policy.Repository, ":") {
			ref = policy.Repository
		} else if policy.Tag == "" {
			ref = policy.Repository + ":latest"
		} else {
			ref = policy.Repository + ":" + policy.Tag
		}
		log.G(ctx).Infof("Downloading: %s\n", ref)
		_, _, err = oras.Pull(ctx, resolver, ref, fileStore)
		if err != nil {
			log.G(ctx).Fatalf("Downloading policy failed: %v\n", err)
		}
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(testCmd)

	RootCmd.PersistentFlags().StringP("policy", "p", "policy", "directory for Rego policy files")
	RootCmd.PersistentFlags().BoolP("debug", "", false, "enable more verbose log output")

	testCmd.Flags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")
	testCmd.Flags().BoolP("update", "", false, "update any policies before running the tests")
	RootCmd.PersistentFlags().StringP("namespace", "", "main", "namespace in which to find deny and warn rules")

	RootCmd.SetVersionTemplate(`{{.Version}}`)

	viper.BindPFlag("policy", RootCmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))

	viper.BindPFlag("fail-on-warn", testCmd.Flags().Lookup("fail-on-warn"))
	viper.BindPFlag("update", testCmd.Flags().Lookup("update"))
	viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
}

func initConfig() {
	viper.SetEnvPrefix("CONFTEST")
	viper.SetConfigName("conftest")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.ErrorLevel)
	}
}
