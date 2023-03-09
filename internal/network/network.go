package network

import (
	"net"
	"strings"
)

func Hostname(ref string) string {
	ref = strings.TrimPrefix(ref, "oci://")

	colon := strings.Index(ref, ":")
	slash := strings.Index(ref, "/")

	cut := colon
	if colon == -1 || (colon > slash && slash != -1) {
		cut = slash
	}

	if cut < 0 {
		return ref
	}

	return ref[0:cut]
}

func IsLoopback(host string) bool {
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || host == "0:0:0:0:0:0:0:1" {
		// fast path
		return true
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return false
	}

	for _, ip := range ips {
		if ip.IsLoopback() {
			return true
		}
	}

	return false
}
