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
	"github.com/hashicorp/go-multierror"
	"github.com/instrumenta/conftest/util"
	"github.com/logrusorgru/aurora"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	auth "github.com/deislabs/oras/pkg/auth/docker"
)

const (
	OpenPolicyAgentConfigMediaType        = "application/vnd.cncf.openpolicyagent.config.v1+json"
	OpenPolicyAgentManifestLayerMediaType = "application/vnd.cncf.openpolicyagent.manifest.layer.v1+json"
	OpenPolicyAgentPolicyLayerMediaType   = "application/vnd.cncf.openpolicyagent.policy.layer.v1+rego"
	OpenPolicyAgentDataLayerMediaType     = "application/vnd.cncf.openpolicyagent.data.layer.v1+json"
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
	Use:     "conftest <subcommand>",
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

func getAurora() aurora.Aurora {
	enableColors := viper.GetBool("no-color")
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

	parser := util.GetParser(fileName)

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
	Use:   "pull <repository>",
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
		if util.RepositoryNameContainsTag(policy.Repository) {
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

var pushCmd = &cobra.Command{
	Use:   "push <repository> [filepath]",
	Short: "Upload OPA bundles to an OCI registry",
	Long:  `Upload Open Policy Agent bundles to an OCI registry`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var path string
		var err error
		if len(args) == 2 {
			path = args[1]
		} else {
			path, err = os.Getwd()
			if err != nil {
				log.G(ctx).Fatal(err)
			}
		}
		uploadBundle(ctx, args[0], path)
	},
}

func buildLayer(ctx context.Context, paths []string, root string, memoryStore *content.Memorystore, mediaType string) []ocispec.Descriptor {
	var layer ocispec.Descriptor
	var layers []ocispec.Descriptor
	for _, file := range paths {
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			log.G(ctx).Fatal(err)
		}
		relative, err := filepath.Rel(root, file)
		if err != nil {
			log.G(ctx).Fatal(err)
		}
		layer = memoryStore.Add(relative, OpenPolicyAgentPolicyLayerMediaType, contents)
		layers = append(layers, layer)
	}
	return layers
}

func buildLayers(ctx context.Context, root string) ([]ocispec.Descriptor, *content.Memorystore) {
	var data []string
	var policy []string
	var layers []ocispec.Descriptor
	var err error

	root, err = filepath.Abs(root)
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	info, err := os.Stat(root)
	if err != nil {
		log.G(ctx).Fatal(err)
	}
	if !info.IsDir() {
		log.G(ctx).Fatalf("%s isn't a directory", root)
	}

	memoryStore := content.NewMemoryStore()

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".rego" {
			policy = append(policy, path)
		}
		if filepath.Ext(path) == ".json" {
			data = append(data, path)
		}
		return nil
	})
	if err != nil {
		log.G(ctx).Fatal(err)
	}

	policyLayers := buildLayer(ctx, policy, root, memoryStore, OpenPolicyAgentPolicyLayerMediaType)
	dataLayers := buildLayer(ctx, data, root, memoryStore, OpenPolicyAgentDataLayerMediaType)
	layers = append(policyLayers, dataLayers...)

	return layers, memoryStore
}

func uploadBundle(ctx context.Context, repository string, root string) {

	cli, err := auth.NewClient()
	if err != nil {
		log.G(ctx).Warnf("Error loading auth file: %v\n", err)
	}
	resolver, err := cli.Resolver(ctx)
	if err != nil {
		log.G(ctx).Warnf("Error loading resolver: %v\n", err)
		resolver = docker.NewResolver(docker.ResolverOptions{})
	}

	var ref string
	if strings.Contains(repository, ":") {
		ref = repository
	} else {
		ref = repository + ":latest"
	}

	layers, memoryStore := buildLayers(ctx, root)

	log.G(ctx).Infof("Pushing bundle to %s\n", ref)
	extraOpts := []oras.PushOpt{oras.WithConfigMediaType(OpenPolicyAgentConfigMediaType)}
	manifest, err := oras.Push(ctx, resolver, ref, memoryStore, layers, extraOpts...)
	if err != nil {
		log.G(ctx).Fatal(err)
	}
	log.G(ctx).Infof("Pushed bundle to %s with digest %s\n", ref, manifest.Digest)
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.AddCommand(updateCmd)
	RootCmd.AddCommand(pullCmd)
	RootCmd.AddCommand(pushCmd)
	RootCmd.AddCommand(testCmd)

	RootCmd.PersistentFlags().StringP("policy", "p", "policy", "path to the Rego policy files directory. For the test command, specifying a specific .rego file is allowed.")
	RootCmd.PersistentFlags().BoolP("debug", "", false, "enable more verbose log output")
	RootCmd.PersistentFlags().BoolP("trace", "", false, "enable more verbose trace output for rego queries")

	testCmd.Flags().BoolP("fail-on-warn", "", false, "return a non-zero exit code if only warnings are found")
	testCmd.Flags().BoolP("update", "", false, "update any policies before running the tests")
	RootCmd.PersistentFlags().StringP("namespace", "", "main", "namespace in which to find deny and warn rules")
	RootCmd.PersistentFlags().BoolP("no-color", "", true, " Disable color when printing;")

	RootCmd.SetVersionTemplate(`{{.Version}}`)

	viper.BindPFlag("policy", RootCmd.PersistentFlags().Lookup("policy"))
	viper.BindPFlag("debug", RootCmd.PersistentFlags().Lookup("debug"))
	viper.BindPFlag("trace", RootCmd.PersistentFlags().Lookup("trace"))

	viper.BindPFlag("fail-on-warn", testCmd.Flags().Lookup("fail-on-warn"))
	viper.BindPFlag("update", testCmd.Flags().Lookup("update"))
	viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("color", RootCmd.PersistentFlags().Lookup("color"))
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
