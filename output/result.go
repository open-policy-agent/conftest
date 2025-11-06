package output

import (
	"encoding/json"
	"fmt"
	"slices"
)

// Result describes the result of a single rule evaluation.
type Result struct {
	Message  string         `json:"msg"`
	Location *Location      `json:"loc,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Outputs  []string       `json:"outputs,omitempty"`
}

// Location describes the origin location in the configuration file that
// caused the result to be produced.
type Location struct {
	File string      `json:"file,omitempty"`
	Line json.Number `json:"line,omitempty"`
}

func (l Location) String() string {
	return fmt.Sprintf("%s L%s", l.File, l.Line.String())
}

const (
	msgField = "msg"
	locField = "_loc"
)

var reservedFields = []string{
	msgField,
	locField,
}

// NewResult creates a new result. An error is returned if the
// metadata could not be successfully parsed.
func NewResult(metadata map[string]any) (Result, error) {
	if metadata == nil {
		return Result{}, fmt.Errorf("metadata must be supplied")
	}
	msg, ok := lookup[string](metadata, msgField)
	if !ok {
		return Result{}, fmt.Errorf("%q field must be present and a string", msgField)
	}

	result := Result{
		Message:  msg,
		Metadata: make(map[string]any),
	}

	if loc, ok := metadata[locField]; ok {
		if l := parseLocation(loc); l != nil {
			result.Location = l
		}
	}

	for k, v := range metadata {
		if !slices.Contains(reservedFields, k) {
			result.Metadata[k] = v
		}
	}

	return result, nil
}

func parseLocation(location any) *Location {
	loc, ok := location.(map[string]any)
	if !ok {
		return nil
	}

	l := &Location{}
	if file, ok := lookup[string](loc, "file"); ok {
		l.File = file
	}
	if line, ok := lookup[json.Number](loc, "line"); ok {
		l.Line = line
	}

	if l.File == "" && l.Line.String() == "" {
		return nil
	}

	return l
}

func lookup[T any](m map[string]any, k string) (value T, ok bool) {
	x, ok := m[k]
	if !ok {
		return
	}
	value, ok = x.(T)
	return
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
