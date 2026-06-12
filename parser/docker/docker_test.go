package docker

import (
	"reflect"
	"testing"
)

func TestParser_Unmarshal(t *testing.T) {
	parser := Parser{}

	sample := `FROM foo
COPY . /
RUN echo hello`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	dockerFile := input.([]any)[0]
	commands := dockerFile.([]any)[0]

	expected := "from"
	actual := commands.(map[string]any)["Cmd"]

	if actual != expected {
		t.Errorf("first Docker command should be '%v', was '%v'", expected, actual)
	}
}

func TestParser_Unmarshal_Multistage(t *testing.T) {
	parser := Parser{}

	sample := `FROM golang:1.13-alpine as base
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
	
FROM base as builder
RUN go build -o conftest`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	if input == nil {
		t.Error("there should be information parsed but its nil")
	}

	dockerFile := input.([]any)[0]
	commands := dockerFile.([]any)

	cmd := commands[1]
	stage := cmd.(map[string]any)["Stage"].(float64)
	if stage != 0 {
		t.Errorf("expected command to be in stage 0, not stage: %v", stage)
	}

	cmd = commands[6]
	stage = cmd.(map[string]any)["Stage"].(float64)
	if stage != 1 {
		t.Errorf("expected command to be in stage 1, not stage: %v", stage)
	}
}

func TestParser_Unmarshal_EdgeCases(t *testing.T) {
	parser := Parser{}

	tests := []struct {
		name     string
		sample   string
		wantCmds []map[string]any
	}{
		{
			name: "multi-stage with AS aliases",
			sample: `FROM golang:1.13 AS builder
WORKDIR /app
FROM alpine AS runtime
COPY --from=builder /app/bin /usr/local/bin`,
			wantCmds: []map[string]any{
				{"Cmd": "from", "Stage": float64(0), "Value": []any{"golang:1.13", "AS", "builder"}},
				{"Cmd": "workdir", "Stage": float64(0), "Value": []any{"/app"}},
				{"Cmd": "from", "Stage": float64(1), "Value": []any{"alpine", "AS", "runtime"}},
				{"Cmd": "copy", "Stage": float64(1), "Flags": []any{"--from=builder"}, "Value": []any{"/app/bin", "/usr/local/bin"}},
			},
		},
		{
			name: "comments interleaved with instructions",
			sample: `# Build stage
FROM golang:1.13 AS builder
# download modules
RUN echo hello
# Final stage
FROM alpine
RUN echo world`,
			wantCmds: []map[string]any{
				{"Cmd": "comment", "Stage": float64(0), "Value": []any{"Build stage"}},
				{"Cmd": "from", "Stage": float64(0), "Value": []any{"golang:1.13", "AS", "builder"}},
				{"Cmd": "comment", "Stage": float64(0), "Value": []any{"download modules"}},
				{"Cmd": "run", "Stage": float64(0), "Value": []any{"echo hello"}},
				{"Cmd": "comment", "Stage": float64(1), "Value": []any{"Final stage"}},
				{"Cmd": "from", "Stage": float64(1), "Value": []any{"alpine"}},
				{"Cmd": "run", "Stage": float64(1), "Value": []any{"echo world"}},
			},
		},
		{
			name: "continuation lines with backslashes",
			sample: `FROM alpine
RUN echo hello \
    && echo world`,
			wantCmds: []map[string]any{
				{"Cmd": "from", "Stage": float64(0), "Value": []any{"alpine"}},
				{"Cmd": "run", "Stage": float64(0), "Value": []any{"echo hello     && echo world"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input any
			if err := parser.Unmarshal([]byte(tt.sample), &input); err != nil {
				t.Fatalf("parser should not have thrown an error: %v", err)
			}

			dockerFile := input.([]any)[0]
			commands := dockerFile.([]any)

			if len(commands) != len(tt.wantCmds) {
				t.Fatalf("expected %d commands, got %d", len(tt.wantCmds), len(commands))
			}

			for i, want := range tt.wantCmds {
				got := commands[i].(map[string]any)
				for key, wantVal := range want {
					gotVal, ok := got[key]
					if !ok {
						t.Errorf("command %d: missing key %q", i, key)
						continue
					}
					if !reflect.DeepEqual(gotVal, wantVal) {
						t.Errorf("command %d: key %q = %v, want %v", i, key, gotVal, wantVal)
					}
				}
			}
		})
	}
}
