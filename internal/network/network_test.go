package network

import "testing"

func TestHostname(t *testing.T) {
	cases := []struct {
		ref      string
		hostname string
	}{
		{ref: "", hostname: ""},
		{ref: "hostname", hostname: "hostname"},
		{ref: "hostname:1234", hostname: "hostname"},
		{ref: "hostname/path", hostname: "hostname"},
		{ref: "hostname:1234/path", hostname: "hostname"},
		{ref: "hostname/path:1234", hostname: "hostname"},
		{ref: "oci://hostname", hostname: "hostname"},
		{ref: "oci://hostname:1234", hostname: "hostname"},
		{ref: "oci://hostname/path", hostname: "hostname"},
		{ref: "oci://hostname:1234/path", hostname: "hostname"},
		{ref: "oci://hostname/path:1234", hostname: "hostname"},
	}

	for _, c := range cases {
		t.Run(c.ref, func(t *testing.T) {
			got := Hostname(c.ref)
			if c.hostname != got {
				t.Errorf(`expecting Hostname("%s") == "%s", but it was "%s"`, c.ref, c.hostname, got)
			}
		})
	}
}

func TestIsLocalhost(t *testing.T) {
	cases := []struct {
		host  string
		local bool
	}{
		{host: "", local: false},
		{host: "google.com", local: false},
		{host: "1.1.1.1", local: false},
		{host: "2606:4700:4700::1111", local: false},
		{host: "localhost", local: true},
		{host: "127.0.0.1", local: true},
		{host: "127.0.0.2", local: true},
		{host: "::1", local: true},
		{host: "0:0:0:0:0:0:0:1", local: true},
	}

	for _, c := range cases {
		t.Run(c.host, func(t *testing.T) {
			got := IsLoopback(c.host)
			if c.local != got {
				t.Errorf(`expecting IsLocalhost("%s") == %v, but it was %v`, c.host, c.local, got)
			}
		})
	}
}
