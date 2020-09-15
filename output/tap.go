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

	if t.tracing {
		t.logger.Print("# " + cr.FileName)

		for _, queryResult := range cr.Queries {
			var resultLine string
			if queryResult.Passed() {
				resultLine = "ok " + queryResult.Query
			} else {
				resultLine = "not ok " + queryResult.Query
			}
			t.logger.Print(resultLine)

			for index, trace := range queryResult.Traces {
				t.logger.Print("# ", index, " ", trace)
			}
		}

		return nil
	}

	var indicator string
	if cr.FileName == "-" {
		indicator = " - "
	} else {
		indicator = fmt.Sprintf(" - %s - ", cr.FileName)
	}

	issues := cr.Successes + len(cr.Failures) + len(cr.Warnings) + len(cr.Exceptions)
	if issues == 0 {
		return nil
	}

	t.logger.Print(fmt.Sprintf("1..%d", issues))
	for i, r := range cr.Failures {
		t.logger.Print("not ok ", i+1, indicator, r.Message)

	}

	if len(cr.Warnings) > 0 {
		t.logger.Print("# warnings")
		for i, r := range cr.Warnings {
			counter := i + 1 + len(cr.Failures)
			t.logger.Print("not ok ", counter, indicator, r.Message)
		}
	}

	if len(cr.Exceptions) > 0 {
		t.logger.Print("# exceptions")
		for i, r := range cr.Exceptions {
			counter := i + 1 + len(cr.Failures) + len(cr.Warnings)
			t.logger.Print("ok ", counter, indicator, r.Message)
		}
	}

	if cr.Successes > 0 {
		t.logger.Print("# successes")
		for i := 0; i < cr.Successes; i++ {
			counter := i + 1 + len(cr.Failures) + len(cr.Warnings) + len(cr.Exceptions)
			t.logger.Print("ok ", counter, indicator, "")
		}
	}

	return nil
}

// Flush is currently a NOOP
func (t *TAPOutputManager) Flush() error {
	return nil
}
