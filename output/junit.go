package output

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
)

// JUnitOutputManager formats its output as a JUnit test result
type JUnitOutputManager struct {
	p       parser.Package
	writer  io.Writer
	tracing bool
}

// NewDefaultJUnitOutputManager creates a new JUnitOutputManager using standard out
func NewDefaultJUnitOutputManager() *JUnitOutputManager {
	return NewJUnitOutputManager(os.Stdout)
}

// NewJUnitOutputManager creates a new JUnitOutputManager with a given Writer
func NewJUnitOutputManager(w io.Writer) *JUnitOutputManager {
	return &JUnitOutputManager{
		writer: w,
		p: parser.Package{
			Name:  "conftest",
			Tests: []*parser.Test{},
		},
	}
}

// WithTracing adds tracing to the output.
func (j *JUnitOutputManager) WithTracing() OutputManager {
	j.tracing = true
	return j
}

// Put puts the result of the check to the manager in the managers buffer
func (j *JUnitOutputManager) Put(cr CheckResult) error {
	getOutput := func(r Result) []string {
		out := []string{
			r.Message,
		}
		for _, trace := range r.Traces {
			out = append(out, trace)
		}
		return out
	}

	convert := func(r Result, status parser.Result) *parser.Test {
		// We have to make sure that the name of the test is unique
		name := fmt.Sprintf(
			"%s - %s",
			cr.Filename,
			strings.Split(r.Message, "\n")[0],
		)

		return &parser.Test{
			Name:   name,
			Result: status,
			Output: getOutput(r),
		}
	}

	for _, result := range cr.Warnings {
		j.p.Tests = append(j.p.Tests, convert(result, parser.FAIL))
	}

	for _, result := range cr.Failures {
		j.p.Tests = append(j.p.Tests, convert(result, parser.FAIL))
	}

	for i := 0; i < cr.Successes; i++ {
		j.p.Tests = append(j.p.Tests, convert(Result{}, parser.PASS))
	}

	return nil
}

// Flush writes the contents of the managers buffer to the console
func (j *JUnitOutputManager) Flush() error {
	report := &parser.Report{
		Packages: []parser.Package{
			j.p,
		},
	}
	return formatter.JUnitReportXML(report, false, runtime.Version(), j.writer)
}
