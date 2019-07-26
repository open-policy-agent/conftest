package test

import (
	"bytes"
	"errors"
	"github.com/bmizerany/assert"
	"log"
	"strings"
	"testing"
)

func Test_stdOutputManager_put(t *testing.T) {
	type args struct {
		fileName string
		cr       checkResult
	}

	tests := []struct {
		msg    string
		args   args
		exp    []string
		expErr error
	}{
		{
			msg: "outputs filenames correctly",
			args: args{
				fileName: "foo.yaml",
				cr: checkResult{
				},
			},
			exp:    []string{"foo.yaml"},
		},
		{
			msg: "records failure and warnings",
			args: args{
				fileName: "foo.yaml",
				cr: checkResult{
					warnings: []error{errors.New("first warning")},
					failures: []error{errors.New("first failure")},
				},
			},
			exp:    []string{"foo.yaml", "\tfirst warning", "\tfirst failure"},
		},
		{
			msg: "skips filenames for stdin",
			args: args{
				fileName: "-",
				cr: checkResult{
					warnings: []error{errors.New("first warning")},
					failures: []error{errors.New("first failure")},
				},
			},
			exp:    []string{"\tfirst warning", "\tfirst failure"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			s := newStdOutputManager(log.New(buf, "", 0), false)

			err := s.put(tt.args.fileName, tt.args.cr)
			if err != nil {
				assert.Equal(t, tt.expErr, err)
			}

			// split on newlines but remove last one for easier comparisons
			res := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
			assert.Equal(t, tt.exp, res)
		})
	}
}
