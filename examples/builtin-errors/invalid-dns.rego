package main
import rego.v1

deny_dnsresolution contains "testing DNS resolution" if {
  net.lookup_ip_addr("not-real-domainxyz")
}
