package nginx

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNginxParser(t *testing.T) {
	t.Parallel()
	parser := &Parser{}
	sample := `
worker_processes 1;
http {
    server {
        listen 80;
        server_name example.com;
        location / {
            root /var/www/html;
        }
    }
}
`

	var got Config
	if err := parser.Unmarshal([]byte(sample), &got); err != nil {
		t.Fatalf("parser should not have thrown an error: %v", err)
	}

	want := Config{
		Directives: []Directive{
			{Name: "worker_processes", Parameters: []string{"1"}},
			{
				Name: "http",
				Block: &Block{
					Directives: []Directive{
						{
							Name: "server",
							Block: &Block{
								Directives: []Directive{
									{Name: "listen", Parameters: []string{"80"}},
									{Name: "server_name", Parameters: []string{"example.com"}},
									{
										Name:       "location",
										Parameters: []string{"/"},
										Block: &Block{
											Directives: []Directive{
												{Name: "root", Parameters: []string{"/var/www/html"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestConfigMarshalJSON(t *testing.T) {
	t.Parallel()
	config := Config{
		Directives: []Directive{
			{Name: "worker_processes", Parameters: []string{"1"}},
			{
				Name: "http",
				Block: &Block{
					Directives: []Directive{
						{
							Name: "server",
							Block: &Block{
								Directives: []Directive{
									{Name: "listen", Parameters: []string{"80"}},
									{Name: "server_name", Parameters: []string{"example.com"}},
								},
							},
						},
					},
				},
			},
		},
	}

	want := `{"directives":[{"name":"worker_processes","parameters":["1"]},{"name":"http","block":{"directives":[{"name":"server","block":{"directives":[{"name":"listen","parameters":["80"]},{"name":"server_name","parameters":["example.com"]}]}}]}}]}`

	got, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("unexpected error marshaling Config: %v", err)
	}

	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestNginxParserInvalidConfig(t *testing.T) {
	t.Parallel()
	parser := &Parser{}
	sample := `http {`

	var input any
	if err := parser.Unmarshal([]byte(sample), &input); err == nil {
		t.Error("expected error for invalid nginx config with unclosed block")
	}
}
