package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
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
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules":          []any{},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            0,
								"exitCodeDescription": "No policy violations found",
							},
						},
						"results": []any{},
					},
				},
			}),
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
							Metadata: map[string]any{
								"package": "test",
								"rule":    "rule1",
							},
						},
					},
				},
			},
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/deny",
										"shortDescription": map[string]any{
											"text": "Policy violation",
										},
										"properties": map[string]any{
											"package": "test",
											"rule":    "rule1",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            1,
								"exitCodeDescription": "Policy violations found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/deny",
								"ruleIndex": 0,
								"level":     "error",
								"message": map[string]any{
									"text": "test failure",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
							Metadata: map[string]any{
								"foo": "bar",
							},
						},
					},
				},
			},
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/warn",
										"shortDescription": map[string]any{
											"text": "Policy warning",
										},
										"properties": map[string]any{
											"foo": "bar",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            0,
								"exitCodeDescription": "Policy warnings found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/warn",
								"ruleIndex": 0,
								"level":     "warning",
								"message": map[string]any{
									"text": "test warning",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
							Metadata: map[string]any{
								"description": "test exception description",
							},
						},
					},
				},
			},
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/allow",
										"shortDescription": map[string]any{
											"text": "Policy exception",
										},
										"properties": map[string]any{
											"description": "test exception description",
										},
									},
									{
										"id": "main/success",
										"shortDescription": map[string]any{
											"text": "Policy was satisfied successfully",
										},
										"properties": map[string]any{
											"description": "Policy was satisfied successfully",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            0,
								"exitCodeDescription": "No policy violations found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/allow",
								"ruleIndex": 0,
								"level":     "note",
								"message": map[string]any{
									"text": "test exception",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
							{
								"ruleId":    "main/success",
								"ruleIndex": 1,
								"level":     "none",
								"message": map[string]any{
									"text": "Policy was satisfied successfully",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/skip",
										"shortDescription": map[string]any{
											"text": "Policy check was skipped",
										},
										"properties": map[string]any{
											"description": "Policy check was skipped",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            0,
								"exitCodeDescription": "No policy violations found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/skip",
								"ruleIndex": 0,
								"level":     "none",
								"message": map[string]any{
									"text": "Policy check was skipped",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
							Metadata: map[string]any{
								"package": "test",
								"rule":    "rule1",
							},
						},
						{
							Message: "test failure 2",
							Metadata: map[string]any{
								"package": "test",
								"rule":    "rule1",
							},
						},
					},
				},
			},
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/deny",
										"shortDescription": map[string]any{
											"text": "Policy violation",
										},
										"properties": map[string]any{
											"package": "test",
											"rule":    "rule1",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            1,
								"exitCodeDescription": "Policy violations found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/deny",
								"ruleIndex": 0,
								"level":     "error",
								"message": map[string]any{
									"text": "test failure 1",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test1.yaml",
											},
										},
									},
								},
							},
							{
								"ruleId":    "main/deny",
								"ruleIndex": 0,
								"level":     "error",
								"message": map[string]any{
									"text": "test failure 2",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test1.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
			wantJSON: mustJSON(t, map[string]any{
				"version": "2.1.0",
				"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
				"runs": []map[string]any{
					{
						"tool": map[string]any{
							"driver": map[string]any{
								"informationUri": "https://github.com/open-policy-agent/conftest",
								"name":           "conftest",
								"rules": []map[string]any{
									{
										"id": "main/success",
										"shortDescription": map[string]any{
											"text": "Policy was satisfied successfully",
										},
										"properties": map[string]any{
											"description": "Policy was satisfied successfully",
										},
									},
								},
							},
						},
						"invocations": []map[string]any{
							{
								"executionSuccessful": true,
								"exitCode":            0,
								"exitCodeDescription": "No policy violations found",
							},
						},
						"results": []map[string]any{
							{
								"ruleId":    "main/success",
								"ruleIndex": 0,
								"level":     "none",
								"message": map[string]any{
									"text": "Policy was satisfied successfully",
								},
								"locations": []map[string]any{
									{
										"physicalLocation": map[string]any{
											"artifactLocation": map[string]any{
												"uri": "test.yaml",
											},
										},
									},
								},
							},
						},
					},
				},
			}),
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
	var gotJSON, wantJSON any
	if err := json.Unmarshal([]byte(got), &gotJSON); err != nil {
		t.Fatalf("failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(want), &wantJSON); err != nil {
		t.Fatalf("failed to unmarshal expected JSON: %v", err)
	}

	if diff := cmp.Diff(wantJSON, gotJSON); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

// mustJSON converts a value to a JSON string, failing the test if marshaling fails
func mustJSON(t *testing.T, value map[string]any) string {
	t.Helper()
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
