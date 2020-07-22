package output

// Result describes the result of a single rule evaluation.
type Result struct {
	Message  string
	Metadata map[string]interface{}
	Traces   []error
}

func (r Result) Error() string {
	return r.Message
}

// CheckResult describes the result of a conftest evaluation.
// warning and failure "errors" produced by rego should be considered separate
// from other classes of exceptions.
type CheckResult struct {
	FileName   string
	Warnings   []Result
	Failures   []Result
	Exceptions []Result
	Successes  []Result
}

// NewResult creates a new result from the given message
func NewResult(message string, traces []error) Result {
	result := Result{
		Message:  message,
		Metadata: make(map[string]interface{}),
		Traces:   traces,
	}

	return result
}

func IsResultFailure(result CheckResult, failOnWarn bool) bool {
	return len(result.Failures) > 0 || (len(result.Warnings) > 0 && failOnWarn)
}
