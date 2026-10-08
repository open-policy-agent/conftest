package output

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/jstemmer/go-junit-report/formatter"
	"github.com/jstemmer/go-junit-report/parser"
	"github.com/open-policy-agent/opa/v1/tester"
)

// JUnit represents an Outputter that outputs
// results in JUnit format.
type JUnit struct {
	Writer      io.Writer
	hideMessage bool
}

// NewJUnit creates a new JUnit with the given writer.
func NewJUnit(w io.Writer, hideMessage bool) *JUnit {
	jUnit := JUnit{
		Writer:      w,
		hideMessage: hideMessage,
	}

	return &jUnit
}

// Output outputs the results.
func (j *JUnit) Output(results CheckResults) error {
	namespaceTests := make(map[string][]*parser.Test)
	for _, result := range results {
		ns := result.Namespace
		for _, warning := range result.Warnings {
			warningTest := parser.Test{
				Name:   j.formatTestName(result.FileName, warning.Message),
				Result: parser.FAIL,
				Output: strings.Split(warning.Message, "\n"),
			}

			namespaceTests[ns] = append(namespaceTests[ns], &warningTest)
		}

		for _, failure := range result.Failures {
			failingTest := parser.Test{
				Name:   j.formatTestName(result.FileName, failure.Message),
				Result: parser.FAIL,
				Output: strings.Split(failure.Message, "\n"),
			}

			namespaceTests[ns] = append(namespaceTests[ns], &failingTest)
		}

		for _, skipped := range result.Skipped {
			skippedTest := parser.Test{
				Name:   j.formatTestName(result.FileName, skipped.Message),
				Result: parser.SKIP,
				Output: strings.Split(skipped.Message, "\n"),
			}

			namespaceTests[ns] = append(namespaceTests[ns], &skippedTest)
		}

		for s := 0; s < result.Successes; s++ {
			successfulTest := parser.Test{
				Name:   j.formatTestName(result.FileName, ""),
				Result: parser.PASS,
			}

			namespaceTests[ns] = append(namespaceTests[ns], &successfulTest)
		}
	}

	var report parser.Report
	for ns, tests := range namespaceTests {
		report.Packages = append(report.Packages, parser.Package{
			Name:  "conftest." + ns,
			Tests: tests,
		})
	}

	if err := formatter.JUnitReportXML(&report, false, runtime.Version(), j.Writer); err != nil {
		return fmt.Errorf("format junit: %w", err)
	}

	return nil
}

func (j JUnit) formatTestName(fileName, message string) string {
	if j.hideMessage || message == "" {
		return fileName
	}
	summary := strings.Split(message, "\n")[0]
	return fmt.Sprintf("%s - %s", fileName, summary)
}

func (j *JUnit) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in JUnit output")
}
