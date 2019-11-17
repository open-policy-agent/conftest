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

@test "Fail due to picking up settings from configuration file" {
  cd examples/configfile
  run ../../conftest test deployment.yaml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Containers must not run as root" ]]
}

@test "Has version flag" {
  run ./conftest --version
  [ "$status" -eq 0 ]
}

@test "Test command has trace flag" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Verify command has trace flag" {
    run ./conftest verify --policy ./examples/kubernetes/policy --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Has help flag" {
  run ./conftest --help
  [ "$status" -eq 0 ]
}

@test "Allow .rego files in the policy flag" {
  run ./conftest test -p examples/terraform/policy/base.rego examples/terraform/gke-show.json
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Terraform plan will change prohibited resources in the following namespaces: google_iam, google_container" ]]
}

@test "Can parse tf files" {
  run ./conftest test -p examples/terraform/policy/gke.rego examples/terraform/gke.tf
  [ "$status" -eq 0 ]
}

@test "Can parse toml files" {
  run ./conftest test -p examples/traefik/policy examples/traefik/traefik.toml
  [ "$status" -eq 1 ]
}

@test "Can parse edn files" {
  run ./conftest test -p examples/edn/policy examples/edn/sample_config.edn
  [ "$status" -eq 1 ]
}

@test "Can parse nested files with name overlap (first)" {
  run ./conftest test -p examples/nested/policy --namespace group1 examples/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can parse nested files with name overlap (second)" {
  run ./conftest test -p examples/nested/policy --namespace group2 examples/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can parse cue files" {
  run ./conftest test -p examples/cue/policy examples/cue/deployment.cue
  [ "$status" -eq 1 ]
  [[ "$output" =~ "The image port should be 8080 in deployment.cue. you got : 8081" ]]
}

@test "Can parse ini files" {
  run ./conftest test -p examples/ini/policy examples/ini/grafana.ini
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Users should verify their e-mail address" ]]
}

@test "Can parse hcl2 files" {
  run ./conftest test -p examples/hcl2/policy examples/hcl2/terraform.tf -i hcl2
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Application environment is should be `staging_environment`" ]]
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
  
@test "Can combine configs and reference by file" {
  run ./conftest test -p examples/terraform/policy/gke_combine.rego examples/terraform/gke.tf --combine
  [ "$status" -eq 0 ]
}

@test "Can parse docker files" {
  run ./conftest test -p examples/docker/policy examples/docker/Dockerfile
  [ "$status" -eq 1 ]
  [[ "$output" =~ "blacklisted image found [\"openjdk:8-jdk-alpine\"]" ]]
}

@test "Can disable color" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/service.yaml --no-color
  [ "$status" -eq 0 ]
  [[ "$output" != *"[33m"* ]]
}

@test "Output results only once" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml
  count="${#lines[@]}"
  [ "$count" -eq 4 ]
}

@test "Can verify rego tests" {
  run ./conftest verify --policy ./examples/kubernetes/policy
  [ "$status" -eq 0 ]
  [[ "$output" =~ "test_services_not_denied" ]]
}

@test "Can parse inputs with 'conftest parse'" {
  run ./conftest parse examples/docker/Dockerfile
  [ "$status" -eq 0 ]
  [[ "$output" =~ "\"Cmd\": \"from\"" ]]
}

@test "Can change output format in test command" {
  run ./conftest test -p examples/kubernetes/policy/ -o tap examples/kubernetes/deployment.yaml
  [[ "$output" =~ "not ok" ]]
}

@test "Can change output format in verify command" {
  run ./conftest verify -p examples/kubernetes/policy/ -o tap
  [[ "$output" =~ "ok" ]]
}

@test "Multi-file tests correctly fail when last file is fine" {
  run ./conftest test -p examples/kubernetes/policy examples/kubernetes/deployment.yaml examples/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when unit test rego" {
  run ./conftest verify -p examples/traefik/policy
  [ "$status" -eq 1 ]
}
