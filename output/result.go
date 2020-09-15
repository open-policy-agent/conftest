package output

import "fmt"

// Result represents the result of a single policy evaluation.
type Result struct {
	Message  string                 `json:"msg"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Traces   []string               `json:"traces,omitempty"`
}

// CheckResult describes the result of a conftest policy evaluation.
// Errors produced by rego should be considered separate
// from other classes of exceptions.
type CheckResult struct {
	Filename   string   `json:"filename"`
	Successes  int      `json:"successes"`
	Warnings   []Result `json:"warnings"`
	Failures   []Result `json:"failures"`
	Exceptions []Result `json:"exceptions"`
}

// NewResult creates a new result from the given message.
func NewResult(message string, traces []error) Result {
	result := Result{
		Message:  message,
		Metadata: make(map[string]interface{}),
		Traces:   errsToStrings(traces),
	}

	return result
}

// NewResultWithMetadata creates a new result from metadata. An error is returned if the
// metadata could not be successfully parsed.
func NewResultWithMetadata(metadata map[string]interface{}, traces []error) (Result, error) {
	if _, ok := metadata["msg"]; !ok {
		return Result{}, fmt.Errorf("rule missing msg field: %v", metadata)
	}
	if _, ok := metadata["msg"].(string); !ok {
		return Result{}, fmt.Errorf("msg field must be string: %v", metadata)
	}

	result := NewResult(metadata["msg"].(string), traces)
	for k, v := range metadata {
		if k != "msg" {
			result.Metadata[k] = v
		}
	}

	return result, nil
}

// ExitCode returns the exit code that should be returned
// given all of the returned results.
func ExitCode(results []CheckResult) int {
	var hasFailure bool
	for _, result := range results {
		if len(result.Failures) > 0 {
			hasFailure = true
		}
	}

	if hasFailure {
		return 1
	}

	return 0
}

// ExitCodeFailOnWarn returns the exit code that should be returned
// given all of the returned results, and will consider warnings
// as failures.
func ExitCodeFailOnWarn(results []CheckResult) int {
	var hasFailure bool
	var hasWarning bool
	for _, result := range results {
		if len(result.Failures) > 0 {
			hasFailure = true
		}

		if len(result.Warnings) > 0 {
			hasWarning = true
		}
	}

	if hasFailure {
		return 2
	}

	if hasWarning {
		return 1
	}

	return 0
}

func errsToStrings(errs []error) []string {
	res := []string{}
	for _, err := range errs {
		res = append(res, err.Error())
	}

	return res
}
