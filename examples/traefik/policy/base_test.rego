package main

test_ip_with_disallowed_ciphers {
    deny["IPs should not use disallowed ciphers"] with input as {"entryPoints": {"http": {"tls": {"cipherSuites": ["TLS_RSA_WITH_AES_256_GCM_SHA384"]}}}}
}
