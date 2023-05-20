package textproto

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	const (
		testProtoDef = `
syntax = "proto3";
package conftest.prototext;

enum AnOption {
	OPTION_UNDEFIED = 0;
	OPTION_GOOD = 1;
	OPTION_GREAT = 2;
}

message TestMessage {
	string name = 1;
	int32 number = 2;
	bool truthy = 3;
	AnOption option = 4;
}
`
		testTextProto = `
# proto-message: conftest.prototext.TestMessage

name: "foobarbaz"
number: 123123123
truthy: true
option: OPTION_GOOD
`
	)

	if err := load("test_file.proto", strings.NewReader(testProtoDef)); err != nil {
		t.Fatalf("Load test proto: %v", err)
	}

	testCases := []struct {
		desc    string
		input   string
		want    map[string]any
		wantErr bool
	}{
		{
			desc:  "valid case",
			input: testTextProto,
			want: map[string]any{
				"name":   "foobarbaz",
				"number": float64(123123123),
				"truthy": true,
				"option": "OPTION_GOOD",
			},
		},
		{
			desc:  "omitted fields are OK",
			input: "# proto-message: conftest.prototext.TestMessage\nnumber: 123123123",
			want: map[string]any{
				"number": float64(123123123),
			},
		},
		{
			desc:    "missing proto-message raises error",
			input:   "number: 123123123",
			wantErr: true,
			want:    make(map[string]any),
		},
		{
			desc:    "unknown proto-message raises error",
			input:   strings.ReplaceAll(testTextProto, "conftest", "another_package"),
			wantErr: true,
			want:    make(map[string]any),
		},
		{
			desc:    "known but invalid message raises an error",
			input:   strings.ReplaceAll(testTextProto, "conftest.prototext.TestMessage", "google.protobuf.FieldDescriptorProto"),
			wantErr: true,
			want:    make(map[string]any),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			parser := &Parser{}
			got := make(map[string]any)

			err := parser.Unmarshal([]byte(tc.input), &got)
			if err == nil && tc.wantErr || err != nil && !tc.wantErr {
				t.Errorf("unexpected error state, got %v", err)
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("unexpected diff (+got, -want):\n%s", diff)
			}
		})
	}
}
