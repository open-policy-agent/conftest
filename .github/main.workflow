workflow "Build" {
  on = "push"
  resolves = ["test", "release"]
}

action "test" {
  uses = "actions/docker/cli@master"
  args = "build --target acceptance ."
}

action "is-tag" {
  uses = "actions/bin/filter@master"
  args = "tag"
}

action "release" {
  uses = "docker://goreleaser/goreleaser"
  secrets = [
    "GITHUB_TOKEN",
  ]
  args = "release"
  needs = ["test", "is-tag"]
}
