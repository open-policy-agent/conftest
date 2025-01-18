package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	// Test file and namespace
	testFileName  = "examples/kubernetes/service.yaml"
	testNamespace = "namespace"

	// Test messages
	testFailureMsg = "first failure"
	testWarningMsg = "first warning"
	testSecondWarn = "second warning"

	// Test metadata
	testFailureDesc = "A detailed description of the failure"
	testFailureURL  = "https://example.com/docs"
	testFailureHelp = "How to fix this failure"
	testWarnDesc    = "A detailed description of the warning"
	testWarnURL     = "https://example.com/warnings"
	testWarnHelp    = "How to fix this warning"

	// Test policy metadata
	testPackage  = "security.container"
	testRuleName = "no_root_user"
)

func TestSARIF(t *testing.T) {
	tests := []struct {
		name     string
		input    []CheckResult
		validate func(t *testing.T, output string)
	}{
		{
			name: "success path - no violations",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result (skipped), got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelNone,
					kind:      kindSkipped,
					message:   skippedDesc,
					ruleID:    ruleSkippedBase,
					namespace: testNamespace,
				})

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   0,
					exitDesc:   exitNoViolations,
				})

				validateTimestamps(t, run.Invocations[0])
			},
		},
		{
			name: "single failure with basic result structure",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures:  []Result{{Message: testFailureMsg}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    ruleFailureBase,
					namespace: testNamespace,
				})

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "multiple warnings and failures",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Warnings:  []Result{{Message: testWarningMsg}, {Message: testSecondWarn}},
					Failures:  []Result{{Message: testFailureMsg}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 3 {
					t.Errorf("expected 3 results, got %d", len(run.Results))
				}

				// Count warnings and failures while validating each result
				warningCount := 0
				failureCount := 0
				for _, result := range run.Results {
					switch result.Level {
					case levelWarning:
						warningCount++
						validateResult(t, result, struct {
							level     string
							kind      string
							message   string
							ruleID    string
							namespace string
						}{
							level:     levelWarning,
							kind:      kindReview,
							message:   result.Message.Text,
							ruleID:    ruleWarningBase,
							namespace: testNamespace,
						})
					case levelError:
						failureCount++
						validateResult(t, result, struct {
							level     string
							kind      string
							message   string
							ruleID    string
							namespace string
						}{
							level:     levelError,
							kind:      kindFail,
							message:   testFailureMsg,
							ruleID:    ruleFailureBase,
							namespace: testNamespace,
						})
					}
				}

				if warningCount != 2 {
					t.Errorf("expected 2 warnings, got %d", warningCount)
				}
				if failureCount != 1 {
					t.Errorf("expected 1 failure, got %d", failureCount)
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "handles stdin input",
			input: []CheckResult{
				{
					FileName:  "-",
					Namespace: testNamespace,
					Failures:  []Result{{Message: testFailureMsg}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				result := run.Results[0]
				validateResult(t, result, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    ruleFailureBase,
					namespace: testNamespace,
				})

				// Verify location URI for stdin
				if result.Locations[0].PhysicalLocation.ArtifactLocation.URI != "-" {
					t.Errorf("expected URI '-' for stdin, got '%s'", result.Locations[0].PhysicalLocation.ArtifactLocation.URI)
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "includes metadata in rules and results",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{{
						Message: testFailureMsg,
						Metadata: map[string]interface{}{
							"description": testFailureDesc,
							"url":         testFailureURL,
							"help":        testFailureHelp,
						},
					}},
					Warnings: []Result{{
						Message: testWarningMsg,
						Metadata: map[string]interface{}{
							"description": testWarnDesc,
							"url":         testWarnURL,
							"help":        testWarnHelp,
						},
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 2 {
					t.Errorf("expected 2 results, got %d", len(run.Results))
				}

				// Find and validate failure rule/result
				expectedFailureRuleID := fmt.Sprintf("%s/%s", ruleFailureBase, strings.ToLower(strings.ReplaceAll(testFailureDesc, " ", "-")))
				failureRule, failureResult, err := findRule(run, expectedFailureRuleID)
				if err != nil {
					t.Fatal(err)
				}

				validateRuleMetadata(t, failureRule, struct {
					description string
					helpURI     string
					helpText    string
					namespace   string
				}{
					description: testFailureDesc,
					helpURI:     testFailureURL,
					helpText:    testFailureHelp,
					namespace:   testNamespace,
				})

				validateResult(t, *failureResult, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    expectedFailureRuleID,
					namespace: testNamespace,
				})

				// Find and validate warning rule/result
				expectedWarningRuleID := fmt.Sprintf("%s/%s", ruleWarningBase, strings.ToLower(strings.ReplaceAll(testWarnDesc, " ", "-")))
				warningRule, warningResult, err := findRule(run, expectedWarningRuleID)
				if err != nil {
					t.Fatal(err)
				}

				validateRuleMetadata(t, warningRule, struct {
					description string
					helpURI     string
					helpText    string
					namespace   string
				}{
					description: testWarnDesc,
					helpURI:     testWarnURL,
					helpText:    testWarnHelp,
					namespace:   testNamespace,
				})

				validateResult(t, *warningResult, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelWarning,
					kind:      kindReview,
					message:   testWarningMsg,
					ruleID:    expectedWarningRuleID,
					namespace: testNamespace,
				})
			},
		},
		{
			name: "handles exceptions",
			input: []CheckResult{
				{
					FileName:  "examples/exceptions/deployments.yaml",
					Namespace: "main",
					Failures: []Result{{
						Message: "Containers must not run as root",
					}},
					Exceptions: []Result{{
						Message: "data.main.exception[_][_] == \"run_as_root\"",
						Metadata: map[string]interface{}{
							"description": "Exception for running as root",
							"error":       "run_as_root",
							"code":        "policy_exception",
						},
					}},
				},
				{
					FileName:  "examples/exceptions/other.yaml",
					Namespace: "main",
					Exceptions: []Result{{
						Message: "data.main.exception[_][_] == \"other\"",
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 3 {
					t.Errorf("expected 3 results (1 failure, 2 exceptions), got %d", len(run.Results))
				}

				exceptionCount := 0
				failureCount := 0
				for _, result := range run.Results {
					switch result.Kind {
					case kindInfo:
						exceptionCount++
						validateResult(t, result, struct {
							level     string
							kind      string
							message   string
							ruleID    string
							namespace string
						}{
							level:     levelNote,
							kind:      kindInfo,
							message:   result.Message.Text,
							ruleID:    ruleExceptionBase,
							namespace: "main",
						})

						// Verify error details if present
						if result.Message.Text == "data.main.exception[_][_] == \"run_as_root\"" {
							if errType, ok := result.Properties["error"].(string); !ok || errType != "run_as_root" {
								t.Error("expected error type in properties")
							}
							if code, ok := result.Properties["code"].(string); !ok || code != "policy_exception" {
								t.Error("expected error code in properties")
							}
						}

					case kindFail:
						failureCount++
						validateResult(t, result, struct {
							level     string
							kind      string
							message   string
							ruleID    string
							namespace string
						}{
							level:     levelError,
							kind:      kindFail,
							message:   "Containers must not run as root",
							ruleID:    ruleFailureBase,
							namespace: "main",
						})
					}
				}

				if exceptionCount != 2 {
					t.Errorf("expected 2 exceptions, got %d", exceptionCount)
				}
				if failureCount != 1 {
					t.Errorf("expected 1 failure, got %d", failureCount)
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "success result when checks pass",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Successes: 1,
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelNone,
					kind:      kindPass,
					message:   successDesc,
					ruleID:    rulePassBase,
					namespace: testNamespace,
				})

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   0,
					exitDesc:   exitNoViolations,
				})
			},
		},
		{
			name: "skipped result when no checks run",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Successes: 0,
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelNone,
					kind:      kindSkipped,
					message:   skippedDesc,
					ruleID:    ruleSkippedBase,
					namespace: testNamespace,
				})

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   0,
					exitDesc:   exitNoViolations,
				})
			},
		},
		{
			name: "no success/skipped result when failures exist",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Successes: 5,
					Failures:  []Result{{Message: testFailureMsg}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result (failure only), got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    ruleFailureBase,
					namespace: testNamespace,
				})
			},
		},
		{
			name: "rule ID generation with policy metadata",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{{
						Message: testFailureMsg,
						Metadata: map[string]interface{}{
							"package":     testPackage,
							"rule":        testRuleName,
							"description": testFailureDesc,
						},
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				result := run.Results[0]
				expectedRuleID := fmt.Sprintf("%s/%s/%s", testNamespace, testPackage, testRuleName)
				if result.RuleID != expectedRuleID {
					t.Errorf("expected rule ID '%s', got '%s'", expectedRuleID, result.RuleID)
				}

				// Find and validate the rule
				rule, foundResult, err := findRule(run, expectedRuleID)
				if err != nil {
					t.Fatal(err)
				}
				if foundResult == nil {
					t.Fatal("no result found for rule")
				}
				result = *foundResult

				// Verify package and rule in rule properties
				if pkg, ok := rule.Properties["package"].(string); !ok || pkg != testPackage {
					t.Errorf("expected package '%s' in rule properties, got '%v'", testPackage, rule.Properties["package"])
				}
				if ruleName, ok := rule.Properties["rule"].(string); !ok || ruleName != testRuleName {
					t.Errorf("expected rule name '%s' in rule properties, got '%v'", testRuleName, rule.Properties["rule"])
				}

				validateResult(t, result, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    expectedRuleID,
					namespace: testNamespace,
				})
			},
		},
		{
			name: "rule ID generation with description fallback",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{{
						Message: testFailureMsg,
						Metadata: map[string]interface{}{
							"description": testFailureDesc,
						},
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				result := run.Results[0]
				expectedRuleID := fmt.Sprintf("%s/%s", ruleFailureBase, strings.ToLower(strings.ReplaceAll(testFailureDesc, " ", "-")))
				if result.RuleID != expectedRuleID {
					t.Errorf("expected rule ID '%s', got '%s'", expectedRuleID, result.RuleID)
				}

				validateResult(t, result, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    expectedRuleID,
					namespace: testNamespace,
				})
			},
		},
		{
			name: "rule ID generation with no metadata",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{{
						Message:  testFailureMsg,
						Metadata: map[string]interface{}{},
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   testFailureMsg,
					ruleID:    ruleFailureBase,
					namespace: testNamespace,
				})
			},
		},
		{
			name: "rule ID generation with message hash",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{{
						Message: testFailureMsg,
						Metadata: map[string]interface{}{
							"severity": "HIGH",
						},
					}},
				},
			},
			validate: func(t *testing.T, output string) {
				var report sarifReport
				if err := json.NewDecoder(strings.NewReader(output)).Decode(&report); err != nil {
					t.Fatalf("failed to decode SARIF output: %v", err)
				}

				run := report.Runs[0]
				result := run.Results[0]
				expectedRuleID := ruleFailureBase
				if result.RuleID != expectedRuleID {
					t.Errorf("expected rule ID '%s', got '%s'", expectedRuleID, result.RuleID)
				}
			},
		},
		{
			name: "multiple violations from same rule",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Failures: []Result{
						{
							Message: "Container 'app1' runs as privileged",
							Metadata: map[string]interface{}{
								"rule":      "no_privileged",
								"container": "app1",
								"query":     "data.kubernetes.deny_privileged_container",
							},
						},
						{
							Message: "Container 'app2' runs as privileged",
							Metadata: map[string]interface{}{
								"rule":      "no_privileged",
								"container": "app2",
								"query":     "data.kubernetes.deny_privileged_container",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 2 {
					t.Errorf("expected 2 results, got %d", len(run.Results))
				}

				// Both results should have the same ruleId and ruleIndex
				firstResult := run.Results[0]
				secondResult := run.Results[1]
				if firstResult.RuleID != secondResult.RuleID {
					t.Errorf("expected same rule ID, got '%s' and '%s'", firstResult.RuleID, secondResult.RuleID)
				}
				if firstResult.RuleIndex != secondResult.RuleIndex {
					t.Errorf("expected same rule index, got %d and %d", firstResult.RuleIndex, secondResult.RuleIndex)
				}

				// Validate each result
				for _, result := range run.Results {
					validateResult(t, result, struct {
						level     string
						kind      string
						message   string
						ruleID    string
						namespace string
					}{
						level:     levelError,
						kind:      kindFail,
						message:   result.Message.Text,
						ruleID:    fmt.Sprintf("%s-kubernetes-deny_privileged_container", ruleFailureBase),
						namespace: testNamespace,
					})

					// Verify query path in properties
					if query, ok := result.Properties["query"].(string); !ok || query != "data.kubernetes.deny_privileged_container" {
						t.Errorf("unexpected query: %v", result.Properties["query"])
					}

					// Verify container in properties
					if container, ok := result.Properties["container"].(string); !ok || (container != "app1" && container != "app2") {
						t.Errorf("unexpected container: %v", result.Properties["container"])
					}
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "cross package rules",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: "kubernetes",
					Failures: []Result{
						{
							Message: "Container runs as privileged",
							Metadata: map[string]interface{}{
								"query": "data.kubernetes.deny_privileged_container",
							},
						},
					},
				},
				{
					FileName:  testFileName,
					Namespace: "custom",
					Failures: []Result{
						{
							Message: "Custom rule violation",
							Metadata: map[string]interface{}{
								"query": "data.custom.deny_custom",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 2 {
					t.Errorf("expected 2 results, got %d", len(run.Results))
				}

				// Validate kubernetes rule result
				validateResult(t, run.Results[0], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   "Container runs as privileged",
					ruleID:    fmt.Sprintf("%s-kubernetes-deny_privileged_container", ruleFailureBase),
					namespace: "kubernetes",
				})

				// Validate custom rule result
				validateResult(t, run.Results[1], struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   "Custom rule violation",
					ruleID:    fmt.Sprintf("%s-custom-deny_custom", ruleFailureBase),
					namespace: "custom",
				})

				// Verify query paths in properties
				if query, ok := run.Results[0].Properties["query"].(string); !ok || query != "data.kubernetes.deny_privileged_container" {
					t.Errorf("unexpected query in first result: %v", run.Results[0].Properties["query"])
				}
				if query, ok := run.Results[1].Properties["query"].(string); !ok || query != "data.custom.deny_custom" {
					t.Errorf("unexpected query in second result: %v", run.Results[1].Properties["query"])
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   1,
					exitDesc:   exitViolations,
				})
			},
		},
		{
			name: "exception handling",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: testNamespace,
					Exceptions: []Result{
						{
							Message: "Division by zero in policy evaluation",
							Metadata: map[string]interface{}{
								"error": "divide by zero",
								"code":  "rego_type_error",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				result := run.Results[0]
				validateResult(t, result, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelNote,
					kind:      kindInfo,
					message:   "Division by zero in policy evaluation",
					ruleID:    ruleExceptionBase,
					namespace: testNamespace,
				})

				// Verify error details in properties
				if errMsg, ok := result.Properties["error"].(string); !ok || errMsg != "divide by zero" {
					t.Errorf("expected error 'divide by zero', got '%v'", result.Properties["error"])
				}
				if code, ok := result.Properties["code"].(string); !ok || code != "rego_type_error" {
					t.Errorf("expected code 'rego_type_error', got '%v'", result.Properties["code"])
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   0,
					exitDesc:   exitExceptions,
				})
			},
		},
		{
			name: "query path inclusion",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: "kubernetes.security",
					Failures: []Result{
						{
							Message: "Security violation",
							Metadata: map[string]interface{}{
								"query":   "data.kubernetes.security.deny_privileged",
								"package": "kubernetes.security",
								"rule":    "deny_privileged",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 1 {
					t.Errorf("expected 1 result, got %d", len(run.Results))
				}

				result := run.Results[0]
				expectedRuleID := fmt.Sprintf("%s/%s/%s", "kubernetes.security", "kubernetes.security", "deny_privileged")

				validateResult(t, result, struct {
					level     string
					kind      string
					message   string
					ruleID    string
					namespace string
				}{
					level:     levelError,
					kind:      kindFail,
					message:   "Security violation",
					ruleID:    expectedRuleID,
					namespace: "kubernetes.security",
				})

				// Verify query path in properties
				if query, ok := result.Properties["query"].(string); !ok || query != "data.kubernetes.security.deny_privileged" {
					t.Errorf("unexpected query path: %v", result.Properties["query"])
				}

				// Find and validate the rule
				rule, _, err := findRule(run, expectedRuleID)
				if err != nil {
					t.Fatal(err)
				}

				// Verify package and rule in rule properties
				if pkg, ok := rule.Properties["package"].(string); !ok || pkg != "kubernetes.security" {
					t.Errorf("unexpected package: %v", rule.Properties["package"])
				}
				if ruleName, ok := rule.Properties["rule"].(string); !ok || ruleName != "deny_privileged" {
					t.Errorf("unexpected rule: %v", rule.Properties["rule"])
				}
			},
		},
		{
			name: "policy warnings",
			input: []CheckResult{
				{
					FileName:  testFileName,
					Namespace: "kubernetes",
					Warnings: []Result{
						{
							Message: "Memory limits not set",
							Metadata: map[string]interface{}{
								"query":     "data.kubernetes.warn_no_memory_limits",
								"package":   "kubernetes",
								"rule":      "warn_memory_limits",
								"container": "app",
							},
						},
						{
							Message: "CPU limits not set",
							Metadata: map[string]interface{}{
								"query":     "data.kubernetes.warn_no_cpu_limits",
								"package":   "kubernetes",
								"rule":      "warn_cpu_limits",
								"container": "app",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, output string) {
				report := decodeSARIFReport(t, output)
				run := report.Runs[0]

				if len(run.Results) != 2 {
					t.Errorf("expected 2 warning results, got %d", len(run.Results))
				}

				// Expected rule IDs based on package and rule names
				expectedRuleIDs := map[string]bool{
					"kubernetes/kubernetes/warn_memory_limits": false,
					"kubernetes/kubernetes/warn_cpu_limits":    false,
				}

				for _, result := range run.Results {
					// Mark rule ID as found
					expectedRuleIDs[result.RuleID] = true

					validateResult(t, result, struct {
						level     string
						kind      string
						message   string
						ruleID    string
						namespace string
					}{
						level:     levelWarning,
						kind:      kindReview,
						message:   result.Message.Text,
						ruleID:    result.RuleID,
						namespace: "kubernetes",
					})

					// Find and validate the rule
					rule, _, err := findRule(run, result.RuleID)
					if err != nil {
						t.Fatalf("rule not found: %s", result.RuleID)
					}

					// Verify rule properties
					if pkg, ok := rule.Properties["package"].(string); !ok || pkg != "kubernetes" {
						t.Errorf("unexpected package in rule: %v", rule.Properties["package"])
					}
					if ruleName, ok := rule.Properties["rule"].(string); !ok || !strings.HasPrefix(ruleName, "warn_") {
						t.Errorf("unexpected rule name: %v", rule.Properties["rule"])
					}

					// Verify query path and container in properties
					if query, ok := result.Properties["query"].(string); !ok || !strings.HasPrefix(query, "data.kubernetes.warn_") {
						t.Errorf("unexpected query: %v", result.Properties["query"])
					}
					if container, ok := result.Properties["container"].(string); !ok || container != "app" {
						t.Errorf("unexpected container: %v", result.Properties["container"])
					}
				}

				// Verify all expected rule IDs were found
				for ruleID, found := range expectedRuleIDs {
					if !found {
						t.Errorf("expected rule ID not found: %s", ruleID)
					}
				}

				validateInvocation(t, run.Invocations[0], struct {
					successful bool
					exitCode   int
					exitDesc   string
				}{
					successful: true,
					exitCode:   0,
					exitDesc:   exitWarnings,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSARIF(&buf)

			if err := s.Output(tt.input); err != nil {
				t.Fatalf("failed to output SARIF: %v", err)
			}

			tt.validate(t, buf.String())
		})
	}
}

func TestSARIFReport(t *testing.T) {
	var buf bytes.Buffer
	sarif := NewSARIF(&buf)

	err := sarif.Report(nil, "")
	if err == nil {
		t.Error("expected error for Report call")
	}

	wantErrMsg := "report is not supported in SARIF output"
	if !strings.Contains(err.Error(), wantErrMsg) {
		t.Errorf("expected error containing '%s', got: %v", wantErrMsg, err)
	}
}

func TestGetRuleID(t *testing.T) {
	tests := []struct {
		name      string
		result    Result
		rType     resultType
		namespace string
		want      string
	}{
		{
			name: "success result uses base ID",
			result: Result{
				Message: "success",
				Metadata: map[string]interface{}{
					"query": "data.main.deny",
				},
			},
			rType:     successResultType,
			namespace: "main",
			want:      "conftest-pass",
		},
		{
			name: "skipped result uses base ID",
			result: Result{
				Message: "skipped",
				Metadata: map[string]interface{}{
					"query": "data.main.deny",
				},
			},
			rType:     skippedResultType,
			namespace: "main",
			want:      "conftest-skipped",
		},
		{
			name: "uses package and rule from metadata when available",
			result: Result{
				Message: "violation",
				Metadata: map[string]interface{}{
					"package": "kubernetes",
					"rule":    "deny_privileged",
					"query":   "data.kubernetes.deny",
				},
			},
			rType:     failureResultType,
			namespace: "main",
			want:      "main/kubernetes/deny_privileged",
		},
		{
			name: "uses query path when package/rule not available",
			result: Result{
				Message: "violation",
				Metadata: map[string]interface{}{
					"query": "data.kubernetes.deny",
				},
			},
			rType:     failureResultType,
			namespace: "main",
			want:      "conftest-failure-kubernetes-deny",
		},
		{
			name: "uses description when no query or package/rule",
			result: Result{
				Message: "violation",
				Metadata: map[string]interface{}{
					"description": "No privileged containers allowed",
				},
			},
			rType:     failureResultType,
			namespace: "main",
			want:      "conftest-failure/no-privileged-containers-allowed",
		},
		{
			name: "falls back to base ID when no identifying information",
			result: Result{
				Message:  "violation",
				Metadata: map[string]interface{}{},
			},
			rType:     failureResultType,
			namespace: "main",
			want:      "conftest-failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRuleID(tt.result, tt.rType, tt.namespace)
			if got != tt.want {
				t.Errorf("getRuleID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryInformation(t *testing.T) {
	var buf bytes.Buffer
	s := NewSARIF(&buf)

	results := []CheckResult{
		{
			FileName:  "test.yaml",
			Namespace: "main",
			Failures: []Result{
				{
					Message: "violation found",
					Metadata: map[string]interface{}{
						"query":   "data.main.deny",
						"traces":  []string{"trace1", "trace2"},
						"outputs": []string{"output1", "output2"},
						"package": "main",
						"rule":    "deny",
					},
				},
			},
		},
	}

	if err := s.Output(results); err != nil {
		t.Fatal(err)
	}

	report := decodeSARIFReport(t, buf.String())
	run := report.Runs[0]

	if len(run.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(run.Results))
	}

	expectedRuleID := "main/main/deny"
	validateResult(t, run.Results[0], struct {
		level     string
		kind      string
		message   string
		ruleID    string
		namespace string
	}{
		level:     levelError,
		kind:      kindFail,
		message:   "violation found",
		ruleID:    expectedRuleID,
		namespace: "main",
	})

	// Find and validate the rule
	_, foundResult, err := findRule(run, expectedRuleID)
	if err != nil {
		t.Fatal(err)
	}
	if foundResult == nil {
		t.Fatal("no result found for rule")
	}
	result := *foundResult

	// Verify query information in properties
	if query, ok := result.Properties["query"].(string); !ok || query != "data.main.deny" {
		t.Errorf("unexpected query: %v", result.Properties["query"])
	}

	// Verify traces and outputs are preserved
	if traces, ok := result.Properties["traces"].([]interface{}); !ok || len(traces) != 2 {
		t.Errorf("unexpected traces: %v", result.Properties["traces"])
	}
	if outputs, ok := result.Properties["outputs"].([]interface{}); !ok || len(outputs) != 2 {
		t.Errorf("unexpected outputs: %v", result.Properties["outputs"])
	}

	validateInvocation(t, run.Invocations[0], struct {
		successful bool
		exitCode   int
		exitDesc   string
	}{
		successful: true,
		exitCode:   1,
		exitDesc:   exitViolations,
	})
}

// Helper functions for testing
func decodeSARIFReport(t *testing.T, output string) sarifReport {
	t.Helper()
	var report sarifReport
	if err := json.NewDecoder(strings.NewReader(output)).Decode(&report); err != nil {
		t.Fatalf("failed to decode SARIF output: %v", err)
	}
	return report
}

// validateResult validates that a SARIF result matches the expected values for level, kind,
// message, rule ID and namespace
func validateResult(t *testing.T, result sarifResult, expected struct {
	level     string
	kind      string
	message   string
	ruleID    string
	namespace string
}) {
	t.Helper()
	if result.Level != expected.level {
		t.Errorf("expected level '%s', got '%s'", expected.level, result.Level)
	}
	if result.Kind != expected.kind {
		t.Errorf("expected kind '%s', got '%s'", expected.kind, result.Kind)
	}
	if result.Message.Text != expected.message {
		t.Errorf("expected message '%s', got '%s'", expected.message, result.Message.Text)
	}
	if result.RuleID != expected.ruleID {
		t.Errorf("expected rule ID '%s', got '%s'", expected.ruleID, result.RuleID)
	}
	if ns, ok := result.Properties["namespace"].(string); !ok || ns != expected.namespace {
		t.Errorf("expected namespace '%s' in properties, got '%v'", expected.namespace, result.Properties["namespace"])
	}
}

// validateInvocation validates that a SARIF invocation matches the expected values for
// execution success, exit code and exit code description.
func validateInvocation(t *testing.T, invocation sarifInvocation, expected struct {
	successful bool
	exitCode   int
	exitDesc   string
}) {
	t.Helper()
	if invocation.ExecutionSuccessful != expected.successful {
		t.Error("unexpected execution success state")
	}
	if invocation.ExitCode != expected.exitCode {
		t.Errorf("expected exit code %d, got %d", expected.exitCode, invocation.ExitCode)
	}
	if invocation.ExitCodeDescription != expected.exitDesc {
		t.Errorf("unexpected exit code description: %s", invocation.ExitCodeDescription)
	}
}

// validateTimestamps validates that the start and end times in a SARIF invocation
// are valid RFC3339 timestamps and chronologically ordered.
func validateTimestamps(t *testing.T, invocation sarifInvocation) {
	t.Helper()
	startTime, err := time.Parse(time.RFC3339, invocation.StartTimeUtc)
	if err != nil {
		t.Errorf("invalid start time format: %v", err)
	}
	endTime, err := time.Parse(time.RFC3339, invocation.EndTimeUtc)
	if err != nil {
		t.Errorf("invalid end time format: %v", err)
	}
	if endTime.Before(startTime) {
		t.Error("end time should not be before start time")
	}
}

// findRule returns the rule and its associated result for a given rule ID.
// If the rule is not found, it returns an error.
func findRule(run sarifRun, ruleID string) (rule *sarifRule, result *sarifResult, err error) {
	for i := range run.Tool.Driver.Rules {
		r := run.Tool.Driver.Rules[i] // Take reference to array element, not loop variable
		if r.ID == ruleID {
			rule = &r
			// Find the first result that references this rule
			for j := range run.Results {
				res := run.Results[j] // Take reference to array element, not loop variable
				if res.RuleIndex == i {
					result = &res
					return rule, result, nil
				}
			}
			// Rule found but no results reference it
			return rule, nil, nil
		}
	}
	return nil, nil, fmt.Errorf("rule not found: %s", ruleID)
}

// validateRuleMetadata validates that a SARIF rule matches the expected values for description, help URI,
// help text and namespace.
func validateRuleMetadata(t *testing.T, rule *sarifRule, expected struct {
	description string
	helpURI     string
	helpText    string
	namespace   string
}) {
	t.Helper()
	if rule == nil {
		t.Fatal("rule is nil")
	}
	if rule.FullDescription == nil || rule.FullDescription.Text != expected.description {
		t.Error("invalid rule description")
	}
	if rule.HelpURI != expected.helpURI {
		t.Error("invalid help URI")
	}
	if rule.Help == nil || rule.Help.Text != expected.helpText {
		t.Error("invalid help text")
	}
	if ns, ok := rule.Properties["namespace"].(string); !ok || ns != expected.namespace {
		t.Errorf("expected namespace '%s' in rule properties, got '%v'", expected.namespace, ns)
	}
}
