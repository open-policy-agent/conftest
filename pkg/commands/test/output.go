package test

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"log"
	"os"
)

// outputManager controls how results of the `ccheck` evaluation will be recorded
// and reported to the end user.
type outputManager interface {
	put(fileName string, cr checkResult) error
	flush() error
}

// stdOutputManager reports `ccheck` results to stdout.
type stdOutputManager struct {
	logger *log.Logger
	color  aurora.Aurora
}

// newDefaultStdOutputManager instantiates a new instance of stdOutputManager
// using the default logger.
func newDefaultStdOutputManager(color bool) *stdOutputManager {
	return newStdOutputManager(log.New(os.Stdout, "", 0), color)
}

// newStdOutputManager constructs an instance of stdOutputManager given a
// logger instance.
func newStdOutputManager(l *log.Logger, color bool) *stdOutputManager {
	return &stdOutputManager{
		logger: l,
		// control color output within the logger
		color: aurora.NewAurora(color),
	}
}

func (s *stdOutputManager) put(fileName string, cr checkResult) error {
	var indicator string
	if fileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", fileName)
	}

	// print warnings and then print errors
	for _, r := range cr.warnings {
		s.logger.Print(s.color.Colorize("WARN", aurora.YellowFg), indicator, r)
	}

	for _, r := range cr.failures {
		s.logger.Print(s.color.Colorize("FAIL", aurora.RedFg), indicator, r)
	}

	return nil
}

func (s *stdOutputManager) flush() error {
	// no op
	return nil
}
