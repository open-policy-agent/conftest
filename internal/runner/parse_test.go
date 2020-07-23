package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/open-policy-agent/conftest/parser"
)

func Test_Run_if_GetConfigurations_fails(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fileList := []string{"sut"}
	ctx := context.Background()

	manager := parser.NewMockCustomConfigManager(mockCtrl)
	manager.EXPECT().GetConfigurations(ctx, "", fileList).Return(nil, errors.New("dummy"))
	params := ParseParams{Input: "", Combine: false}
	runner := ParseRunner{Params: params, ConfigManager: manager}

	result, err := runner.Run(ctx, fileList)
	if result != "" {
		t.Fatalf("result should \"\" if an error occurs calling GetConfigurations() but got: %v", result)
	}

	if err == nil {
		t.Fatalf("err shouldn't be nil if an error occurs calling GetConfigurations() but got: %v", err)
	}
}

func Test_Run_if_GetConfigurations_succeed(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	fileList := []string{"sut"}
	ctx := context.Background()
	configurations := make(map[string]interface{})

	config := struct {
		Property string
	}{
		Property: "value",
	}
	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	manager := parser.NewMockCustomConfigManager(mockCtrl)
	manager.EXPECT().GetConfigurations(ctx, "", fileList).Return(configurations, nil)
	params := ParseParams{Input: "", Combine: false}
	runner := ParseRunner{Params: params, ConfigManager: manager}

	result, err := runner.Run(ctx, fileList)
	expected := `file.json
{
	"Property": "value"
}
`
	if result != expected {
		t.Fatalf("expected result: %v, but got: %v", expected, result)
	}

	if err != nil {
		t.Fatalf("err expected to be nil but got: %v", err)
	}
}

func TestParse_ByDefault_AddsIndentationAndNewline(t *testing.T) {
	params := ParseParams{Input: "", Combine: false}
	runner := ParseRunner{Params: params, ConfigManager: nil}
	configurations := make(map[string]interface{})

	config := struct {
		Property string
	}{
		Property: "value",
	}

	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	actual, err := runner.parseConfigurations(configurations)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `
{
	"Property": "value"
}
`

	if !strings.Contains(actual, expected) {
		t.Errorf("unexpected parsed config. expected %v actual %v", expected, actual)
	}

	if !strings.Contains(actual, expectedFileName) {
		t.Errorf("unexpected parsed filename. expected %v actual %v", expected, actual)
	}
}

func TestParse_MultiFileCombineFlag(t *testing.T) {
	params := ParseParams{Input: "", Combine: true}
	runner := ParseRunner{Params: params, ConfigManager: nil}
	configurations := make(map[string]interface{})

	config := struct {
		Sut string
	}{
		Sut: "test",
	}

	config2 := struct {
		Foo string
	}{
		Foo: "bar",
	}

	configurations["file1.json"] = config
	configurations["file2.json"] = config2

	actual, err := runner.parseConfigurations(configurations)
	if err != nil {
		t.Fatalf("parsing configs: %s", err)
	}

	expected := `{
	"file1.json": {
		"Sut": "test"
	},
	"file2.json": {
		"Foo": "bar"
	}
}
`

	if !strings.Contains(actual, expected) {
		t.Errorf("unexpected parsed config. expected %v actual %v", expected, actual)
	}
}
