package output

import (
	"fmt"
	"slices"
)

// Result describes the result of a single rule evaluation.
type Result struct {
	Message  string         `json:"msg"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Outputs  []string       `json:"outputs,omitempty"`
}

// NewResult creates a new result. An error is returned if the
// metadata could not be successfully parsed.
func NewResult(metadata map[string]any) (Result, error) {
	if _, ok := metadata["msg"]; !ok {
		return Result{}, fmt.Errorf("rule missing msg field: %v", metadata)
	}
	if _, ok := metadata["msg"].(string); !ok {
		return Result{}, fmt.Errorf("msg field must be string: %v", metadata)
	}

	result := Result{
		Message:  metadata["msg"].(string),
		Metadata: make(map[string]any),
	}

	for k, v := range metadata {
		if k != "msg" {
			result.Metadata[k] = v
		}
	}

	return result, nil
}

// Passed returns true if the result did not fail a policy.
func (r Result) Passed() bool {
	return r.Message == ""
}

// QueryResult describes the result of evaluting a query.
type QueryResult struct {

	// Query is the fully qualified query that was used
	// to determine the result. Ex: (data.main.deny)
	Query string `json:"query"`

	// Results are the individual results of the query.
	// When querying data.main.deny, multiple deny rules can
	// exist, producing multiple results.
	Results []Result `json:"results"`

	// Traces represents a single trace of how the query was
	// evaluated. Each trace value is a trace line.
	Traces []string `json:"traces"`

	// Output represents anything print()'ed during the query
	// evaluation. Each value is a print() call's result.
	Outputs []string `json:"outputs,omitempty"`
}

// Passed returns true if all of the results in the query
// passed and no failures were found.
func (q QueryResult) Passed() bool {
	for _, result := range q.Results {
		if !result.Passed() {
			return false
		}
	}

	return true
}

// CheckResult describes the result of a conftest policy evaluation.
// Errors produced by rego should be considered separate
// from other classes of exceptions.
type CheckResult struct {
	FileName   string        `json:"filename"`
	Namespace  string        `json:"namespace"`
	Successes  int           `json:"successes"`
	Skipped    []Result      `json:"skipped,omitempty"`
	Warnings   []Result      `json:"warnings,omitempty"`
	Failures   []Result      `json:"failures,omitempty"`
	Exceptions []Result      `json:"exceptions,omitempty"`
	Queries    []QueryResult `json:"queries,omitempty"`
}

// HasFailure returns true if any failures were encountered.
func (cr CheckResult) HasFailure() bool {
	return len(cr.Failures) > 0
}

// HasWarning returns true if any warnings were encountered.
func (cr CheckResult) HasWarning() bool {
	return len(cr.Warnings) > 0
}

// HasException returns true if any exceptions were encountered.
func (cr CheckResult) HasException() bool {
	return len(cr.Exceptions) > 0
}

// OnlySuccess returns true if there are no failures, warnings, or exceptions.
func (cr CheckResult) OnlySuccess() bool {
	return len(cr.Failures) == 0 && len(cr.Warnings) == 0 && len(cr.Exceptions) == 0
}

// CheckResults is a slice of CheckResult.
type CheckResults []CheckResult

// HasFailure returns true if any of the checks in the list has a failure.
func (cr CheckResults) HasFailure() bool {
	return slices.ContainsFunc(cr, func(x CheckResult) bool { return x.HasFailure() })
}

// HasWarning returns true if any of the checks in the list has a warning.
func (cr CheckResults) HasWarning() bool {
	return slices.ContainsFunc(cr, func(x CheckResult) bool { return x.HasWarning() })
}

// HasException returns true if any of the checks in the list has an exception.
func (cr CheckResults) HasException() bool {
	return slices.ContainsFunc(cr, func(x CheckResult) bool { return x.HasException() })
}

// OnlySuccess returns true if all of the checks have only success messages.
func (cr CheckResults) OnlySuccess() bool {
	return !slices.ContainsFunc(cr, func(x CheckResult) bool { return !x.OnlySuccess() })
}

// ExitCode returns the exit code that should be returned
// given all of the returned results.
func (cr CheckResults) ExitCode() int {
	if cr.HasFailure() {
		return 1
	}
	return 0
}

// ExitCodeFailOnWarn returns the exit code that should be returned
// given all of the returned results, and will consider warnings
// as failures.
func (cr CheckResults) ExitCodeFailOnWarn() int {
	if cr.HasFailure() {
		return 2
	}
	if cr.HasWarning() {
		return 1
	}
	return 0
}
