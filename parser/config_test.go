package parser

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseConfigurations(t *testing.T) {
	testTable := []struct {
		name     string
		fileList []string
	}{
		{
			"Test single type",
			[]string{"../examples/kubernetes/service.yaml"},
		},
		{
			"Test two input with same type",
			[]string{"../examples/kubernetes/service.yaml", "../examples/kubernetes/deployment.yaml"},
		},
		{
			"Test different types",
			[]string{
				"../examples/kubernetes/service.yaml",
				"../examples/traefik/traefik.toml",
				"../examples/hcl1/gke.tf",
				"../examples/edn/sample_config.edn",
			},
		},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			_, err := ParseConfigurations(testUnit.fileList)
			if err != nil {
				t.Fatalf("error while getting configurations: %v", err)
			}
		})
	}
}

func TestGetFileType(t *testing.T) {
	testTable := []struct {
		name             string
		fileName         string
		expectedFileType string
	}{
		{"Test YAML file", "example/kubernetes/deployment.yaml", "yaml"},
		{"Test not YAML file", "example/traefik/traefik.toml", "toml"},
		{"Test default file type", "-", "yaml"},
	}

	for _, testUnit := range testTable {
		t.Run(testUnit.name, func(t *testing.T) {
			fileType := getFileType(testUnit.fileName)
			if fileType != testUnit.expectedFileType {
				t.Fatalf("got wrong filetype got:%s want:%s", fileType, testUnit.expectedFileType)
			}
		})
	}
}

func TestParseConfiguration(t *testing.T) {
	testTable := []struct {
		name           string
		path           string
		contents       []byte
		expectedResult map[string]interface{}
	}{
		{
			name:     "a single reader",
			path:     "sample.yml",
			contents: []byte("sample: true"),
			expectedResult: map[string]interface{}{
				"sample.yml": map[string]interface{}{
					"sample": true,
				},
			},
		},
		{
			name: "a single reader with multiple yaml subdocs",
			path: "sample.yml",
			contents: []byte(`---
sample: true
---
hello: true
---
nice: true`),
			expectedResult: map[string]interface{}{
				"sample.yml": []interface{}{
					map[string]interface{}{
						"sample": true,
					},
					map[string]interface{}{
						"hello": true,
					},
					map[string]interface{}{
						"nice": true,
					},
				},
			},
		},
	}

	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			var unmarshalledConfigs map[string]interface{}
			unmarshalledConfigs, err := parseConfiguration(test.path, test.contents, "")
			if err != nil {
				t.Errorf("errors unmarshalling: %v", err)
			}

			if unmarshalledConfigs == nil {
				t.Error("error seeing the actual value of object, received nil")
			}

			if !reflect.DeepEqual(test.expectedResult, unmarshalledConfigs) {
				t.Errorf("\nResult\n%v\n and type %T\n Expected\n%v\n and type %T\n", unmarshalledConfigs, unmarshalledConfigs, test.expectedResult, test.expectedResult)
			}
		})
	}
}

func TestFormatAll(t *testing.T) {
	configurations := make(map[string]interface{})
	config := struct {
		Property string
	}{
		Property: "value",
	}

	const expectedFileName = "file.json"
	configurations[expectedFileName] = config

	actual, err := FormatAll(configurations)
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

func TestFormat(t *testing.T) {
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

	actual, err := Format(configurations)
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
