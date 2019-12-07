package policy

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
			"should add latest tag",
			"user.azurecr.io/policies",
			"oci://user.azurecr.io/policies:latest",
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
