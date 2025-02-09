package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func TestSARIF_Output(t *testing.T) {
	tests := []struct {
		name     string
		results  []CheckResult
		wantErr  bool
		wantJSON string
	}{
		{
			name:    "empty results",
			results: []CheckResult{},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": []
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 0,
								"exitCodeDescription": "No policy violations found"
							}
						],
						"results": []
					}
				]
			}`,
		},
		{
			name: "single failure",
			results: []CheckResult{
				{
					FileName:  "test.yaml",
					Namespace: "main",
					Failures: []Result{
						{
							Message: "test failure",
							Metadata: map[string]interface{}{
								"package": "test",
								"rule":    "rule1",
							},
						},
					},
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/deny",
										"shortDescription": {
											"text": "Policy violation"
										},
										"properties": {
											"package": "test",
											"rule": "rule1"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 1,
								"exitCodeDescription": "Policy violations found"
							}
						],
						"results": [
							{
								"ruleId": "main/deny",
								"ruleIndex": 0,
								"level": "error",
								"message": {
									"text": "test failure"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
		{
			name: "single warning",
			results: []CheckResult{
				{
					FileName:  "test.yaml",
					Namespace: "main",
					Warnings: []Result{
						{
							Message: "test warning",
							Metadata: map[string]interface{}{
								"foo": "bar",
							},
						},
					},
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/warn",
										"shortDescription": {
											"text": "Policy warning"
										},
										"properties": {
											"foo": "bar"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 0,
								"exitCodeDescription": "Policy warnings found"
							}
						],
						"results": [
							{
								"ruleId": "main/warn",
								"ruleIndex": 0,
								"level": "warning",
								"message": {
									"text": "test warning"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
		{
			name: "single exception",
			results: []CheckResult{
				{
					FileName:  "test.yaml",
					Namespace: "main",
					Exceptions: []Result{
						{
							Message: "test exception",
							Metadata: map[string]interface{}{
								"description": "test exception description",
							},
						},
					},
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/allow",
										"shortDescription": {
											"text": "Policy exception"
										},
										"properties": {
											"description": "test exception description"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 0,
								"exitCodeDescription": "Policy exceptions found"
							}
						],
						"results": [
							{
								"ruleId": "main/allow",
								"ruleIndex": 0,
								"level": "note",
								"message": {
									"text": "test exception"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
		{
			name: "skipped result",
			results: []CheckResult{
				{
					FileName:  "test.yaml",
					Namespace: "main",
					Successes: 0,
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/skip",
										"shortDescription": {
											"text": "Policy check was skipped"
										},
										"properties": {
											"description": "Policy check was skipped"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 0,
								"exitCodeDescription": "No policy violations found"
							}
						],
						"results": [
							{
								"ruleId": "main/skip",
								"ruleIndex": 0,
								"level": "none",
								"message": {
									"text": "Policy check was skipped"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
		{
			name: "multiple results same rule",
			results: []CheckResult{
				{
					FileName:  "test1.yaml",
					Namespace: "main",
					Failures: []Result{
						{
							Message: "test failure 1",
							Metadata: map[string]interface{}{
								"package": "test",
								"rule":    "rule1",
							},
						},
						{
							Message: "test failure 2",
							Metadata: map[string]interface{}{
								"package": "test",
								"rule":    "rule1",
							},
						},
					},
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/deny",
										"shortDescription": {
											"text": "Policy violation"
										},
										"properties": {
											"package": "test",
											"rule": "rule1"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 1,
								"exitCodeDescription": "Policy violations found"
							}
						],
						"results": [
							{
								"ruleId": "main/deny",
								"ruleIndex": 0,
								"level": "error",
								"message": {
									"text": "test failure 1"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test1.yaml"
											}
										}
									}
								]
							},
							{
								"ruleId": "main/deny",
								"ruleIndex": 0,
								"level": "error",
								"message": {
									"text": "test failure 2"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test1.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
		{
			name: "successful policy check",
			results: []CheckResult{
				{
					FileName:  "test.yaml",
					Namespace: "main",
					Successes: 1,
				},
			},
			wantJSON: `{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": [
					{
						"tool": {
							"driver": {
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name": "conftest",
								"rules": [
									{
										"id": "main/success",
										"shortDescription": {
											"text": "Policy was satisfied successfully"
										},
										"properties": {
											"description": "Policy was satisfied successfully"
										}
									}
								]
							}
						},
						"invocations": [
							{
								"executionSuccessful": true,
								"exitCode": 0,
								"exitCodeDescription": "No policy violations found"
							}
						],
						"results": [
							{
								"ruleId": "main/success",
								"ruleIndex": 0,
								"level": "none",
								"message": {
									"text": "Policy was satisfied successfully"
								},
								"locations": [
									{
										"physicalLocation": {
											"artifactLocation": {
												"uri": "test.yaml"
											}
										}
									}
								]
							}
						]
					}
				]
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			s := NewSARIF(&buf)

			err := s.Output(tt.results)
			if (err != nil) != tt.wantErr {
				t.Errorf("SARIF.Output() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				compareJSON(t, buf.String(), tt.wantJSON)
			}
		})
	}
}

func TestGetRuleID(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		ruleType  string
		want      string
	}{
		{
			name:      "failure",
			namespace: "main",
			ruleType:  "deny",
			want:      "main/deny",
		},
		{
			name:      "warning",
			namespace: "main",
			ruleType:  "warn",
			want:      "main/warn",
		},
		{
			name:      "success",
			namespace: "main",
			ruleType:  "success",
			want:      "main/success",
		},
		{
			name:      "skipped",
			namespace: "main",
			ruleType:  "skip",
			want:      "main/skip",
		},
		{
			name:      "different namespace",
			namespace: "kubernetes",
			ruleType:  "deny",
			want:      "kubernetes/deny",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRuleID(tt.namespace, tt.ruleType); got != tt.want {
				t.Errorf("getRuleID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSARIF_Report(t *testing.T) {
	var buf bytes.Buffer
	s := NewSARIF(&buf)
	err := s.Report(nil, "test")
	if err == nil {
		t.Error("SARIF.Report() should return error")
	}
	const expectedErr = "report is not supported in SARIF output"
	if err.Error() != expectedErr {
		t.Errorf("expected '%v', got: '%v'", expectedErr, err)
	}
}

// compareJSON normalizes and compares two JSON strings.
// JSON strings are normalised to their canonical form without whitespace.
func compareJSON(t *testing.T, got, want string) {
	t.Helper()
	var gotJSON, wantJSON interface{}
	if err := json.Unmarshal([]byte(got), &gotJSON); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	gotBytes, _ := json.Marshal(gotJSON)
	wantBytes, _ := json.Marshal(wantJSON)

	if string(gotBytes) != string(wantBytes) {
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(want, got, false)
		t.Errorf("JSON mismatch:\n%s", dmp.DiffPrettyText(diffs))
	}
}
