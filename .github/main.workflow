workflow "Quality" {
  on = "push"
  resolves = ["test"]
}

action "test" {
  uses = "actions/docker/cli@master"
  args = "build --target acceptance ."
}

