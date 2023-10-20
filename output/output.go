package output

import (
	"os"

	"github.com/open-policy-agent/opa/tester"
)

// Outputter controls how results of an evaluation will
// be recorded and reported to the end user.
type Outputter interface {
	Output([]CheckResult) error
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
)

// Get returns a type that can render output in the given format.
func Get(format string, options Options) Outputter {
	if options.File == nil {
		options.File = os.Stdout
	}

	switch format {
	case OutputStandard:
		return &Standard{Writer: options.File, NoColor: options.NoColor, SuppressExceptions: options.SuppressExceptions, Tracing: options.Tracing, ShowSkipped: options.ShowSkipped}
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
	default:
		return NewStandard(options.File)
	}
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
	}
}
