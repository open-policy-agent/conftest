package report

import (
	"log"
	"os"

	"github.com/logrusorgru/aurora"
)

// StdOutReporter reports messages to stdout 
type StdOutReporter struct {
	logger *log.Logger
	// control color output within the logger
	color aurora.Aurora
}


// NewDefaultStdOutReporter instantiates a new instance of the StdOutReporter
// using the stdout logger.
func NewDefaultStdOutReporter(color bool) *StdOutReporter {
	return NewStdOutReporter(log.New(os.Stdout, "", 0), color)
}

// NewStdOutReporter instantiates a new instance of the StdOutReporter
// using the given logger.
func NewStdOutReporter(logger *log.Logger, color bool) *StdOutReporter {
	return &StdOutReporter{
		logger: logger,
		color:  aurora.NewAurora(color),
	}
}

// Report messages in the following format
// WARN/ERROR - FILENAME - MSG
func (r *StdOutReporter) Report(results <-chan Result) error {
	for result := range results {
		indicator := getIndicatorForFile(result.FileName)

		r.logger.Print(printColorizedLevel(result.Level, r.color), indicator, result.Msg)
	}

	return nil
}

func printColorizedLevel(level Level, color aurora.Aurora) aurora.Value {
	switch level {
	case Warn:
		return color.Colorize("WARN", aurora.YellowFg)
	default:
		return color.Colorize("FAIL", aurora.RedFg)
	}
}