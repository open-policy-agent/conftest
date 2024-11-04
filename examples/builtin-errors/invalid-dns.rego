package main

deny_dnsresolution["testing DNS resolution"] {
  net.lookup_ip_addr("not-real-domainxyz")
}
