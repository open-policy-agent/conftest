#!/usr/bin/env bats

@test "Not fail when testing a service with a warning" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Not fail when passed an explicit blank filename" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml ""
  [ "$status" -eq 0 ]
}

@test "Fail when testing a deployment with root containers" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing a service with warnings" {
  run ./conftest test --fail-on-warn -p examples/kubernetes/policy examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

@test "Pass when testing a blank namespace" {
  run ./conftest test --namespace notpresent -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  [ "$status" -eq 0 ]
}

@test "when testing a YAML document via stdin, default parser should be yaml if no input flag is passed" {
  run ./conftest test -p examples/kubernetes/policy - < examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Pass when testing a YAML document via stdin" {
  run ./conftest test -i yaml -p examples/kubernetes/policy - < examples/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Can parse stdin with input flag" {
  run bash -c "cat examples/ini/grafana.ini | ./conftest test -p examples/ini/policy --input ini -"
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Users should verify their e-mail address" ]]
  [[ "$output" != *"Basic auth should be enabled"* ]]
}

@test "Using -i/--input should force the chosen parser and fail the rego policy" {
  run ./conftest test -p examples/terraform/policy/gke.rego examples/terraform/gke.tf -i ini
  [ "$status" -eq 1 ]
}

@test "Fail due to picking up settings from configuration file" {
  cd examples/configfile
  run ../../conftest test deployment.yaml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Containers must not run as root" ]]
}

@test "Has trace flag" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Allow .rego files in the policy flag" {
  run ./conftest test -p examples/terraform/policy/base.rego examples/terraform/gke-show.json
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Terraform plan will change prohibited resources in the following namespaces: google_iam, google_container" ]]
}

@test "Can parse nested files with name overlap (first)" {
  run ./conftest test -p examples/nested/policy --namespace group1 examples/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can parse nested files with name overlap (second)" {
  run ./conftest test -p examples/nested/policy --namespace group2 examples/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can combine configs and reference by file" {
  run ./conftest test -p examples/terraform/policy/gke_combine.rego examples/terraform/gke.tf --combine
  [ "$status" -eq 0 ]
}

@test "Multi-file tests correctly fail when last file is fine" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}
