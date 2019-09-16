package main


deny[msg] {
  not input["resource.aws_elastic_beanstalk_environment.example"].application = "testing"
  msg = "Application is should be `testing`"
}

deny[msg] {
  not input["resource.aws_elastic_beanstalk_environment.example"].application = "staging_environment"
  msg = "Application environment is should be `staging_environment`"
}

deny[msg] {
  output := sprintf("%s", [input["resource.aws_elastic_beanstalk_environment.example"].setting])
  status = contains(output, "\"namespace\": \"aws:autoscaling:asg\"")
  not status
  msg = "The namespace should defined as `aws:autoscaling:asg`"
}

deny[msg] {
  output := input["resource.aws_elastic_beanstalk_environment.example"]["dynamic.setting"].for_each
  status = contains(output, "${data.consul_key_prefix.environment.var}")
  not status
  msg = "aws_elastic_beanstalk_environment dynamic.setting should contains a valid `for_each` `equals to data.consul_key_prefix.environment.var`"
}