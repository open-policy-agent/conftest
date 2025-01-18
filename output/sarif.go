package output

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/open-policy-agent/opa/tester"
)

const (
	// SARIF schema and version
	sarifSchemaURI = "https://docs.oasis-open.org/sarif/sarif/v2.1.0/errata01/os/schemas/sarif-schema-2.1.0.json"
	sarifVersion   = "2.1.0"

	// Tool information
	toolName = "conftest"
	toolURI  = "https://github.com/open-policy-agent/conftest"

	// SARIF levels
	levelError   = "error"
	levelWarning = "warning"
	levelNote    = "note"
	levelNone    = "none"

	// SARIF kinds
	kindFail    = "fail"
	kindReview  = "review"
	kindInfo    = "informational"
	kindPass    = "pass"
	kindSkipped = "notApplicable"

	// Rule ID prefixes
	ruleFailureBase   = "conftest-failure"
	ruleWarningBase   = "conftest-warning"
	ruleExceptionBase = "conftest-exception"
	rulePassBase      = "conftest-pass"
	ruleSkippedBase   = "conftest-skipped"

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
// https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html
type SARIF struct {
	writer io.Writer
}

// NewSARIF creates a new SARIF with the given writer.
func NewSARIF(w io.Writer) *SARIF {
	return &SARIF{
		writer: w,
	}
}

// sarifReport represents the root object of a SARIF log file
type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun represents a single run of a tool
type sarifRun struct {
	Tool        sarifTool         `json:"tool"`
	Results     []sarifResult     `json:"results"`
	Invocations []sarifInvocation `json:"invocations"`
}

// sarifTool represents the analysis tool that was run
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver represents the analysis tool component that contains rule metadata
type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

// sarifRule represents a rule that was evaluated during the scan
type sarifRule struct {
	ID               string                 `json:"id"`
	ShortDescription sarifMessage           `json:"shortDescription"`
	FullDescription  *sarifMessage          `json:"fullDescription,omitempty"`
	Help             *sarifMessage          `json:"help,omitempty"`
	HelpURI          string                 `json:"helpUri,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// sarifResult represents a single analysis result
type sarifResult struct {
	RuleID     string                 `json:"ruleId"`
	RuleIndex  int                    `json:"ruleIndex"`
	Kind       string                 `json:"kind"`
	Level      string                 `json:"level"`
	Message    sarifMessage           `json:"message"`
	Locations  []sarifLocation        `json:"locations"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// sarifLocation represents a location within a programming artifact
type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

// sarifPhysicalLocation represents the physical location where the result was detected
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

// sarifArtifactLocation represents the location of an artifact
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifMessage represents a message string or message with arguments
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifInvocation represents the runtime environment of the analysis tool run
type sarifInvocation struct {
	ExecutionSuccessful bool   `json:"executionSuccessful"`
	ExitCode            int    `json:"exitCode"`
	ExitCodeDescription string `json:"exitCodeDescription"`
	StartTimeUtc        string `json:"startTimeUtc"`
	EndTimeUtc          string `json:"endTimeUtc"`
}

// resultKind represents the type of result being processed
type resultKind int

const (
	resultKindSuccess resultKind = iota
	resultKindSkipped
	resultKindException
	resultKindFailure
	resultKindWarning
)

// resultType represents the type of result being processed
type resultType struct {
	kind         resultKind
	ruleIDPrefix string
	kindStr      string
	level        string
	description  string
}

var (
	failureResultType = resultType{
		kind:         resultKindFailure,
		ruleIDPrefix: ruleFailureBase,
		kindStr:      kindFail,
		level:        levelError,
		description:  failureDesc,
	}
	warningResultType = resultType{
		kind:         resultKindWarning,
		ruleIDPrefix: ruleWarningBase,
		kindStr:      kindReview,
		level:        levelWarning,
		description:  warningDesc,
	}
	exceptionResultType = resultType{
		kind:         resultKindException,
		ruleIDPrefix: ruleExceptionBase,
		kindStr:      kindInfo,
		level:        levelNote,
		description:  exceptionDesc,
	}
	successResultType = resultType{
		kind:         resultKindSuccess,
		ruleIDPrefix: rulePassBase,
		kindStr:      kindPass,
		level:        levelNone,
		description:  successDesc,
	}
	skippedResultType = resultType{
		kind:         resultKindSkipped,
		ruleIDPrefix: ruleSkippedBase,
		kindStr:      kindSkipped,
		level:        levelNone,
		description:  skippedDesc,
	}
)

// getRuleID generates a unique rule ID based on metadata and result type
func getRuleID(result Result, rType resultType, namespace string) string {
	// Always use base ID for success, skipped, and exception results
	switch rType.kind {
	case resultKindSuccess, resultKindSkipped, resultKindException:
		return rType.ruleIDPrefix
	}

	// Use package and rule from metadata when available
	if pkg, ok := result.Metadata["package"].(string); ok {
		if rule, ok := result.Metadata["rule"].(string); ok {
			return fmt.Sprintf("%s/%s/%s", namespace, pkg, rule)
		}
	}

	// Use query path if available
	if query, ok := result.Metadata["query"].(string); ok {
		// Remove "data." prefix and convert dots to dashes
		query = strings.TrimPrefix(query, "data.")
		query = strings.ReplaceAll(query, ".", "-")
		return fmt.Sprintf("%s-%s", rType.ruleIDPrefix, query)
	}

	// Use description if available
	if desc, ok := result.Metadata["description"].(string); ok {
		return fmt.Sprintf("%s/%s", rType.ruleIDPrefix, strings.ToLower(strings.ReplaceAll(desc, " ", "-")))
	}

	// Fallback to base ID if no identifying information is available
	return rType.ruleIDPrefix
}

// createRule creates a new SARIF rule from a result
func createRule(result Result, rType resultType, namespace string) sarifRule {
	ruleID := getRuleID(result, rType, namespace)

	rule := sarifRule{
		ID: ruleID,
		ShortDescription: sarifMessage{
			Text: result.Message,
		},
		Properties: make(map[string]interface{}),
	}

	// Add policy metadata to rule properties
	if pkg, ok := result.Metadata["package"].(string); ok {
		rule.Properties["package"] = pkg
	}
	if ruleName, ok := result.Metadata["rule"].(string); ok {
		rule.Properties["rule"] = ruleName
	}
	if query, ok := result.Metadata["query"].(string); ok {
		rule.Properties["query"] = query
	}
	rule.Properties["namespace"] = namespace

	// Add additional rule metadata if available
	if desc, ok := result.Metadata["description"].(string); ok {
		rule.FullDescription = &sarifMessage{Text: desc}
	}
	if url, ok := result.Metadata["url"].(string); ok {
		rule.HelpURI = url
	}
	if help, ok := result.Metadata["help"].(string); ok {
		rule.Help = &sarifMessage{Text: help}
	}

	// Add any remaining metadata to properties
	for k, v := range result.Metadata {
		switch k {
		case "package", "rule", "description", "url", "help", "query":
			// Skip already processed fields
			continue
		default:
			rule.Properties[k] = v
		}
	}

	return rule
}

// createLocation creates a new SARIF location from a file path
func createLocation(filePath string) sarifLocation {
	return sarifLocation{
		PhysicalLocation: sarifPhysicalLocation{
			ArtifactLocation: sarifArtifactLocation{
				URI: filepath.ToSlash(filePath),
			},
		},
	}
}

// createProperties creates a new properties map with namespace information
func createProperties(metadata map[string]interface{}, namespace string) map[string]interface{} {
	properties := make(map[string]interface{})
	for k, v := range metadata {
		switch k {
		case "package", "rule", "description", "url", "help":
			// Skip rule-level metadata
			continue
		case "query", "traces", "outputs":
			// Include query-related information
			properties[k] = v
		default:
			properties[k] = v
		}
	}

	// Always include namespace
	properties["namespace"] = namespace

	return properties
}

// processResults processes a slice of results and adds them to the SARIF run
func processResults(run *sarifRun, results []Result, rType resultType, fileName, namespace string, ruleMap map[string]bool) error {
	for _, result := range results {
		// Create or get rule
		rule := createRule(result, rType, namespace)
		if !ruleMap[rule.ID] {
			run.Tool.Driver.Rules = append(run.Tool.Driver.Rules, rule)
			ruleMap[rule.ID] = true
		}

		// Find rule index
		ruleIndex := -1
		for i, r := range run.Tool.Driver.Rules {
			if r.ID == rule.ID {
				ruleIndex = i
				break
			}
		}

		if ruleIndex == -1 {
			return fmt.Errorf("rule %s not found in rules array after being added", rule.ID)
		}

		// Create result
		run.Results = append(run.Results, sarifResult{
			RuleID:     rule.ID,
			RuleIndex:  ruleIndex,
			Kind:       rType.kindStr,
			Level:      rType.level,
			Message:    sarifMessage{Text: result.Message},
			Locations:  []sarifLocation{createLocation(fileName)},
			Properties: createProperties(result.Metadata, namespace),
		})
	}
	return nil
}

// createSuccessResult creates a success result for a given file and namespace
func createSuccessResult(run *sarifRun, fileName, namespace string, ruleMap map[string]bool) error {
	result := Result{
		Message: successResultType.description,
	}
	return processResults(run, []Result{result}, successResultType, fileName, namespace, ruleMap)
}

// createSkippedResult creates a skipped result for a given file and namespace
func createSkippedResult(run *sarifRun, fileName, namespace string, ruleMap map[string]bool) error {
	result := Result{
		Message: skippedResultType.description,
	}
	return processResults(run, []Result{result}, skippedResultType, fileName, namespace, ruleMap)
}

// Output outputs the results in SARIF format.
func (s *SARIF) Output(results []CheckResult) error {
	startTime := time.Now().UTC()

	// Create SARIF report structure
	report := sarifReport{
		Schema:  sarifSchemaURI,
		Version: sarifVersion,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           toolName,
						InformationURI: toolURI,
						Rules:          []sarifRule{},
					},
				},
				Results:     []sarifResult{},
				Invocations: []sarifInvocation{},
			},
		},
	}

	run := &report.Runs[0]
	ruleMap := make(map[string]bool)

	// Process all results
	for _, result := range results {
		err := processResults(run, result.Failures, failureResultType, result.FileName, result.Namespace, ruleMap)
		if err != nil {
			return fmt.Errorf("process failures: %w", err)
		}
		err = processResults(run, result.Warnings, warningResultType, result.FileName, result.Namespace, ruleMap)
		if err != nil {
			return fmt.Errorf("process warnings: %w", err)
		}
		err = processResults(run, result.Exceptions, exceptionResultType, result.FileName, result.Namespace, ruleMap)
		if err != nil {
			return fmt.Errorf("process exceptions: %w", err)
		}

		// Add success result if no failures/warnings/exceptions
		if len(result.Failures) == 0 && len(result.Warnings) == 0 && len(result.Exceptions) == 0 {
			if result.Successes > 0 {
				err = createSuccessResult(run, result.FileName, result.Namespace, ruleMap)
				if err != nil {
					return fmt.Errorf("create success result: %w", err)
				}
			} else {
				err = createSkippedResult(run, result.FileName, result.Namespace, ruleMap)
				if err != nil {
					return fmt.Errorf("create skipped result: %w", err)
				}
			}
		}
	}

	// Add invocation information
	exitCode := 0
	exitDesc := exitNoViolations
	if hasFailures(results) {
		exitCode = 1
		exitDesc = exitViolations
	} else if hasWarnings(results) {
		exitCode = 0
		exitDesc = exitWarnings
	} else if hasExceptions(results) {
		exitCode = 0
		exitDesc = exitExceptions
	}

	run.Invocations = []sarifInvocation{
		{
			ExecutionSuccessful: true,
			ExitCode:            exitCode,
			ExitCodeDescription: exitDesc,
			StartTimeUtc:        startTime.Format(time.RFC3339),
			EndTimeUtc:          time.Now().UTC().Format(time.RFC3339),
		},
	}

	// Marshal to JSON
	encoder := json.NewEncoder(s.writer)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode sarif: %w", err)
	}

	return nil
}

// Report is not supported in SARIF output
func (s *SARIF) Report(_ []*tester.Result, _ string) error {
	return fmt.Errorf("report is not supported in SARIF output")
}

// hasFailures returns true if any of the results contain failures
func hasFailures(results []CheckResult) bool {
	for _, result := range results {
		if len(result.Failures) > 0 {
			return true
		}
	}
	return false
}

// hasWarnings returns true if any of the results contain warnings
func hasWarnings(results []CheckResult) bool {
	for _, result := range results {
		if len(result.Warnings) > 0 {
			return true
		}
	}
	return false
}

// hasExceptions returns true if any of the results contain exceptions
func hasExceptions(results []CheckResult) bool {
	for _, result := range results {
		if len(result.Exceptions) > 0 {
			return true
		}
	}
	return false
}
