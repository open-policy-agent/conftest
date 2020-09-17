#!/usr/bin/env bats

@test "Not fail when testing a service with a warning" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Not fail when passed an explicit blank filename" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml ""
  [ "$status" -eq 0 ]
}

@test "Fail when testing a deployment with root containers" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing a service with warnings" {
  run ./conftest test --fail-on-warn -p testdata/kubernetes/policy testdata/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when testing with no policies path" {
  run ./conftest test -p internal/ testdata/kubernetes/deployment.yaml
  [ "$status" -eq 1 ]
}

@test "Pass when testing a blank namespace" {
  run ./conftest test --namespace notpresent -p testdata/kubernetes/policy testdata/kubernetes/deployment.yaml
  [ "$status" -eq 0 ]
}

@test "when testing a YAML document via stdin, default parser should be yaml if no input flag is passed" {
  run ./conftest test -p testdata/kubernetes/policy - < testdata/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Pass when testing a YAML document via stdin" {
  run ./conftest test -i yaml -p testdata/kubernetes/policy - < testdata/kubernetes/service.yaml
  [ "$status" -eq 0 ]
}

@test "Fail due to picking up settings from configuration file" {
  cd testdata/configfile
  run ../../conftest test deployment.yaml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Containers must not run as root" ]]
}

@test "Has version flag" {
  run ./conftest --version
  [ "$status" -eq 0 ]
}

@test "Test command with multiple input type" {
  run ./conftest test testdata/traefik/traefik.toml testdata/kubernetes/service.yaml -p testdata/kubernetes/policy
  [ "$status" -eq 0 ]
  [[ "$output" =~ "Found service hello-kubernetes but services are not allowed" ]]
}

@test "Test command has trace flag" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Test command with json output and trace flag" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml -o json --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Test command with tap output and trace flag" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml -o tap --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Test command with all namespaces flag" {
  run ./conftest test -p testdata/docker/policy testdata/docker/Dockerfile --all-namespaces
  [ "$status" -eq 1 ]
  [[ "$output" =~ "unallowed image found [\"openjdk:8-jdk-alpine\"]" ]]
  [[ "$output" =~ "unallowed commands found [\"apk add --no-cache python3 python3-dev build-base && pip3 install awscli==1.18.1\"]" ]]
}

@test "Test command works with nested namespaces" {
  run ./conftest test --namespace main.gke -p testdata/hcl1/policy/ testdata/hcl1/gke.tf --no-color
  [ "$status" -eq 1 ]
  [ "${lines[1]}" = "1 test, 0 passed, 0 warnings, 1 failure, 0 exceptions" ]
}

@test "Verify command has trace flag" {
    run ./conftest verify --policy ./testdata/kubernetes/policy --trace
  [ "$status" -eq 0 ]
  [[ "$output" =~ "data.kubernetes.is_service" ]]
}

@test "Fail when verifying with no policies path" {
  run ./conftest verify -p internal/
  [ "$status" -eq 1 ]
}

@test "Has help flag" {
  run ./conftest --help
  [ "$status" -eq 0 ]
}

@test "Allow .rego files in the policy flag" {
  run ./conftest test -p testdata/hcl1/policy/base.rego testdata/hcl1/gke-show.json
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Terraform plan will change prohibited resources in the following namespaces: google_iam, google_container" ]]
}

@test "Can parse hcl1 files" {
  run ./conftest test -p testdata/hcl1/policy/gke.rego testdata/hcl1/gke.tf
  [ "$status" -eq 0 ]
}

@test "Can parse toml files" {
  run ./conftest test -p testdata/traefik/policy testdata/traefik/traefik.toml
  [ "$status" -eq 1 ]
}

@test "Can parse edn files" {
  run ./conftest test -p testdata/edn/policy testdata/edn/sample_config.edn
  [ "$status" -eq 1 ]
}

@test "Can parse xml files" {
  run ./conftest test -p testdata/xml/policy testdata/xml/pom.xml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "--- maven-plugin must have the version: 3.6.1" ]]
}

@test "Can parse hocon files" {
  run ./conftest test -p testdata/hocon/policy testdata/hocon/hocon.conf -i hocon
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Play http server port should be 9000" ]]
}

@test "Can parse vcl files" {
  run ./conftest test -p testdata/vcl/policy testdata/vcl/varnish.vcl
  [ "$status" -eq 1 ]
  [[ "$output" =~ "default backend port should be 8080" ]]
}

@test "Can parse jsonnet files" {
  run ./conftest test -p testdata/jsonnet/policy testdata/jsonnet/arith.jsonnet
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Concat array should be less than 3" ]]
}

@test "Can parse multi-type files" {
  run ./conftest test -p testdata/multitype/policy testdata/multitype/deployment.yaml testdata/multitype/grafana.ini
  [ "$status" -eq  1 ]
  [[ "$output" =~ "Port should be" ]]
}

@test "Can parse nested files with name overlap (first)" {
  run ./conftest test -p testdata/nested/policy --namespace group1 testdata/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can parse nested files with name overlap (second)" {
  run ./conftest test -p testdata/nested/policy --namespace group2 testdata/nested/data.json
  [ "$status" -eq 1 ]
}

@test "Can parse cue files" {
  run ./conftest test -p testdata/cue/policy testdata/cue/deployment.cue
  [ "$status" -eq 1 ]
  [[ "$output" =~ "The image port should be 8080 in deployment.cue. you got : 8081" ]]
}

@test "Can parse ini files" {
  run ./conftest test -p testdata/ini/policy testdata/ini/grafana.ini
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Users should verify their e-mail address" ]]
}

@test "Can parse hcl files" {
  run ./conftest test -p testdata/hcl/policy testdata/hcl/terraform.tf
  [ "$status" -eq 1 ]
  [[ "$output" =~ "ALB \`my-alb-listener\` is using HTTP rather than HTTP" ]]
}

@test "Can parse stdin with input flag" {
  run bash -c "cat testdata/ini/grafana.ini | ./conftest test -p testdata/ini/policy --input ini -"
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Users should verify their e-mail address" ]]
  [[ "$output" != *"Basic auth should be enabled"* ]]
}

@test "Using -i/--input should force the chosen parser and fail the rego policy" {
  run ./conftest test -p testdata/terraform/policy/gke.rego testdata/terraform/gke.tf -i ini
  [ "$status" -eq 1 ]
}

@test "Can combine configs and reference by file" {
  run ./conftest test -p testdata/hcl1/policy/gke_combine.rego testdata/hcl1/gke.tf --combine -i hcl1 --all-namespaces
  [ "$status" -eq 0 ]
}

@test "Can parse docker files" {
  run ./conftest test -p testdata/docker/policy testdata/docker/Dockerfile
  [ "$status" -eq 1 ]
  [[ "$output" =~ "unallowed image found [\"openjdk:8-jdk-alpine\"]" ]]
}

@test "Can disable color" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml --no-color
  [ "$status" -eq 0 ]
  [[ "$output" != *"[33m"* ]]
}

@test "Output results only once" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/deployment.yaml
  count="${#lines[@]}"
  [ "$count" -eq 5 ]
}

@test "Can verify rego tests" {
  run ./conftest verify --policy ./testdata/kubernetes/policy
  [ "$status" -eq 0 ]
  [[ "$output" =~ "4 tests, 4 passed" ]]
}

@test "Can parse inputs with 'conftest parse'" {
  run ./conftest parse testdata/docker/Dockerfile
  [ "$status" -eq 0 ]
  [[ "$output" =~ "\"Cmd\": \"from\"" ]]
}

@test "Can output tap format in test command" {
  run ./conftest test -p testdata/kubernetes/policy/ -o tap testdata/kubernetes/deployment.yaml
  [[ "$output" =~ "not ok" ]]
}

@test "Can output tap format in verify command" {
  run ./conftest verify -p testdata/kubernetes/policy/ -o tap
  [[ "$output" =~ "ok" ]]
}

@test "Can output table format in test command" {
  run ./conftest test -p testdata/kubernetes/policy/ -o table testdata/kubernetes/deployment.yaml
  [[ "$output" =~ "failure" ]]
}

@test "Can output table format in verify command" {
  run ./conftest verify -p testdata/kubernetes/policy/ -o table
  [[ "$output" =~ "success" ]]
}

@test "Multi-file tests correctly fail when last file is fine" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/deployment.yaml testdata/kubernetes/service.yaml
  [ "$status" -eq 1 ]
}

@test "Fail when unit test rego" {
  run ./conftest verify -p testdata/traefik/policy
  [ "$status" -eq 1 ]
}

@test "Can load data along with rego policies" {
  run ./conftest test -p testdata/data/policy -d testdata/data/exclusions testdata/data/service.yaml
  [ "$status" -eq 1 ]
  [[ "$output" =~ "Cannot expose one of the following ports" ]]
}

@test "Can load data in unit tests" {
  run ./conftest verify -p testdata/data/policy -d testdata/data/exclusions testdata/data/service.yaml
  [ "$status" -eq 0 ]
  [[ "$output" =~ "1 test, 1 passed, 0 warnings, 0 failures" ]]
}

@test "Can update policies in test command" {
  run ./conftest test --update https://raw.githubusercontent.com/open-policy-agent/conftest/master/examples/compose/policy/deny.rego testdata/compose/docker-compose.yml
  rm -rf policy/deny.rego
  [ "$status" -eq 1 ]
  [[ "$output" =~ "No images tagged latest" ]]
}

@test "Can download or symlink plugins" {
  run ./conftest plugin install testdata/plugins/kubectl/
  [ "$status" -eq 0 ]
  run ./conftest kubectl
  [ "$status" -eq 0 ]
}

@test "The number of tests run is accurate" {
  run ./conftest test -p testdata/kubernetes/policy testdata/kubernetes/service.yaml --no-color
  [ "$status" -eq 0 ]
  [ "${lines[1]}" = "5 tests, 4 passed, 1 warning, 0 failures, 0 exceptions" ]
}

@test "Exceptions reported correctly" {
  run ./conftest test -p testdata/exceptions/policy testdata/exceptions/deployments.yaml --no-color
  [ "$status" -eq 1 ]
  [ "${lines[2]}" = "2 tests, 0 passed, 0 warnings, 1 failure, 1 exception" ]
}

@test "Can have multiple namespace flags" {
  run ./conftest test -p testdata/nested/policy --namespace group1 --namespace group2 testdata/nested/data.json

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures" ]]
}

@test "Can have multiple policy flags" {
  run ./conftest test --policy testdata/multidir/org --policy testdata/multidir/team testdata/multidir/data.json

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 0 passed, 0 warnings, 2 failures" ]]
}

@test "Can combine yaml files" {
  run ./conftest test -p examples/combine/policy examples/combine/team.yaml examples/combine/user1.yaml examples/combine/user2.yaml --combine 

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 1 passed, 0 warnings, 1 failure" ]]
}

@test "Combining multi-document yaml file has same result" {
  run ./conftest test -p examples/combine/policy examples/combine/team.yaml examples/combine/users.yaml --combine 

  [ "$status" -eq 1 ]
  [[ "$output" =~ "2 tests, 1 passed, 0 warnings, 1 failure" ]]
}
