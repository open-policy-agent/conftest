package output

import (
	"fmt"
	"log"
	"os"
)

// TAPOutputManager formats its output in TAP format
type TAPOutputManager struct {
	logger  *log.Logger
	tracing bool
}

// NewDefaultTAPOutputManager creates a new TAPOutputManager using the default logger
func NewDefaultTAPOutputManager() *TAPOutputManager {
	return NewTAPOutputManager(log.New(os.Stdout, "", 0))
}

// NewTAPOutputManager creates a new TAPOutputManager with a given logger instance
func NewTAPOutputManager(l *log.Logger) *TAPOutputManager {
	return &TAPOutputManager{
		logger: l,
	}
}

// WithTracing adds tracing to the output.
func (t *TAPOutputManager) WithTracing() OutputManager {
	t.tracing = true
	return t
}

// Put puts the result of the check to the manager in the managers buffer
func (t *TAPOutputManager) Put(cr CheckResult) error {
	var indicator string
	if cr.Filename == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", cr.Filename)
	}

	printResults := func(r Result, prefix string, counter int) {
		t.logger.Print(prefix, counter, indicator, r.Message)
		if len(r.Traces) > 0 && t.tracing {
			t.logger.Print("# Traces")
			for j, trace := range r.Traces {
				t.logger.Print("trace ", counter, j+1, indicator, trace)
			}
		}
	}

	issues := len(cr.Failures) + len(cr.Warnings) + cr.Successes
	if issues > 0 {
		t.logger.Print(fmt.Sprintf("1..%d", issues))
		for i, r := range cr.Failures {
			printResults(r, "not ok ", i+1)

		}

		if len(cr.Warnings) > 0 {
			t.logger.Print("# Warnings")
			for i, r := range cr.Warnings {
				counter := i + 1 + len(cr.Failures)
				printResults(r, "not ok ", counter)
			}
		}

		if cr.Successes > 0 {
			t.logger.Print("# Successes")
			for i := 0; i < cr.Successes; i++ {
				counter := i + 1 + len(cr.Failures) + len(cr.Warnings)
				printResults(Result{}, "not ok ", counter)
			}
		}
	}

	return nil
}

// Flush is currently a NOOP
func (t *TAPOutputManager) Flush() error {
	return nil
}
