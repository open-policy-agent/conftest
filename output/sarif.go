package output

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/open-policy-agent/opa/tester"
	"github.com/owenrumney/go-sarif/v2/sarif"
	"golang.org/x/exp/slices"
)

const (
	// Tool information
	toolName     = "conftest"
	toolURI      = "https://github.com/open-policy-agent/conftest"
	sarifVersion = sarif.Version210

	// Result descriptions
	successDesc   = "Policy was satisfied successfully"
	skippedDesc   = "Policy check was skipped"
	failureDesc   = "Policy violation"
	warningDesc   = "Policy warning"
	exceptionDesc = "Policy exception"

	// Exit code descriptions
	exitNoViolations = "No policy violations found"
	exitViolations   = "Policy violations found"
	exitWarnings     = "Policy warnings found"
	exitExceptions   = "Policy exceptions found"
)

// SARIF represents an Outputter that outputs results in SARIF format.
type SARIF struct {
	writer io.Writer
}

// NewSARIF creates a new SARIF with the given writer.
func NewSARIF(w io.Writer) *SARIF {
	return &SARIF{
		writer: w,
	}
}

// getRuleID generates a stable rule ID based on namespace and rule type
func getRuleID(namespace string, ruleType string) string {
	return fmt.Sprintf("%s/%s", namespace, ruleType)
}

// getRuleDescription returns the appropriate description based on the rule type
func getRuleDescription(ruleID string) string {
	switch {
	case strings.HasSuffix(ruleID, "/success"):
		return successDesc
	case strings.HasSuffix(ruleID, "/skip"):
		return skippedDesc
	case strings.HasSuffix(ruleID, "/allow"):
		return exceptionDesc
	case strings.HasSuffix(ruleID, "/warn"):
		return warningDesc
	default:
		return failureDesc
	}
}

// getRuleIndex returns the index for a rule if it exists in the indices map.
// The bool return indicates if the rule was found.
func getRuleIndex(ruleID string, indices map[string]int) (int, bool) {
	idx, ok := indices[ruleID]
	return idx, ok
}

// addRuleIndex adds a new rule to the SARIF run and returns its index.
func addRuleIndex(run *sarif.Run, ruleID string, result Result, indices map[string]int) int {
	addRule(run, ruleID, result)
	idx := len(run.Tool.Driver.Rules) - 1
	indices[ruleID] = idx
	return idx
}

// addRule adds a new rule to the SARIF run with the given ID and result metadata.
func addRule(run *sarif.Run, ruleID string, result Result) {
	desc := getRuleDescription(ruleID)
	rule := run.AddRule(ruleID).
		WithDescription(desc).
		WithShortDescription(&sarif.MultiformatMessageString{
			Text: &desc,
		})

	if result.Metadata != nil {
		props := sarif.NewPropertyBag()
		for k, v := range result.Metadata {
			props.Add(k, v)
		}
		rule.WithProperties(props.Properties)
	}
}

// Output outputs the results in SARIF format.
func (s *SARIF) Output(results []CheckResult) error {
	report, err := sarif.New(sarifVersion)
	if err != nil {
		return fmt.Errorf("create sarif report: %w", err)
	}

	run := sarif.NewRunWithInformationURI(toolName, toolURI)
	indices := make(map[string]int)

	for _, result := range results {
		// Process failures
		for _, failure := range result.Failures {
			ruleID := getRuleID(result.Namespace, "deny")
			var idx int
			if existingIdx, ok := getRuleIndex(ruleID, indices); ok {
				idx = existingIdx
			} else {
				idx = addRuleIndex(run, ruleID, failure, indices)
			}

			run.CreateResultForRule(ruleID).
				WithRuleIndex(idx).
				WithLevel("error").
				WithMessage(sarif.NewTextMessage(failure.Message)).
				AddLocation(
					sarif.NewLocationWithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation(filepath.ToSlash(result.FileName)),
							),
					),
				)
		}

		// Process warnings
		for _, warning := range result.Warnings {
			ruleID := getRuleID(result.Namespace, "warn")
			var idx int
			if existingIdx, ok := getRuleIndex(ruleID, indices); ok {
				idx = existingIdx
			} else {
				idx = addRuleIndex(run, ruleID, warning, indices)
			}

			run.CreateResultForRule(ruleID).
				WithRuleIndex(idx).
				WithLevel("warning").
				WithMessage(sarif.NewTextMessage(warning.Message)).
				AddLocation(
					sarif.NewLocationWithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation(filepath.ToSlash(result.FileName)),
							),
					),
				)
		}

		// Process exceptions
		for _, exception := range result.Exceptions {
			ruleID := getRuleID(result.Namespace, "allow")
			var idx int
			if existingIdx, ok := getRuleIndex(ruleID, indices); ok {
				idx = existingIdx
			} else {
				idx = addRuleIndex(run, ruleID, exception, indices)
			}

			run.CreateResultForRule(ruleID).
				WithRuleIndex(idx).
				WithLevel("note").
				WithMessage(sarif.NewTextMessage(exception.Message)).
				AddLocation(
					sarif.NewLocationWithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation(filepath.ToSlash(result.FileName)),
							),
					),
				)
		}

		// Add success or skipped result if no other results
		if len(result.Failures) == 0 && len(result.Warnings) == 0 && len(result.Exceptions) == 0 {
			text := successDesc
			ruleType := "success"
			if result.Successes == 0 {
				text = skippedDesc
				ruleType = "skip"
			}

			emptyResult := Result{
				Message: text,
				Metadata: map[string]interface{}{
					"description": text,
				},
			}
			ruleID := getRuleID(result.Namespace, ruleType)
			var idx int
			if existingIdx, ok := getRuleIndex(ruleID, indices); ok {
				idx = existingIdx
			} else {
				idx = addRuleIndex(run, ruleID, emptyResult, indices)
			}

			run.CreateResultForRule(ruleID).
				WithRuleIndex(idx).
				WithLevel("none").
				WithMessage(sarif.NewTextMessage(text)).
				AddLocation(
					sarif.NewLocationWithPhysicalLocation(
						sarif.NewPhysicalLocation().
							WithArtifactLocation(
								sarif.NewSimpleArtifactLocation(filepath.ToSlash(result.FileName)),
							),
					),
				)
		}
	}

	// Add run metadata
	exitCode := 0
	exitDesc := exitNoViolations
	if hasFailures(results) {
		exitCode = 1
		exitDesc = exitViolations
	} else if hasWarnings(results) {
		exitDesc = exitWarnings
	} else if hasExceptions(results) {
		exitDesc = exitExceptions
	}

	successful := true
	invocation := sarif.NewInvocation()
	invocation.ExecutionSuccessful = &successful
	invocation.ExitCode = &exitCode
	invocation.ExitCodeDescription = &exitDesc

	run.Invocations = []*sarif.Invocation{invocation}

	// Add the run to the report
	report.AddRun(run)

	// Write the report
	return report.Write(s.writer)
}

// Report is not supported in SARIF output
func (s *SARIF) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in SARIF output")
}

// hasResults returns true if any of the results contain items in the specified field
func hasResults(results []CheckResult, field string) bool {
	return slices.ContainsFunc(results, func(r CheckResult) bool {
		switch field {
		case "failures":
			return len(r.Failures) > 0
		case "warnings":
			return len(r.Warnings) > 0
		case "exceptions":
			return len(r.Exceptions) > 0
		default:
			return false
		}
	})
}

// hasFailures returns true if any of the results contain failures
func hasFailures(results []CheckResult) bool {
	return hasResults(results, "failures")
}

// hasWarnings returns true if any of the results contain warnings
func hasWarnings(results []CheckResult) bool {
	return hasResults(results, "warnings")
}

// hasExceptions returns true if any of the results contain exceptions
func hasExceptions(results []CheckResult) bool {
	return hasResults(results, "exceptions")
}
