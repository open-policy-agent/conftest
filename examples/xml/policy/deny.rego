package main

plugin_list = input.project.build.plugins.plugin

deny[msg] {
  expected_version := "3.6.1"

  plugin_list[i].artifactId == "maven-compiler-plugin"
  not plugin_list[i].version = expected_version
  msg = sprintf("in %s \n--- maven-plugin must have the version: %s \n", [plugin_list[i], expected_version])
}

deny[msg] {
  plugin_list[i].artifactId == "activejdbc-instrumentation"
  not plugin_list[i].executions.execution.goals.goal = "instrument"
  msg = sprintf("in %s \n--- There must be defined 'instrument goal' for activejdbc-instrumentation \n", [plugin_list[i]])
}

deny[msg] {
  expected_version := "2.18.1"
  
  plugin_list[i].artifactId == "maven-surefire-plugin"
  not plugin_list[i].version = expected_version
  msg = sprintf("in %s \n--- Version must be %s for maven-surefire-plugin \n", [plugin_list[i],expected_version])
}