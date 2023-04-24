package downloader

import "testing"

func TestOCIDetector_Detect(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"should detect azurecr",
			"user.azurecr.io/policies:tag",
			"oci://user.azurecr.io/policies:tag",
		},
		{
			"should detect gcr",
			"gcr.io/conftest/policies:tag",
			"oci://gcr.io/conftest/policies:tag",
		},
		{
			"should detect ecr",
			"123456789012.dkr.ecr.us-east-1.amazonaws.com/conftest/policies:tag",
			"oci://123456789012.dkr.ecr.us-east-1.amazonaws.com/conftest/policies:tag",
		},
		{
			"should detect gitlab",
			"registry.gitlab.com/conftest/policies:tag",
			"oci://registry.gitlab.com/conftest/policies:tag",
		},
		{
			"should add latest tag",
			"user.azurecr.io/policies",
			"oci://user.azurecr.io/policies:latest",
		},
		{
			"should detect 127.0.0.1:5000 as most likely being an OCI registry",
			"127.0.0.1:5000/policies:tag",
			"oci://127.0.0.1:5000/policies:tag",
		},
		{
			"should detect 127.0.0.1:5000 as most likely being an OCI registry and tag it properly if no tag is supplied",
			"127.0.0.1:5000/policies",
			"oci://127.0.0.1:5000/policies:latest",
		},
		{
			"should detect localhost:5000 as most likely being an OCI registry and tag it properly if no tag is supplied",
			"localhost:5000/policies",
			"oci://localhost:5000/policies:latest",
		},
		{
			"should detect Quay",
			"quay.io/conftest/policies:tag",
			"oci://quay.io/conftest/policies:tag",
		},
		{
			"should detect localhost:32123/policies:tag as most likely being an OCI registry",
			"localhost:32123/policies:tag",
			"oci://localhost:32123/policies:tag",
		},
		{
			"should detect 127.0.0.1:32123/policies:tag as most likely being an OCI registry",
			"127.0.0.1:32123/policies:tag",
			"oci://127.0.0.1:32123/policies:tag",
		},
		{
			"should detect ::1:32123/policies:tag as most likely being an OCI registry",
			"::1:32123/policies:tag",
			"oci://::1:32123/policies:tag",
		},
	}
	pwd := "/pwd"
	d := &OCIDetector{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, ok, err := d.Detect(tt.input, pwd)
			if err != nil {
				t.Fatalf("OCIDetector.Detect() error = %v", err)
			}
			if !ok {
				t.Fatal("OCIDetector.Detect() not ok, should have detected")
			}
			if out != tt.expected {
				t.Errorf("OCIDetector.Detect() output = %v, want %v", out, tt.expected)
			}
		})
	}
}
