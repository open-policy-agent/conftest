package output

import (
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/v1/tester"
)

// Outputter controls how results of an evaluation will
// be recorded and reported to the end user.
type Outputter interface {
	Output(CheckResults) error
	Report([]*tester.Result, string) error
}

// Options represents the options available when configuring
// an Outputter.
type Options struct {
	Tracing            bool
	NoColor            bool
	SuppressExceptions bool
	ShowSkipped        bool
	JUnitHideMessage   bool
	File               *os.File
}

// The defined output formats represent all of the supported formats
// that can be used to format and render results.
const (
	OutputStandard    = "stdout"
	OutputJSON        = "json"
	OutputTAP         = "tap"
	OutputTable       = "table"
	OutputJUnit       = "junit"
	OutputGitHub      = "github"
	OutputAzureDevOps = "azuredevops"
	OutputSARIF       = "sarif"
)

// Get returns a type that can render output in the given format.
func Get(format string, options Options) Outputter {
	if options.File == nil {
		options.File = os.Stdout
	}

	// If tracing is enabled, output trace to stderr first,
	// then return the requested outputter
	if options.Tracing {
		traceHandler := &Standard{
			Writer:  os.Stderr,
			NoColor: options.NoColor,
			Tracing: true,
		}

		// Return a trace outputter that handles both trace and regular output
		return newTraceOutputter(traceHandler, newOutputter(format, options))
	}

	// If no tracing, return the regular outputter
	return newOutputter(format, options)
}

// newOutputter creates an outputter based on the format and options
func newOutputter(format string, options Options) Outputter {
	switch format {
	case OutputStandard:
		return &Standard{
			Writer:             options.File,
			NoColor:            options.NoColor,
			SuppressExceptions: options.SuppressExceptions,
			Tracing:            options.Tracing,
			ShowSkipped:        options.ShowSkipped,
		}
	case OutputJSON:
		return NewJSON(options.File)
	case OutputTAP:
		return NewTAP(options.File)
	case OutputTable:
		return NewTable(options.File)
	case OutputJUnit:
		return NewJUnit(options.File, options.JUnitHideMessage)
	case OutputGitHub:
		return NewGitHub(options.File)
	case OutputAzureDevOps:
		return NewAzureDevOps(options.File)
	case OutputSARIF:
		return NewSARIF(options.File)
	default:
		return NewStandard(options.File)
	}
}

// traceOutputter handles outputting trace to stderr while sending regular output to stdout
type traceOutputter struct {
	traceHandler  *Standard
	mainOutputter Outputter
}

// newTraceOutputter creates a new traceOutputter with the given trace handler and main outputter
func newTraceOutputter(traceHandler *Standard, mainOutputter Outputter) *traceOutputter {
	return &traceOutputter{
		traceHandler:  traceHandler,
		mainOutputter: mainOutputter,
	}
}

// Output outputs the results, handling trace separately
func (t *traceOutputter) Output(results CheckResults) error {
	// First, output trace to stderr
	if err := t.traceHandler.outputTraceOnly(results); err != nil {
		return err
	}

	// Then, output regular format to stdout
	return t.mainOutputter.Output(results)
}

// Report passes through to the main outputter
func (t *traceOutputter) Report(results []*tester.Result, flag string) error {
	return t.mainOutputter.Report(results, flag)
}

// Outputs returns the available output formats.
func Outputs() []string {
	return []string{
		OutputStandard,
		OutputJSON,
		OutputTAP,
		OutputTable,
		OutputJUnit,
		OutputGitHub,
		OutputAzureDevOps,
		OutputSARIF,
	}
}

func plural(msg string, n int) string {
	if n != 1 {
		return msg + "s"
	}
	return msg
}

func relPath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return rel
}
