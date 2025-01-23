package jsonnet

import (
	"reflect"
	"strings"
	"testing"
)

func TestJsonnetParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "basic jsonnet with self reference",
			input: `{
				person1: {
					name: "Alice",
					welcome: "Hello " + self.name + "!",
				},
				person2: self.person1 { name: "Bob" },
			}`,
			want: map[string]interface{}{
				"person1": map[string]interface{}{
					"name":    "Alice",
					"welcome": "Hello Alice!",
				},
				"person2": map[string]interface{}{
					"name":    "Bob",
					"welcome": "Hello Bob!",
				},
			},
			wantErr: false,
		},
		{
			name: "arithmetic operations",
			input: `{
				a: 1 + 2,
				b: 6 * 3,
				c: 10 - 5,
				d: 15 / 3,
			}`,
			want: map[string]interface{}{
				"a": float64(3),
				"b": float64(18),
				"c": float64(5),
				"d": float64(5),
			},
			wantErr: false,
		},
		{
			name:    "invalid jsonnet",
			input:   `{ invalid syntax `,
			want:    nil,
			wantErr: true,
			errMsg:  "evaluate anonymous snippet:",
		},
		{
			name: "array and nested objects",
			input: `{
				numbers: [1, 2, 3],
				nested: {
					a: { b: { c: "deep" } },
				},
			}`,
			want: map[string]interface{}{
				"numbers": []interface{}{float64(1), float64(2), float64(3)},
				"nested": map[string]interface{}{
					"a": map[string]interface{}{
						"b": map[string]interface{}{
							"c": "deep",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "stack overflow prevention",
			input: `
				local recurse(x) =
					if x == 0 then
						0
					else
						recurse(x-1) + 1;
				{ result: recurse(1000) }
			`,
			want:    nil,
			wantErr: true,
			errMsg:  "max stack frames exceeded",
		},
	}

	parser := &Parser{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			err := parser.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("Parser.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parser.Unmarshal() = %v, want %v", got, tt.want)
			}
		})
	}
}
