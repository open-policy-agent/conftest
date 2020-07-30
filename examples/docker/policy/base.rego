package main

denylist = [
  "openjdk"
]

# Deny usage of the following tags
image_tag_list = [
    "latest",
]

# This alias list is used to skip check for latest tag in multistage dockerfile
image_alias_list = [
    "build",
    "base",
    "publish",
	"test",
]


deny[msg] {
  input[i].Cmd == "from"
  val := input[i].Value
  contains(val[i], denylist[_])

  msg = sprintf("unallowed image found %s", [val])
}

# Deny "latest" docker image
warn[msg] {
    input[i].Cmd == "from"
    val := split(input[i].Value[0], ":")
    contains(lower(val[1]), image_tag_list[_])
    msg = sprintf("Do not use latest tag with image: %s", [input[i].Value])
}

# Deny FROM image without tag, unless listed in the "image_alias_list" list
is_line_contains_alias_image(s) = true {
  contains(s, image_alias_list[_])
} else = false { true }

warn[msg] {
    input[i].Cmd == "from"
    val := split(input[i].Value[0], ":")
    not val[1]
    is_from_alias := is_line_contains_alias_image(val[0])
    is_from_alias != true
    msg = sprintf("Do not use image without any tag: %s", [input[i].Value])
}

# Deny  build without 'user' command
file_contains_user_cmd = true {
  input[_].Cmd = "user"
} else = false { true }

deny[msg] {
    input[i].Cmd == "from"
    is_user_cmd_present := file_contains_user_cmd
    is_user_cmd_present != true
    msg = "Dockerfile does not contain USER, run as a root suspected"
}
