package main
import rego.v1

disallowed_ciphers := ["TLS_RSA_WITH_AES_256_GCM_SHA384"]

deny contains msg if {
	check_trusted_ips(input.entryPoints.http.tls.cipherSuites, disallowed_ciphers)
	msg := sprintf("Following ciphers are not allowed: %v", [disallowed_ciphers])
}

check_trusted_ips(ciphers, denylist) if {
	ciphers[_] = denylist[_]
}
